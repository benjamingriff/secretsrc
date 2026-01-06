package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	LastProfile string `json:"last_profile"`
	LastRegion  string `json:"last_region"`
}

// CachedCredentials represents cached AWS credentials
type CachedCredentials struct {
	AccessKeyID     string    `json:"access_key_id"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// CredentialsCache stores cached credentials for multiple profiles
type CredentialsCache struct {
	Profiles map[string]CachedCredentials `json:"profiles"`
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".aws", "secretsrc")
	configFile := filepath.Join(configDir, "config.json")

	return configFile, nil
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configFile, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	configFile, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getCredentialsCachePath returns the path to the credentials cache file
func getCredentialsCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".aws", "secretsrc")
	cacheFile := filepath.Join(configDir, "cache.json")

	return cacheFile, nil
}

// LoadCredentialsCache loads cached credentials from disk
func LoadCredentialsCache() (*CredentialsCache, error) {
	cacheFile, err := getCredentialsCachePath()
	if err != nil {
		return nil, err
	}

	// If cache file doesn't exist, return empty cache
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return &CredentialsCache{
			Profiles: make(map[string]CachedCredentials),
		}, nil
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials cache: %w", err)
	}

	var cache CredentialsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse credentials cache: %w", err)
	}

	if cache.Profiles == nil {
		cache.Profiles = make(map[string]CachedCredentials)
	}

	return &cache, nil
}

// SaveCredentialsCache saves cached credentials to disk
func SaveCredentialsCache(cache *CredentialsCache) error {
	cacheFile, err := getCredentialsCachePath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials cache: %w", err)
	}

	// Write with restricted permissions (0600) for security
	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials cache: %w", err)
	}

	return nil
}

// GetCachedCredentials retrieves cached credentials for a profile if they exist and are still valid
func GetCachedCredentials(profile string) (*CachedCredentials, bool) {
	cache, err := LoadCredentialsCache()
	if err != nil {
		return nil, false
	}

	creds, exists := cache.Profiles[profile]
	if !exists {
		return nil, false
	}

	// Check if credentials have expired
	if time.Now().After(creds.ExpiresAt) {
		// Credentials expired, remove them
		delete(cache.Profiles, profile)
		_ = SaveCredentialsCache(cache) // Ignore errors
		return nil, false
	}

	return &creds, true
}

// SaveCachedCredentials saves credentials for a profile
func SaveCachedCredentials(profile string, creds CachedCredentials) error {
	cache, err := LoadCredentialsCache()
	if err != nil {
		cache = &CredentialsCache{
			Profiles: make(map[string]CachedCredentials),
		}
	}

	cache.Profiles[profile] = creds
	return SaveCredentialsCache(cache)
}
