package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/benjamingriff/secretsrc/pkg/models"
)

// ListSecrets lists secrets from AWS Secrets Manager with pagination support
func (c *Client) ListSecrets(ctx context.Context, maxResults int32, nextToken *string) ([]models.Secret, *string, error) {
	input := &secretsmanager.ListSecretsInput{
		MaxResults: &maxResults,
	}

	if nextToken != nil {
		input.NextToken = nextToken
	}

	result, err := c.sm.ListSecrets(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	secrets := make([]models.Secret, 0, len(result.SecretList))
	for _, entry := range result.SecretList {
		secret := models.Secret{
			ARN:             stringValue(entry.ARN),
			Name:            stringValue(entry.Name),
			Description:     stringValue(entry.Description),
			LastChangedDate: entry.LastChangedDate,
		}

		// Convert tags
		if len(entry.Tags) > 0 {
			secret.Tags = make(map[string]string)
			for _, tag := range entry.Tags {
				if tag.Key != nil && tag.Value != nil {
					secret.Tags[*tag.Key] = *tag.Value
				}
			}
		}

		secrets = append(secrets, secret)
	}

	return secrets, result.NextToken, nil
}

// GetSecretValue retrieves and decrypts a secret value
func (c *Client) GetSecretValue(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := c.sm.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	// Return the secret string (most secrets are stored as strings)
	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	// If the secret is binary, return a message
	if result.SecretBinary != nil {
		return "[Binary secret - not displayable as text]", nil
	}

	return "", fmt.Errorf("secret has no value")
}

// stringValue safely dereferences a string pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
