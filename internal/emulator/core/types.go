package emulator

import (
	"context"
	"net/http"
)

type Service interface {
	HandleRequest(ctx context.Context, req *AWSRequest) (*AWSResponse, error)
	ServiceName() string
}

// ActionExtractor is an optional interface that services can implement
// to extract the action from a request before the generic handler logs it.
// This is useful for REST-based services like S3 that derive actions from
// HTTP method and path rather than headers or query parameters.
type ActionExtractor interface {
	Service
	ExtractAction(req *AWSRequest) string
}

// ActionProvider is an optional interface that Query Protocol services can implement
// to register their supported actions for request routing. This eliminates the need
// for hardcoded action lists in the router.
type ActionProvider interface {
	Service
	// SupportedActions returns the list of AWS API actions this service handles.
	// These are used by the router to determine which service handles a given
	// Query Protocol request when subdomain routing is not available.
	SupportedActions() []string
}

type AWSRequest struct {
	Method     string
	Path       string
	Headers    map[string]string
	Body       []byte
	Action     string
	Parameters map[string]interface{}
}

type AWSResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

type StateManager interface {
	Get(key string, result interface{}) error
	Set(key string, value interface{}) error
	Delete(key string) error
	List(prefix string) ([]string, error)
	// Exists checks if a key exists in the state store without retrieving its value.
	Exists(key string) bool
	// Update atomically reads a value, applies an update function, and writes it back.
	// The updateFn receives a pointer to the current value and should modify it in place.
	// If the key doesn't exist, returns an error.
	Update(key string, result interface{}, updateFn func() error) error
}

type Validator interface {
	ValidateRequest(req *AWSRequest) error
	ValidateAction(action string, params map[string]interface{}) error
}

type RequestRouter interface {
	Route(req *http.Request) (Service, error)
	RegisterService(service Service) error
	GetServices() []Service
}
