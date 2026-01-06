package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Client wraps the AWS SDK client for Secrets Manager
type Client struct {
	sm      *secretsmanager.Client
	profile string
	region  string
}

// NewClient creates a new AWS client with the specified profile and region
func NewClient(ctx context.Context, profile, region string) (*Client, error) {
	// Load AWS configuration with profile and region
	var opts []func(*config.LoadOptions) error

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Secrets Manager client
	sm := secretsmanager.NewFromConfig(cfg)

	return &Client{
		sm:      sm,
		profile: profile,
		region:  cfg.Region,
	}, nil
}

// GetProfile returns the current AWS profile
func (c *Client) GetProfile() string {
	return c.profile
}

// GetRegion returns the current AWS region
func (c *Client) GetRegion() string {
	return c.region
}

// GetSecretsManagerClient returns the underlying Secrets Manager client
func (c *Client) GetSecretsManagerClient() *secretsmanager.Client {
	return c.sm
}
