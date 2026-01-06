package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"

	"github.com/FreePeak/infra-mcp-server/internal/logger"
	awspkg "github.com/FreePeak/infra-mcp-server/pkg/aws"
)

// AWSManager manages AWS service integrations
type AWSManager struct {
	config            *awspkg.AWSConfig
	clientManager     *awspkg.ClientManager
	cloudwatchService *awspkg.CloudWatchService
	ecsService        *awspkg.ECSService
	rdsService        *awspkg.RDSService
	ec2Service        *awspkg.EC2Service
	lambdaService     *awspkg.LambdaService
	secretsService    *awspkg.SecretsService
	metricsService    *awspkg.CloudWatchMetricsService
}

// NewAWSManager creates a new AWS manager
func NewAWSManager() *AWSManager {
	config := awspkg.NewAWSConfig()
	clientManager := awspkg.NewClientManager(config)

	return &AWSManager{
		config:            config,
		clientManager:     clientManager,
		cloudwatchService: awspkg.NewCloudWatchService(clientManager),
		ecsService:        awspkg.NewECSService(clientManager),
		rdsService:        awspkg.NewRDSService(clientManager),
		ec2Service:        awspkg.NewEC2Service(clientManager),
		lambdaService:     awspkg.NewLambdaService(clientManager),
		secretsService:    awspkg.NewSecretsService(clientManager),
		metricsService:    awspkg.NewCloudWatchMetricsService(clientManager),
	}
}

// InitializeProfiles initializes AWS profiles from configuration
func (am *AWSManager) InitializeProfiles(ctx context.Context, profiles []awspkg.ProfileConfig) error {
	for _, profile := range profiles {
		if err := am.config.AddProfile(&profile); err != nil {
			logger.Warn("Failed to add AWS profile %s: %v", profile.ID, err)
			continue
		}

		if err := am.clientManager.InitializeProfile(ctx, profile.ID); err != nil {
			logger.Warn("Failed to initialize AWS profile %s: %v", profile.ID, err)
			continue
		}

		logger.Info("Initialized AWS profile: %s (%s)", profile.ID, profile.Description)
	}

	return nil
}

// RegisterTools registers all AWS tools for all profiles
func (am *AWSManager) RegisterTools(ctx context.Context, mcpServer *server.MCPServer) error {
	profiles := am.config.ListProfiles()
	if len(profiles) == 0 {
		logger.Info("No AWS profiles configured, skipping AWS tool registration")
		return nil
	}

	logger.Info("Registering AWS tools for %d profile(s)", len(profiles))

	skippedCount := 0
	registeredCount := 0

	for _, profileID := range profiles {
		// Skip profiles that are pending (either by tag or TODO credentials)
		if am.isProfilePending(profileID) {
			logger.Info("Skipping pending AWS profile: %s (credentials not yet configured)", profileID)
			skippedCount++
			continue
		}

		if err := am.registerProfileTools(ctx, mcpServer, profileID); err != nil {
			logger.Warn("Failed to register AWS tools for profile %s: %v", profileID, err)
			continue
		}
		registeredCount++
	}

	logger.Info("AWS tool registration complete: %d registered, %d skipped (pending)", registeredCount, skippedCount)

	return nil
}

// isProfilePending checks if a profile should be skipped due to pending credentials
func (am *AWSManager) isProfilePending(profileID string) bool {
	profile, err := am.config.GetProfile(profileID)
	if err != nil {
		return false
	}

	// Check if profile has "pending" tag
	for _, tag := range profile.Tags {
		if tag == "pending" {
			return true
		}
	}

	// Check if credentials are TODO placeholders
	if profile.AccessKeyID == "TODO" || profile.SecretAccessKey == "TODO" {
		return true
	}

	return false
}

// registerProfileTools registers all tools for a specific profile
func (am *AWSManager) registerProfileTools(ctx context.Context, mcpServer *server.MCPServer, profileID string) error {
	profile, err := am.config.GetProfile(profileID)
	if err != nil {
		return err
	}

	logger.Info("Registering AWS tools for profile: %s", profileID)

	// Register CloudWatch Logs tools
	am.registerCloudWatchLogsTools(ctx, mcpServer, profileID, profile)

	// Register ECS tools
	am.registerECSTools(ctx, mcpServer, profileID, profile)

	// Register RDS tools
	am.registerRDSTools(ctx, mcpServer, profileID, profile)

	// Register EC2 tools
	am.registerEC2Tools(ctx, mcpServer, profileID, profile)

	// Register Lambda tools
	am.registerLambdaTools(ctx, mcpServer, profileID, profile)

	// Register Secrets Manager tools
	am.registerSecretsTools(ctx, mcpServer, profileID, profile)

	return nil
}

