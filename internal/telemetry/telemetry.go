package telemetry

import (
	"maps"
	"runtime"
	"time"

	"github.com/amplitude/analytics-go/amplitude"
	"go.uber.org/zap"
)

const (
	AmplitudeAPIKey = "9dc54881885bd60f8ccbb9cef2dfaa7a" //nolint:gosec // TODO - rotate and inject via env var
)

var (
	flushQueueSize = 10
	flushInterval  = 10 * time.Second
)

type Client struct {
	amplitude amplitude.Client
	enabled   bool
	userID    string
}

type Config struct {
	Enabled bool
	UserID  string
	Logger  *zap.SugaredLogger
}

// Event types for infraspec
const (
	EventTestRun       = "test_run"
	EventTestComplete  = "test_complete"
	EventTestFailed    = "test_failed"
	EventCLIStart      = "cli_start"
	EventConfigLoaded  = "config_loaded"
	EventFeatureLoaded = "feature_loaded"
)

// New creates a new telemetry client
func New(cfg Config) *Client {
	if !cfg.Enabled {
		return &Client{enabled: false}
	}

	config := amplitude.NewConfig(AmplitudeAPIKey)
	config.FlushQueueSize = flushQueueSize
	config.FlushInterval = flushInterval
	config.Logger = cfg.Logger
	client := amplitude.NewClient(config)

	return &Client{
		amplitude: client,
		enabled:   true,
		userID:    cfg.UserID,
	}
}

// Track sends an event to Amplitude
func (c *Client) Track(eventType string, properties map[string]interface{}) {
	if !c.enabled {
		return
	}

	// Add default properties
	props := c.getDefaultProperties()
	maps.Copy(props, properties)

	event := amplitude.Event{
		EventType:       eventType,
		UserID:          c.userID,
		EventProperties: props,
	}

	c.amplitude.Track(event)
}

// TrackCLIStart tracks when the CLI starts
func (c *Client) TrackCLIStart(args []string) {
	c.Track(EventCLIStart, map[string]interface{}{
		"args_count": len(args),
		"command":    "infraspec",
	})
}

// TrackTestRun tracks when a test run starts
func (c *Client) TrackTestRun(featureFile string) {
	c.Track(EventTestRun, map[string]interface{}{
		"feature_file": featureFile,
	})
}

// TrackTestComplete tracks when a test completes successfully
func (c *Client) TrackTestComplete(featureFile string, duration time.Duration, stepCount int) {
	c.Track(EventTestComplete, map[string]interface{}{
		"feature_file": featureFile,
		"duration_ms":  duration.Milliseconds(),
		"step_count":   stepCount,
		"success":      true,
	})
}

// TrackTestFailed tracks when a test fails
func (c *Client) TrackTestFailed(featureFile string, duration time.Duration, errorMsg string) {
	c.Track(EventTestFailed, map[string]interface{}{
		"feature_file": featureFile,
		"duration_ms":  duration.Milliseconds(),
		"error":        errorMsg,
		"success":      false,
	})
}

// TrackConfigLoaded tracks when config is loaded
func (c *Client) TrackConfigLoaded(configPath string) {
	c.Track(EventConfigLoaded, map[string]interface{}{
		"config_path": configPath,
	})
}

// Flush ensures all events are sent before shutdown
func (c *Client) Flush() {
	if !c.enabled {
		return
	}

	c.amplitude.Flush()
}

// getDefaultProperties returns common properties for all events
func (c *Client) getDefaultProperties() map[string]interface{} {
	return map[string]interface{}{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"tool":       "infraspec",
		"timestamp":  time.Now().Unix(),
	}
}
