package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/benjamingriff/secretsrc/pkg/models"
)

// ListSecrets lists secrets from AWS Secrets Manager with pagination support
func (c *Client) ListSecrets(ctx context.Context, maxResults int32, nextToken *string) ([]models.Entry, *string, error) {
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

	entries := make([]models.Entry, 0, len(result.SecretList))
	for _, item := range result.SecretList {
		entry := models.Entry{
			Kind:             models.KindSecret,
			ARN:              stringValue(item.ARN),
			Name:             stringValue(item.Name),
			Description:      stringValue(item.Description),
			LastModifiedDate: item.LastChangedDate,
		}

		// Convert tags (returned inline by Secrets Manager)
		if len(item.Tags) > 0 {
			entry.Tags = make(map[string]string)
			for _, tag := range item.Tags {
				if tag.Key != nil && tag.Value != nil {
					entry.Tags[*tag.Key] = *tag.Value
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries, result.NextToken, nil
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
