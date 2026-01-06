package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"gopkg.in/ini.v1"
)

// MFAConfig represents MFA configuration for a profile
type MFAConfig struct {
	MFASerial     string
	Required      bool
	SourceProfile string // If set, MFA is for the source profile (role assumption)
}

// ProfileConfig represents profile configuration
type ProfileConfig struct {
	MFASerial     string
	SourceProfile string
	RoleARN       string
	Region        string
}

// GetProfileConfig gets configuration for a profile including source profile info
func GetProfileConfig(profile string) (*ProfileConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".aws", "config")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ProfileConfig{}, nil
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Look for the profile section
	sectionName := fmt.Sprintf("profile %s", profile)
	if profile == "default" {
		sectionName = "default"
	}

	section, err := cfg.GetSection(sectionName)
	if err != nil {
		// Profile not found in config
		return &ProfileConfig{}, nil
	}

	return &ProfileConfig{
		MFASerial:     section.Key("mfa_serial").String(),
		SourceProfile: section.Key("source_profile").String(),
		RoleARN:       section.Key("role_arn").String(),
		Region:        section.Key("region").String(),
	}, nil
}

// GetMFAConfig checks if a profile requires MFA, resolving source profiles
func GetMFAConfig(profile string) (*MFAConfig, error) {
	config, err := GetProfileConfig(profile)
	if err != nil {
		return nil, err
	}

	// If this profile has a source_profile, check the source for MFA
	if config.SourceProfile != "" {
		sourceConfig, err := GetProfileConfig(config.SourceProfile)
		if err != nil {
			return nil, err
		}
		if sourceConfig.MFASerial != "" {
			return &MFAConfig{
				MFASerial:     sourceConfig.MFASerial,
				Required:      true,
				SourceProfile: config.SourceProfile,
			}, nil
		}
	}

	// Check if this profile directly has MFA
	if config.MFASerial != "" {
		return &MFAConfig{
			MFASerial: config.MFASerial,
			Required:  true,
		}, nil
	}

	return &MFAConfig{Required: false}, nil
}

// GetSessionTokenWithMFA gets temporary credentials using MFA
func GetSessionTokenWithMFA(ctx context.Context, profile, region, mfaSerial, mfaToken string) (aws.Credentials, error) {
	// Load config without MFA to get base credentials
	var opts []func(*config.LoadOptions) error
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create STS client
	stsClient := sts.NewFromConfig(cfg)

	// Get session token with MFA
	duration := int32(43200) // 12 hours
	result, err := stsClient.GetSessionToken(ctx, &sts.GetSessionTokenInput{
		DurationSeconds: &duration,
		SerialNumber:    &mfaSerial,
		TokenCode:       &mfaToken,
	})
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to get session token: %w", err)
	}

	if result.Credentials == nil {
		return aws.Credentials{}, fmt.Errorf("no credentials returned from STS")
	}

	return aws.Credentials{
		AccessKeyID:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
		Source:          "STSGetSessionToken",
		CanExpire:       true,
		Expires:         *result.Credentials.Expiration,
	}, nil
}

// NewClientWithMFA creates a new AWS client using MFA credentials
func NewClientWithMFA(ctx context.Context, profile, region string, creds aws.Credentials) (*Client, error) {
	// Create config with the MFA credentials
	var opts []func(*config.LoadOptions) error

	// Use static credentials from the MFA session
	opts = append(opts, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: creds,
	}))

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config with MFA credentials: %w", err)
	}

	// Create Secrets Manager client
	sm := secretsmanager.NewFromConfig(cfg)

	return &Client{
		sm:      sm,
		profile: profile,
		region:  cfg.Region,
	}, nil
}

// NewClientWithMFAForRole creates a new AWS client for a role assumption profile using MFA credentials
func NewClientWithMFAForRole(ctx context.Context, profile, region string, sourceCreds aws.Credentials) (*Client, error) {
	// Get the profile configuration to find the role ARN
	profileConfig, err := GetProfileConfig(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile config: %w", err)
	}

	if profileConfig.RoleARN == "" {
		return nil, fmt.Errorf("profile %s does not have a role_arn configured", profile)
	}

	// Create a config with the source credentials
	var opts []func(*config.LoadOptions) error

	// Use the MFA session credentials from the source profile
	opts = append(opts, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: sourceCreds,
	}))

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	} else if profileConfig.Region != "" {
		opts = append(opts, config.WithRegion(profileConfig.Region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Use STS to assume the role
	stsClient := sts.NewFromConfig(cfg)

	// Assume the role
	assumeRoleOutput, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         &profileConfig.RoleARN,
		RoleSessionName: aws.String(fmt.Sprintf("secretsrc-%s", profile)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	if assumeRoleOutput.Credentials == nil {
		return nil, fmt.Errorf("no credentials returned from AssumeRole")
	}

	// Create config with the assumed role credentials
	roleCreds := aws.Credentials{
		AccessKeyID:     *assumeRoleOutput.Credentials.AccessKeyId,
		SecretAccessKey: *assumeRoleOutput.Credentials.SecretAccessKey,
		SessionToken:    *assumeRoleOutput.Credentials.SessionToken,
		Source:          "AssumeRole",
		CanExpire:       true,
		Expires:         *assumeRoleOutput.Credentials.Expiration,
	}

	roleConfig, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: roleCreds,
		}),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config with assumed role credentials: %w", err)
	}

	// Create Secrets Manager client with assumed role credentials
	sm := secretsmanager.NewFromConfig(roleConfig)

	return &Client{
		sm:      sm,
		profile: profile,
		region:  roleConfig.Region,
	}, nil
}
