package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

const (
	// InfraspecCloudTokenEnvVar is the environment variable name for the InfraSpec Cloud token
	InfraspecCloudTokenEnvVar = "INFRASPEC_CLOUD_TOKEN"
	// UseInfraspecVirtualCloudEnvVar enables InfraSpec Cloud virtual cloud mode when set to a truthy value.
	UseInfraspecVirtualCloudEnvVar = "USE_INFRASPEC_VIRTUAL_CLOUD"
	// ConfigFileName is the name of the user config file
	ConfigFileName = "config.yaml"
)

// UserConfig represents user-specific configuration
type UserConfig struct {
	InfraspecCloudToken string `yaml:"infraspec_cloud_token,omitempty"`
}

// GetUserConfigPath returns the path to the user config file
func GetUserConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	infraspecDir := filepath.Join(configDir, "infraspec")
	return filepath.Join(infraspecDir, ConfigFileName), nil
}

// LoadUserConfig loads the user configuration from the config file
func LoadUserConfig() (*UserConfig, error) {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &UserConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read user config file: %w", err)
	}

	var userConfig UserConfig
	if err := yaml.Unmarshal(data, &userConfig); err != nil {
		return nil, fmt.Errorf("failed to parse user config file: %w", err)
	}

	return &userConfig, nil
}

// SaveUserConfig saves the user configuration to the config file
func SaveUserConfig(userConfig *UserConfig) error {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(userConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal user config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write user config file: %w", err)
	}

	return nil
}

// GetInfraspecCloudToken returns the InfraSpec Cloud token from config or environment variable
// Environment variable takes precedence over config file
func GetInfraspecCloudToken() (string, error) {
	// Check environment variable first
	if token := os.Getenv(InfraspecCloudTokenEnvVar); token != "" {
		Logging.Logger.Info("InfraSpec Cloud Token detected")
		return token, nil
	}

	// Fall back to config file
	userConfig, err := LoadUserConfig()
	if err != nil {
		return "", err
	}

	if userConfig.InfraspecCloudToken != "" {
		Logging.Logger.Info("InfraSpec Cloud Token detected")
	}

	return userConfig.InfraspecCloudToken, nil
}

// UseInfraspecVirtualCloud returns true if InfraSpec Virtual Cloud is enabled.
func UseInfraspecVirtualCloud() bool {
	if env := os.Getenv(UseInfraspecVirtualCloudEnvVar); env != "" {
		if enabled, err := strconv.ParseBool(env); err == nil {
			return enabled
		}
	}

	if currentConfig != nil {
		return currentConfig.VirtualCloud
	}

	return false
}
