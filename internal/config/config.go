package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

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
	Logger *zap.SugaredLogger
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

	return &Config{
		Version:       "1.0",
		Provider:      "aws",
		DefaultRegion: "us-east-1",
		Functions: Functions{
			RandomString: RandomStringConfig{
				Length:  6,
				Charset: "abcdefghijklmnopqrstuvwxyz0123456789",
			},
		},
		Logger: logger,
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
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
