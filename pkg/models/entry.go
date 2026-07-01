package models

import (
	"time"
)

// Kind identifies which AWS data source an Entry came from.
type Kind int

const (
	// KindSecret is an AWS Secrets Manager secret.
	KindSecret Kind = iota
	// KindParameter is an AWS SSM Parameter Store parameter.
	KindParameter
)

// Entry represents a listable item from AWS — either a Secrets Manager secret
// or an SSM parameter. Fields common to both sources are always populated;
// Type and Version are only set for SSM parameters.
type Entry struct {
	Kind             Kind
	Name             string
	ARN              string
	Description      string
	LastModifiedDate *time.Time
	Tags             map[string]string

	// Type and Version are SSM-only and remain empty for secrets.
	Type    string // String | StringList | SecureString
	Version string
}

// AppState represents the application configuration state
type AppState struct {
	CurrentProfile string
	CurrentRegion  string
}
