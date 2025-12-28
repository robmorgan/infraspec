package embedded

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
	"github.com/robmorgan/infraspec/internal/emulator/metadata"
	"github.com/robmorgan/infraspec/internal/emulator/server"
	"github.com/robmorgan/infraspec/internal/emulator/services/applicationautoscaling"
	"github.com/robmorgan/infraspec/internal/emulator/services/dynamodb"
	"github.com/robmorgan/infraspec/internal/emulator/services/ec2"
	"github.com/robmorgan/infraspec/internal/emulator/services/iam"
	"github.com/robmorgan/infraspec/internal/emulator/services/lambda"
	"github.com/robmorgan/infraspec/internal/emulator/services/rds"
	"github.com/robmorgan/infraspec/internal/emulator/services/s3"
	"github.com/robmorgan/infraspec/internal/emulator/services/sqs"
	"github.com/robmorgan/infraspec/internal/emulator/services/sts"
)

// Emulator represents an embedded AWS emulator instance.
type Emulator struct {
	server   *server.Server
	state    *emulator.MemoryStateManager
	router   *emulator.Router
	listener net.Listener
	port     int
	mu       sync.Mutex
	running  bool
}

// instance is the singleton embedded emulator instance
var instance *Emulator

// New creates a new embedded emulator instance.
// The emulator will use a dynamically assigned port.
func New() *Emulator {
	return &Emulator{
		port: 0, // Dynamic port
	}
}

// GetInstance returns the current running emulator instance, or nil if not running.
func GetInstance() *Emulator {
	return instance
}

// Start initializes and starts the embedded emulator.
func (e *Emulator) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("emulator already running")
	}

	// Initialize core components
	e.state = emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	e.router = emulator.NewRouter()

	// Initialize resource relationship graph
	resourceManagerConfig := graph.ResourceManagerConfig{
		StrictValidation:      false,
		DefaultDeleteBehavior: graph.DeleteRestrict,
		DetectCycles:          true,
		UseAWSSchema:          true,
	}
	resourceManager := graph.NewResourceManager(e.state, resourceManagerConfig)

	// Register all service validations
	emulator.RegisterAllServices(validator)

	// Initialize EC2 metadata service
	if err := metadata.InitializeDefaults(e.state); err != nil {
		return fmt.Errorf("failed to initialize metadata service: %w", err)
	}

	// Register all services
	services := []emulator.Service{
		rds.NewRDSService(e.state, validator),
		s3.NewS3Service(e.state, validator),
		dynamodb.NewDynamoDBService(e.state, validator),
		applicationautoscaling.NewApplicationAutoScalingService(e.state, validator),
		sts.NewStsService(e.state, validator),
		ec2.NewEC2ServiceWithGraph(e.state, validator, resourceManager),
		iam.NewIAMServiceWithGraph(e.state, validator, resourceManager),
		sqs.NewSQSService(e.state, validator),
		lambda.NewLambdaService(e.state, validator),
	}

	for _, svc := range services {
		if err := e.router.RegisterService(svc); err != nil {
			return fmt.Errorf("failed to register service %s: %w", svc.ServiceName(), err)
		}
	}

	// Create listener with dynamic port
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", e.port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	e.listener = listener
	e.port = listener.Addr().(*net.TCPAddr).Port

	// Create server (no auth for embedded mode - nil keyStore)
	e.server = server.NewServer(e.port, e.router, nil, e.state)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := e.server.StartWithListener(e.listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Check for immediate failures
	select {
	case err := <-errChan:
		return fmt.Errorf("server failed to start: %w", err)
	case <-time.After(100 * time.Millisecond):
		// Server likely started successfully
	}

	e.running = true
	instance = e

	// Wait for server to be ready
	return e.waitForReady(ctx)
}

// Stop gracefully shuts down the emulator.
func (e *Emulator) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	if err := e.server.Stop(ctx); err != nil {
		return err
	}

	e.running = false
	instance = nil
	return nil
}

// ResetState clears all emulator state.
// This can be used between test scenarios for isolation.
func (e *Emulator) ResetState() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != nil {
		e.state.Clear()
		// Re-initialize metadata defaults
		metadata.InitializeDefaults(e.state)
	}
}

// Port returns the port the emulator is running on.
func (e *Emulator) Port() int {
	return e.port
}

// Endpoint returns the base endpoint URL.
func (e *Emulator) Endpoint() string {
	return fmt.Sprintf("http://127.0.0.1:%d", e.port)
}

// IsRunning returns true if the emulator is currently running.
func (e *Emulator) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *Emulator) waitForReady(ctx context.Context) error {
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/_health", e.port)
	client := &http.Client{Timeout: 1 * time.Second}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := client.Get(healthURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}
