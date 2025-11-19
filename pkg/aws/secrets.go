package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsService provides Secrets Manager operations
type SecretsService struct {
	clientManager *ClientManager
}

// NewSecretsService creates a new Secrets Manager service
func NewSecretsService(clientManager *ClientManager) *SecretsService {
	return &SecretsService{
		clientManager: clientManager,
	}
}

// Secret represents a secret in Secrets Manager
type Secret struct {
	ARN              string
	Name             string
	Description      string
	CreatedDate      string
	LastAccessedDate string
	LastChangedDate  string
	Tags             map[string]string
}

// ListSecrets lists all secrets (without values)
func (s *SecretsService) ListSecrets(ctx context.Context, profileID string) ([]Secret, error) {
	client, err := s.clientManager.GetSecretsManagerClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	secrets := make([]Secret, 0, len(result.SecretList))
	for _, sec := range result.SecretList {
		secret := Secret{
			ARN:         aws.ToString(sec.ARN),
			Name:        aws.ToString(sec.Name),
			Description: aws.ToString(sec.Description),
		}

		if sec.CreatedDate != nil {
			secret.CreatedDate = sec.CreatedDate.String()
		}
		if sec.LastAccessedDate != nil {
			secret.LastAccessedDate = sec.LastAccessedDate.String()
		}
		if sec.LastChangedDate != nil {
			secret.LastChangedDate = sec.LastChangedDate.String()
		}

		// Add tags
		tags := make(map[string]string)
		for _, tag := range sec.Tags {
			tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}
		secret.Tags = tags

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// DescribeSecret gets metadata about a secret (without the value)
func (s *SecretsService) DescribeSecret(ctx context.Context, profileID string, secretName string) (map[string]interface{}, error) {
	client, err := s.clientManager.GetSecretsManagerClient(profileID)
	if err != nil {
		return nil, err
	}

	result, err := client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe secret: %w", err)
	}

	secretInfo := map[string]interface{}{
		"arn":         aws.ToString(result.ARN),
		"name":        aws.ToString(result.Name),
		"description": aws.ToString(result.Description),
	}

	if result.CreatedDate != nil {
		secretInfo["createdDate"] = result.CreatedDate.String()
	}
	if result.LastAccessedDate != nil {
		secretInfo["lastAccessedDate"] = result.LastAccessedDate.String()
	}
	if result.LastChangedDate != nil {
		secretInfo["lastChangedDate"] = result.LastChangedDate.String()
	}
	if result.LastRotatedDate != nil {
		secretInfo["lastRotatedDate"] = result.LastRotatedDate.String()
	}

	// Add tags
	tags := make(map[string]string)
	for _, tag := range result.Tags {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	secretInfo["tags"] = tags

	// Add version info
	if result.VersionIdsToStages != nil {
		secretInfo["versionIdsToStages"] = result.VersionIdsToStages
	}

	return secretInfo, nil
}

// GetSecretValue retrieves the actual secret value
// Note: This should be used with caution and only when explicitly requested
func (s *SecretsService) GetSecretValue(ctx context.Context, profileID string, secretName string) (string, error) {
	client, err := s.clientManager.GetSecretsManagerClient(profileID)
	if err != nil {
		return "", err
	}

	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	return aws.ToString(result.SecretString), nil
}