// registerCloudWatchLogsTools registers CloudWatch Logs tools
func (am *AWSManager) registerCloudWatchLogsTools(ctx context.Context, mcpServer *server.MCPServer, profileID string, profile *awspkg.ProfileConfig) {
	// List log groups
	toolName := fmt.Sprintf("aws_logs_list_%s", profileID)
	tool := tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("List CloudWatch log groups in %s", profile.Description)),
		tools.WithString("prefix", tools.Description("Optional prefix to filter log groups")),
		tools.WithNumber("limit", tools.Description("Maximum number of log groups (default: 50)")),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		prefix, _ := request.Parameters["prefix"].(string)
		limit := int32(50)
		if l, ok := request.Parameters["limit"].(float64); ok {
			limit = int32(l)
		}
		logGroups, err := am.cloudwatchService.ListLogGroups(ctx, profileID, prefix, limit)
		return FormatResponse(logGroups, err)
	})

	// Query logs
	toolName = fmt.Sprintf("aws_logs_query_%s", profileID)
	tool = tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("Query CloudWatch logs in %s. Defaults to last 24 hours if no time range specified.", profile.Description)),
		tools.WithString("log_group", tools.Description("Log group name"), tools.Required()),
		tools.WithString("filter_pattern", tools.Description("Optional filter pattern")),
		tools.WithNumber("start_time", tools.Description("Start time in milliseconds since epoch (defaults to 24 hours ago)")),
		tools.WithNumber("end_time", tools.Description("End time in milliseconds since epoch (defaults to now)")),
		tools.WithNumber("limit", tools.Description("Max events (default: 100)")),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		logGroup, _ := request.Parameters["log_group"].(string)
		filterPattern, _ := request.Parameters["filter_pattern"].(string)
		
		// Default to last 24 hours if no time range specified
		now := time.Now()
		startTime := now.Add(-24 * time.Hour).UnixMilli()
		endTime := now.UnixMilli()
		
		if st, ok := request.Parameters["start_time"].(float64); ok && st > 0 {
			startTime = int64(st)
		}
		if et, ok := request.Parameters["end_time"].(float64); ok && et > 0 {
			endTime = int64(et)
		}
		limit := int32(100)
		if l, ok := request.Parameters["limit"].(float64); ok {
			limit = int32(l)
		}
		events, err := am.cloudwatchService.QueryLogs(ctx, profileID, logGroup, filterPattern, startTime, endTime, limit)
		return FormatResponse(events, err)
	})

	logger.Info("Registered CloudWatch Logs tools for profile %s", profileID)
}

// registerECSTools registers ECS tools
func (am *AWSManager) registerECSTools(ctx context.Context, mcpServer *server.MCPServer, profileID string, profile *awspkg.ProfileConfig) {
	// List clusters
	toolName := fmt.Sprintf("aws_ecs_clusters_%s", profileID)
	tool := tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("List ECS clusters in %s", profile.Description)),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		clusters, err := am.ecsService.ListClusters(ctx, profileID)
		return FormatResponse(clusters, err)
	})

	// List services
	toolName = fmt.Sprintf("aws_ecs_services_%s", profileID)
	tool = tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("List ECS services in %s", profile.Description)),
		tools.WithString("cluster_name", tools.Description("Cluster name or ARN"), tools.Required()),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		clusterName, _ := request.Parameters["cluster_name"].(string)
		services, err := am.ecsService.ListServices(ctx, profileID, clusterName)
		return FormatResponse(services, err)
	})

	logger.Info("Registered ECS tools for profile %s", profileID)
}

// registerRDSTools registers RDS tools
func (am *AWSManager) registerRDSTools(ctx context.Context, mcpServer *server.MCPServer, profileID string, profile *awspkg.ProfileConfig) {
	// List DB instances
	toolName := fmt.Sprintf("aws_rds_list_%s", profileID)
	tool := tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("List RDS instances in %s", profile.Description)),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		instances, err := am.rdsService.ListDBInstances(ctx, profileID)
		return FormatResponse(instances, err)
	})

	// Describe DB instance
	toolName = fmt.Sprintf("aws_rds_describe_%s", profileID)
	tool = tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("Get RDS instance details in %s", profile.Description)),
		tools.WithString("identifier", tools.Description("DB instance identifier"), tools.Required()),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		identifier, _ := request.Parameters["identifier"].(string)
		instance, err := am.rdsService.DescribeDBInstance(ctx, profileID, identifier)
		return FormatResponse(instance, err)
	})

	logger.Info("Registered RDS tools for profile %s", profileID)
}

// registerEC2Tools registers EC2 tools
func (am *AWSManager) registerEC2Tools(ctx context.Context, mcpServer *server.MCPServer, profileID string, profile *awspkg.ProfileConfig) {
	toolName := fmt.Sprintf("aws_ec2_instances_%s", profileID)
	tool := tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("List EC2 instances in %s", profile.Description)),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		instances, err := am.ec2Service.ListInstances(ctx, profileID)
		return FormatResponse(instances, err)
	})
	logger.Info("Registered EC2 tools for profile %s", profileID)
}

// registerLambdaTools registers Lambda tools
func (am *AWSManager) registerLambdaTools(ctx context.Context, mcpServer *server.MCPServer, profileID string, profile *awspkg.ProfileConfig) {
	toolName := fmt.Sprintf("aws_lambda_list_%s", profileID)
	tool := tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("List Lambda functions in %s", profile.Description)),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		functions, err := am.lambdaService.ListFunctions(ctx, profileID)
		return FormatResponse(functions, err)
	})
	logger.Info("Registered Lambda tools for profile %s", profileID)
}

// registerSecretsTools registers Secrets Manager tools
func (am *AWSManager) registerSecretsTools(ctx context.Context, mcpServer *server.MCPServer, profileID string, profile *awspkg.ProfileConfig) {
	toolName := fmt.Sprintf("aws_secrets_list_%s", profileID)
	tool := tools.NewTool(
		toolName,
		tools.WithDescription(fmt.Sprintf("List Secrets Manager secrets in %s (metadata only)", profile.Description)),
	)
	mcpServer.AddTool(ctx, tool, func(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
		secrets, err := am.secretsService.ListSecrets(ctx, profileID)
		return FormatResponse(secrets, err)
	})
	logger.Info("Registered Secrets Manager tools for profile %s", profileID)
}
