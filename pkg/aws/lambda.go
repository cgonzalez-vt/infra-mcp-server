package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaService provides Lambda operations
type LambdaService struct {
	clientManager *ClientManager
}

// NewLambdaService creates a new Lambda service
func NewLambdaService(clientManager *ClientManager) *LambdaService {
	return &LambdaService{
		clientManager: clientManager,
	}
}

// Function represents a Lambda function
type Function struct {
	FunctionName string
	FunctionARN  string
	Runtime      string
	Handler      string
	CodeSize     int64
	Description  string
	Timeout      int32
	MemorySize   int32
	LastModified string
	Role         string
	Environment  map[string]string
}

// ListFunctions lists all Lambda functions
func (l *LambdaService) ListFunctions(ctx context.Context, profileID string) ([]Function, error) {
	client, err := l.clientManager.GetLambdaClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	functions := make([]Function, 0, len(result.Functions))
	for _, fn := range result.Functions {
		function := Function{
			FunctionName: aws.ToString(fn.FunctionName),
			FunctionARN:  aws.ToString(fn.FunctionArn),
			Runtime:      string(fn.Runtime),
			Handler:      aws.ToString(fn.Handler),
			CodeSize:     fn.CodeSize,
			Description:  aws.ToString(fn.Description),
			Timeout:      aws.ToInt32(fn.Timeout),
			MemorySize:   aws.ToInt32(fn.MemorySize),
			LastModified: aws.ToString(fn.LastModified),
			Role:         aws.ToString(fn.Role),
		}

		// Add environment variables
		if fn.Environment != nil && fn.Environment.Variables != nil {
			function.Environment = fn.Environment.Variables
		}

		functions = append(functions, function)
	}

	return functions, nil
}

// GetFunction gets detailed information about a Lambda function
func (l *LambdaService) GetFunction(ctx context.Context, profileID string, functionName string) (*Function, error) {
	client, err := l.clientManager.GetLambdaClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	fn := result.Configuration
	function := &Function{
		FunctionName: aws.ToString(fn.FunctionName),
		FunctionARN:  aws.ToString(fn.FunctionArn),
		Runtime:      string(fn.Runtime),
		Handler:      aws.ToString(fn.Handler),
		CodeSize:     fn.CodeSize,
		Description:  aws.ToString(fn.Description),
		Timeout:      aws.ToInt32(fn.Timeout),
		MemorySize:   aws.ToInt32(fn.MemorySize),
		LastModified: aws.ToString(fn.LastModified),
		Role:         aws.ToString(fn.Role),
	}

	if fn.Environment != nil && fn.Environment.Variables != nil {
		function.Environment = fn.Environment.Variables
	}

	return function, nil
}

// GetFunctionConfiguration gets the configuration of a Lambda function
func (l *LambdaService) GetFunctionConfiguration(ctx context.Context, profileID string, functionName string) (map[string]interface{}, error) {
	client, err := l.clientManager.GetLambdaClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.GetFunctionConfiguration(ctx, &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get function configuration: %w", err)
	}

	config := map[string]interface{}{
		"functionName": aws.ToString(result.FunctionName),
		"functionArn":  aws.ToString(result.FunctionArn),
		"runtime":      string(result.Runtime),
		"handler":      aws.ToString(result.Handler),
		"codeSize":     result.CodeSize,
		"description":  aws.ToString(result.Description),
		"timeout":      aws.ToInt32(result.Timeout),
		"memorySize":   aws.ToInt32(result.MemorySize),
		"lastModified": aws.ToString(result.LastModified),
		"role":         aws.ToString(result.Role),
		"state":        string(result.State),
	}

	if result.Environment != nil && result.Environment.Variables != nil {
		config["environment"] = result.Environment.Variables
	}

	return config, nil
}
