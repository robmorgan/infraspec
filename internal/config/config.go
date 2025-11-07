package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	yaml "gopkg.in/yaml.v3"
)

var randomStringLength = 6

// Config represents the main configuration structure
type Config struct {
	Version         string           `yaml:"version"`
	Provider        string           `yaml:"provider"`
	StepDefinitions []StepDefinition `yaml:"step_definitions"`
	Functions       Functions        `yaml:"functions"`
	Cleanup         CleanupConfig    `yaml:"cleanup"`
	Retries         RetryConfig      `yaml:"retries"`
	Verbose         bool             `yaml:"verbose"` // Enable verbose mode
	Debug           bool             `yaml:"debug"`   // Enable debug mode
	Telemetry       TelemetryConfig  `yaml:"telemetry"`
	VirtualCloud    bool             `yaml:"virtual_cloud"`
}

// StepDefinition defines a mapping between Gherkin steps and actions
type StepDefinition struct {
	Pattern    string            `yaml:"pattern"`
	StoreAs    string            `yaml:"store_as,omitempty"`
	DataTable  bool              `yaml:"data_table,omitempty"`
	Parameters map[string]string `yaml:"parameters,omitempty"`
}

// Functions contains configuration for various utility functions
type Functions struct {
	RandomString RandomStringConfig `yaml:"random_string"`
}

// RandomStringConfig configures random string generation
type RandomStringConfig struct {
	Length  int    `yaml:"length"`
	Charset string `yaml:"charset"`
}

// CleanupConfig defines cleanup behavior
type CleanupConfig struct {
	Automatic bool        `yaml:"automatic"`
	Timeout   int         `yaml:"timeout"`
	Strategy  string      `yaml:"strategy"` // eager or deferred
	Retries   RetryConfig `yaml:"retries"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts     int           `yaml:"max_attempts"`
	Delay           time.Duration `yaml:"delay"`
	MaxDelay        time.Duration `yaml:"max_delay"`
	BackoffFactor   float64       `yaml:"backoff_factor"`
	RetryableErrors []string      `yaml:"retryable_errors"`
}

const defaultConfigPath = "infraspec.yaml"

var currentConfig *Config

// LoadConfig loads configuration from disk, applying default values and overrides from
// environment variables and the virtual cloud CLI flag. If the config file is missing,
// only the defaults are used.
func LoadConfig(path string, virtualCloudFlag bool) (*Config, error) {
	if path == "" {
		path = defaultConfigPath
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	applyDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	_ = v.BindEnv("virtual_cloud", UseInfraspecVirtualCloudEnvVar)

	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if virtualCloudFlag {
		v.Set("virtual_cloud", true)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	normalizeTelemetry(&cfg)

	currentConfig = &cfg
	return &cfg, nil
}

// Current returns the most recently loaded configuration.
func Current() *Config {
	return currentConfig
}

func applyDefaults(v *viper.Viper) {
	telemetryDefaults := LoadTelemetryConfig()

	// configure logging defaults once
	Logging.setLogLevel(zapcore.InfoLevel)
	if os.Getenv("INFRASPEC_DEBUG") != "" {
		Logging.setLogLevel(zapcore.DebugLevel)
	}

	v.SetDefault("version", "1.0")
	v.SetDefault("provider", "aws")
	v.SetDefault("functions.random_string.length", randomStringLength)
	v.SetDefault("functions.random_string.charset", "abcdefghijklmnopqrstuvwxyz0123456789")
	v.SetDefault("telemetry.enabled", telemetryDefaults.Enabled)
	v.SetDefault("telemetry.user_id", telemetryDefaults.UserID)
	v.SetDefault("virtual_cloud", false)
}

func normalizeTelemetry(cfg *Config) {
	defaults := LoadTelemetryConfig()

	if cfg.Telemetry.UserID == "" {
		cfg.Telemetry.UserID = defaults.UserID
	}

	if defaults.Enabled {
		cfg.Telemetry.Enabled = defaults.Enabled || cfg.Telemetry.Enabled
	}
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}

	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	return nil
}
