package config

import (
	"os"
	"strconv"

	"github.com/denisbrodbeck/machineid"
)

// TelemetryConfig holds telemetry-related configuration
type TelemetryConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	UserID  string `yaml:"user_id" json:"user_id"`
}

// LoadTelemetryConfig loads telemetry configuration from environment and config
func LoadTelemetryConfig() TelemetryConfig {
	cfg := TelemetryConfig{
		Enabled: true,
		UserID:  os.Getenv("INFRASPEC_USER_ID"),
	}

	// Allow disabling via environment variable
	if disabled, _ := strconv.ParseBool(os.Getenv("INFRASPEC_TELEMETRY_DISABLED")); disabled {
		cfg.Enabled = false
	}

	// Generate user ID if not provided
	if cfg.UserID == "" {
		deviceID, _ := machineid.ProtectedID("infraspec")
		cfg.UserID = deviceID
	}

	return cfg
}
