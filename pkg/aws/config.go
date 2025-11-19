package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// ProfileConfig represents an AWS profile configuration
type ProfileConfig struct {
	ID              string   `json:"id"`
	AccessKeyID     string   `json:"access_key_id"`
	SecretAccessKey string   `json:"secret_access_key"`
	Region          string   `json:"region"`
	Project         string   `json:"project"`
	Environment     string   `json:"environment"`
	Description     string   `json:"description"`
	Tags            []string `json:"tags"`
}

// AWSConfig manages AWS SDK configuration
type AWSConfig struct {
	profiles map[string]*ProfileConfig
	configs  map[string]aws.Config
}

// NewAWSConfig creates a new AWS configuration manager
func NewAWSConfig() *AWSConfig {
	return &AWSConfig{
		profiles: make(map[string]*ProfileConfig),
		configs:  make(map[string]aws.Config),
	}
}

// AddProfile adds a profile configuration
func (ac *AWSConfig) AddProfile(profile *ProfileConfig) error {
	if profile.ID == "" {
		return fmt.Errorf("profile ID cannot be empty")
	}
	if profile.AccessKeyID == "" {
		return fmt.Errorf("access_key_id cannot be empty for profile %s", profile.ID)
	}
	if profile.SecretAccessKey == "" {
		return fmt.Errorf("secret_access_key cannot be empty for profile %s", profile.ID)
	}
	if profile.Region == "" {
		profile.Region = "us-east-1" // Default region
	}

	ac.profiles[profile.ID] = profile
	return nil
}

// LoadProfile loads AWS configuration for a specific profile
func (ac *AWSConfig) LoadProfile(ctx context.Context, profileID string) (aws.Config, error) {
	// Check if already loaded
	if cfg, exists := ac.configs[profileID]; exists {
		return cfg, nil
	}

	// Get profile configuration
	profile, exists := ac.profiles[profileID]
	if !exists {
		return aws.Config{}, fmt.Errorf("profile %s not found", profileID)
	}

	// Create credentials provider from access key and secret
	credsProvider := credentials.NewStaticCredentialsProvider(
		profile.AccessKeyID,
		profile.SecretAccessKey,
		"", // session token (empty for long-term credentials)
	)

	// Load AWS SDK config with credentials and region
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credsProvider),
		config.WithRegion(profile.Region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config for profile %s: %w", profileID, err)
	}

	// Cache the configuration
	ac.configs[profileID] = cfg
	return cfg, nil
}

// GetProfile returns a profile configuration by ID
func (ac *AWSConfig) GetProfile(profileID string) (*ProfileConfig, error) {
	profile, exists := ac.profiles[profileID]
	if !exists {
		return nil, fmt.Errorf("profile %s not found", profileID)
	}
	return profile, nil
}

// ListProfiles returns all configured profile IDs
func (ac *AWSConfig) ListProfiles() []string {
	profiles := make([]string, 0, len(ac.profiles))
	for id := range ac.profiles {
		profiles = append(profiles, id)
	}
	return profiles
}

// GetConfig returns the AWS SDK config for a profile
func (ac *AWSConfig) GetConfig(profileID string) (aws.Config, error) {
	cfg, exists := ac.configs[profileID]
	if !exists {
		return aws.Config{}, fmt.Errorf("AWS config not loaded for profile %s", profileID)
	}
	return cfg, nil
}

