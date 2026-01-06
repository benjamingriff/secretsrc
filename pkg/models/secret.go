package models

import (
	"time"
)

// Secret represents an AWS Secrets Manager secret
type Secret struct {
	Name            string
	ARN             string
	Description     string
	LastChangedDate *time.Time
	Tags            map[string]string
}

// AppState represents the application configuration state
type AppState struct {
	CurrentProfile string
	CurrentRegion  string
}
