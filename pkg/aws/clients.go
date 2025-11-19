package aws

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// ClientManager manages AWS service clients for multiple profiles
type ClientManager struct {
	config         *AWSConfig
	cloudwatchLogs map[string]*cloudwatchlogs.Client
	ecs            map[string]*ecs.Client
	rds            map[string]*rds.Client
	ec2            map[string]*ec2.Client
	lambda         map[string]*lambda.Client
	secretsManager map[string]*secretsmanager.Client
	cloudwatch     map[string]*cloudwatch.Client
	mu             sync.RWMutex
}

// NewClientManager creates a new AWS client manager
func NewClientManager(config *AWSConfig) *ClientManager {
	return &ClientManager{
		config:         config,
		cloudwatchLogs: make(map[string]*cloudwatchlogs.Client),
		ecs:            make(map[string]*ecs.Client),
		rds:            make(map[string]*rds.Client),
		ec2:            make(map[string]*ec2.Client),
		lambda:         make(map[string]*lambda.Client),
		secretsManager: make(map[string]*secretsmanager.Client),
		cloudwatch:     make(map[string]*cloudwatch.Client),
	}
}

// InitializeProfile initializes all AWS clients for a profile
func (cm *ClientManager) InitializeProfile(ctx context.Context, profileID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Load AWS config for profile
	cfg, err := cm.config.LoadProfile(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to load profile %s: %w", profileID, err)
	}

	// Initialize all service clients
	cm.cloudwatchLogs[profileID] = cloudwatchlogs.NewFromConfig(cfg)
	cm.ecs[profileID] = ecs.NewFromConfig(cfg)
	cm.rds[profileID] = rds.NewFromConfig(cfg)
	cm.ec2[profileID] = ec2.NewFromConfig(cfg)
	cm.lambda[profileID] = lambda.NewFromConfig(cfg)
	cm.secretsManager[profileID] = secretsmanager.NewFromConfig(cfg)
	cm.cloudwatch[profileID] = cloudwatch.NewFromConfig(cfg)

	return nil
}

// GetCloudWatchLogsClient returns the CloudWatch Logs client for a profile
func (cm *ClientManager) GetCloudWatchLogsClient(profileID string) (*cloudwatchlogs.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.cloudwatchLogs[profileID]
	if !exists {
		return nil, fmt.Errorf("CloudWatch Logs client not initialized for profile %s", profileID)
	}
	return client, nil
}

// GetECSClient returns the ECS client for a profile
func (cm *ClientManager) GetECSClient(profileID string) (*ecs.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.ecs[profileID]
	if !exists {
		return nil, fmt.Errorf("ECS client not initialized for profile %s", profileID)
	}
	return client, nil
}

// GetRDSClient returns the RDS client for a profile
func (cm *ClientManager) GetRDSClient(profileID string) (*rds.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.rds[profileID]
	if !exists {
		return nil, fmt.Errorf("RDS client not initialized for profile %s", profileID)
	}
	return client, nil
}

// GetEC2Client returns the EC2 client for a profile
func (cm *ClientManager) GetEC2Client(profileID string) (*ec2.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.ec2[profileID]
	if !exists {
		return nil, fmt.Errorf("EC2 client not initialized for profile %s", profileID)
	}
	return client, nil
}

// GetLambdaClient returns the Lambda client for a profile
func (cm *ClientManager) GetLambdaClient(profileID string) (*lambda.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.lambda[profileID]
	if !exists {
		return nil, fmt.Errorf("Lambda client not initialized for profile %s", profileID)
	}
	return client, nil
}

// GetSecretsManagerClient returns the Secrets Manager client for a profile
func (cm *ClientManager) GetSecretsManagerClient(profileID string) (*secretsmanager.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.secretsManager[profileID]
	if !exists {
		return nil, fmt.Errorf("Secrets Manager client not initialized for profile %s", profileID)
	}
	return client, nil
}

// GetCloudWatchClient returns the CloudWatch client for a profile
func (cm *ClientManager) GetCloudWatchClient(profileID string) (*cloudwatch.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.cloudwatch[profileID]
	if !exists {
		return nil, fmt.Errorf("CloudWatch client not initialized for profile %s", profileID)
	}
	return client, nil
}

// ListProfiles returns all initialized profile IDs
func (cm *ClientManager) ListProfiles() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	profiles := make([]string, 0, len(cm.cloudwatchLogs))
	for profileID := range cm.cloudwatchLogs {
		profiles = append(profiles, profileID)
	}
	return profiles
}

