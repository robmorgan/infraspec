package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"

	"github.com/robmorgan/infraspec/pkg/terratest/logger"
)

// initialize our custom Terratest logger early on so that we can use it in our Terraform steps.
var _ = logger.Default

// Config represents the main configuration structure
type Config struct {
	Version         string           `yaml:"version"`
	Provider        string           `yaml:"provider"`
	DefaultRegion   string           `yaml:"default_region"`
	StepDefinitions []StepDefinition `yaml:"step_definitions"`
	Functions       Functions        `yaml:"functions"`
	Cleanup         CleanupConfig    `yaml:"cleanup"`
	Retries         RetryConfig      `yaml:"retries"`
	//AWS             AWSConfig         `yaml:"aws"`
	//Logging         LoggingConfig     `yaml:"logging"`
	Verbose   bool
	Logger    *zap.SugaredLogger
	Telemetry TelemetryConfig
}

// StepDefinition defines a mapping between Gherkin steps and actions
type StepDefinition struct {
	Pattern         string            `yaml:"pattern"`
	TerratestAction string            `yaml:"terratest_action"`
	StoreAs         string            `yaml:"store_as,omitempty"`
	DataTable       bool              `yaml:"data_table,omitempty"`
	Parameters      map[string]string `yaml:"parameters,omitempty"`
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

// DefaultConfig returns a default configuration
func DefaultConfig() (*Config, error) {
	logger, err := initLogger()
	if err != nil {
		return nil, err
	}

	// try to detect the default AWS region using the AWS SDK
	region, err := defaultAwsRegion()
	if err != nil {
		region = "us-east-1"
		logger.Warnf("Failed to detect default region, using default value: %s", err)
	} else {
		logger.Debugf("Using AWS region: %s", region)
	}

	return &Config{
		Version:       "1.0",
		Provider:      "aws",
		DefaultRegion: region,
		Functions: Functions{
			RandomString: RandomStringConfig{
				Length:  6,
				Charset: "abcdefghijklmnopqrstuvwxyz0123456789",
			},
		},
		Telemetry: LoadTelemetryConfig(),
		Logger:    logger,
	}, nil
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

	if c.Provider == "aws" && c.DefaultRegion == "" {
		return fmt.Errorf("default_region is required for AWS provider")
	}

	return nil
}

// initLogger creates a logger with custom encoding config
func initLogger() (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeLevel = nil
	cfg.EncoderConfig.EncodeDuration = nil
	cfg.EncoderConfig.TimeKey = ""

	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeCaller = nil
	if os.Getenv("INFRASPEC_DEBUG") != "" {
		cfg.DisableStacktrace = false
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}

// defaultAwsRegion returns the default AWS region using the AWS SDK
func defaultAwsRegion() (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	return cfg.Region, nil
}
