package aws

import (
	"context"
	"fmt"
	"strconv"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/benjamingriff/secretsrc/pkg/models"
)

// ListParameters lists parameters from AWS SSM Parameter Store with pagination
// support. DescribeParameters returns metadata only — never values — so this is
// always safe to call regardless of KMS permissions. Tags are not returned by
// DescribeParameters and are intentionally omitted for parameters.
func (c *Client) ListParameters(ctx context.Context, maxResults int32, nextToken *string) ([]models.Entry, *string, error) {
	input := &ssm.DescribeParametersInput{
		MaxResults: &maxResults,
	}

	if nextToken != nil {
		input.NextToken = nextToken
	}

	result, err := c.ssm.DescribeParameters(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list parameters: %w", err)
	}

	entries := make([]models.Entry, 0, len(result.Parameters))
	for _, md := range result.Parameters {
		entries = append(entries, entryFromParameterMetadata(md))
	}

	return entries, result.NextToken, nil
}

// GetParameterValue retrieves and decrypts a parameter value. SecureString
// parameters are decrypted (WithDecryption), which requires kms:Decrypt on the
// parameter's key; a missing permission surfaces as an error to the caller.
func (c *Client) GetParameterValue(ctx context.Context, name string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: awssdk.Bool(true),
	}

	result, err := c.ssm.GetParameter(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get parameter value: %w", err)
	}

	if result.Parameter == nil || result.Parameter.Value == nil {
		return "", fmt.Errorf("parameter has no value")
	}

	return *result.Parameter.Value, nil
}

// entryFromParameterMetadata maps SSM parameter metadata to a models.Entry.
func entryFromParameterMetadata(md ssmtypes.ParameterMetadata) models.Entry {
	return models.Entry{
		Kind:             models.KindParameter,
		Name:             stringValue(md.Name),
		ARN:              stringValue(md.ARN),
		Description:      stringValue(md.Description),
		LastModifiedDate: md.LastModifiedDate,
		Type:             string(md.Type),
		Version:          strconv.FormatInt(md.Version, 10),
	}
}
