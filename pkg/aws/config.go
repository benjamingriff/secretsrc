package aws

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// GetDefaultProfile returns the default AWS profile to use
func GetDefaultProfile() string {
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		return profile
	}
	return "default"
}

// GetDefaultRegion returns the default AWS region to use
func GetDefaultRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	return "us-east-1"
}

// GetAvailableProfiles reads and returns all available AWS profiles from both ~/.aws/credentials and ~/.aws/config
func GetAvailableProfiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	credentialsPath := filepath.Join(homeDir, ".aws", "credentials")
	configPath := filepath.Join(homeDir, ".aws", "config")

	// Use a map to avoid duplicates
	profileMap := make(map[string]bool)

	// Read from credentials file
	if _, err := os.Stat(credentialsPath); err == nil {
		cfg, err := ini.Load(credentialsPath)
		if err == nil {
			for _, section := range cfg.Sections() {
				// Skip the default INI section
				if section.Name() != ini.DefaultSection {
					profileMap[section.Name()] = true
				}
			}
		}
	}

	// Read from config file
	if _, err := os.Stat(configPath); err == nil {
		cfg, err := ini.Load(configPath)
		if err == nil {
			for _, section := range cfg.Sections() {
				name := section.Name()
				// Skip the default INI section
				if name == ini.DefaultSection {
					continue
				}
				// In config file, profiles are named "profile <name>" except for default
				if name == "default" {
					profileMap["default"] = true
				} else if len(name) > 8 && name[:8] == "profile " {
					// Extract profile name from "profile <name>"
					profileName := name[8:]
					profileMap[profileName] = true
				}
			}
		}
	}

	// Convert map to slice
	profiles := make([]string, 0, len(profileMap))
	for profile := range profileMap {
		profiles = append(profiles, profile)
	}

	// If no profiles found, add default
	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}

	// Sort profiles for consistent ordering
	// Simple bubble sort since the list is small
	for i := 0; i < len(profiles)-1; i++ {
		for j := 0; j < len(profiles)-i-1; j++ {
			if profiles[j] > profiles[j+1] {
				profiles[j], profiles[j+1] = profiles[j+1], profiles[j]
			}
		}
	}

	return profiles, nil
}

// GetCommonRegions returns a list of commonly used AWS regions
func GetCommonRegions() []string {
	return []string{
		"us-east-1",      // US East (N. Virginia)
		"us-east-2",      // US East (Ohio)
		"us-west-1",      // US West (N. California)
		"us-west-2",      // US West (Oregon)
		"ca-central-1",   // Canada (Central)
		"eu-west-1",      // Europe (Ireland)
		"eu-west-2",      // Europe (London)
		"eu-west-3",      // Europe (Paris)
		"eu-central-1",   // Europe (Frankfurt)
		"eu-north-1",     // Europe (Stockholm)
		"ap-south-1",     // Asia Pacific (Mumbai)
		"ap-northeast-1", // Asia Pacific (Tokyo)
		"ap-northeast-2", // Asia Pacific (Seoul)
		"ap-northeast-3", // Asia Pacific (Osaka)
		"ap-southeast-1", // Asia Pacific (Singapore)
		"ap-southeast-2", // Asia Pacific (Sydney)
		"sa-east-1",      // South America (São Paulo)
	}
}
