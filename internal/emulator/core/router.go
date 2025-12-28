package emulator

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/auth"
)

type Router struct {
	services    map[string]Service
	actionToSvc map[string]string // maps action name to service name
}

func NewRouter() *Router {
	return &Router{
		services:    make(map[string]Service),
		actionToSvc: make(map[string]string),
	}
}

func (r *Router) RegisterService(service Service) error {
	name := service.ServiceName()
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}
	r.services[name] = service

	// If service implements ActionProvider, register its actions for routing
	if provider, ok := service.(ActionProvider); ok {
		for _, action := range provider.SupportedActions() {
			if existingSvc, exists := r.actionToSvc[action]; exists {
				return fmt.Errorf("action %s already registered by service %s", action, existingSvc)
			}
			r.actionToSvc[action] = name
		}
	}

	return nil
}

func (r *Router) Route(req *http.Request) (Service, error) {
	serviceName := r.extractServiceFromRequest(req)
	if serviceName == "" {
		// Debug logging for failed routing
		log.Printf("DEBUG: Failed to route request - Method: %s, Host: %s, Path: %s, ContentType: %s",
			req.Method, req.Host, req.URL.Path, req.Header.Get("Content-Type"))
		log.Printf("DEBUG: Headers: %v", req.Header)
		return nil, fmt.Errorf("unable to determine service from request")
	}

	service, exists := r.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return service, nil
}

func (r *Router) extractServiceFromRequest(req *http.Request) string {
	// FIRST: Check the request context for service name set by auth middleware
	// This is the most reliable method as it comes from the AWS SigV4 signature
	if serviceName, ok := req.Context().Value(auth.ServiceNameContextKey).(string); ok && serviceName != "" {
		return serviceName
	}

	// SECOND: Check for service-specific subdomains (e.g., dynamodb.infraspec.sh)
	// This takes priority over other detection methods
	host := req.Host
	// Use X-Forwarded-Host if present (for proxied requests like Railway)
	if forwardedHost := req.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	if host != "" {
		// Remove port from host if present
		hostWithoutPort := strings.Split(host, ":")[0]
		parts := strings.Split(hostWithoutPort, ".")

		// Check for service subdomain pattern: {service}.infraspec.sh
		// or {service}.localhost for local testing
		if len(parts) >= 2 {
			subdomain := parts[0]

			// Map service names to internal service identifiers
			serviceMap := map[string]string{
				"dynamodb":    "dynamodb_20120810",
				"autoscaling": "anyscalefrontendservice",
				"sts":         "sts",
				"rds":         "rds",
				"s3":          "s3",
				"ec2":         "ec2",
				"ssm":         "ssm",
				"sqs":         "sqs",
				"iam":         "iam",
				"lambda":      "lambda",
			}
			if internalName, ok := serviceMap[subdomain]; ok {
				return internalName
			}
		}
	}

	// THIRD: Check for X-Amz-Target header (JSON protocol services like DynamoDB)
	target := req.Header.Get("X-Amz-Target")
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			rawServiceName := strings.ToLower(parts[0])
			// Map AWS SDK service prefixes to internal service names
			targetServiceMap := map[string]string{
				"amazonsqs":               "sqs",
				"sqs":                     "sqs",
				"dynamodb_20120810":       "dynamodb_20120810",
				"dynamodb":                "dynamodb_20120810",
				"anyscalefrontendservice": "anyscalefrontendservice",
			}
			if internalName, ok := targetServiceMap[rawServiceName]; ok {
				return internalName
			}
			return rawServiceName
		}
	}

	// FOURTH: Extract service from Authorization header credential scope
	// Format: AWS4-HMAC-SHA256 Credential=ACCESS_KEY/DATE/REGION/SERVICE/aws4_request, ...
	// This is useful when auth is disabled but the client still sends SigV4 headers
	if authHeader := req.Header.Get("Authorization"); authHeader != "" && strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
		if credIdx := strings.Index(authHeader, "Credential="); credIdx != -1 {
			credPart := authHeader[credIdx+11:] // Skip "Credential="
			if commaIdx := strings.Index(credPart, ","); commaIdx != -1 {
				credPart = credPart[:commaIdx]
			}
			// credPart is now: ACCESS_KEY/DATE/REGION/SERVICE/aws4_request
			credComponents := strings.Split(credPart, "/")
			if len(credComponents) >= 4 {
				serviceName := credComponents[3]
				// Normalize service name to internal identifier
				serviceMap := map[string]string{
					"dynamodb":                "dynamodb_20120810",
					"application-autoscaling": "anyscalefrontendservice",
					"autoscaling":             "anyscalefrontendservice",
					"sts":                     "sts",
					"rds":                     "rds",
					"s3":                      "s3",
					"ec2":                     "ec2",
					"ssm":                     "ssm",
					"sqs":                     "sqs",
					"iam":                     "iam",
					"lambda":                  "lambda",
				}
				if internalName, ok := serviceMap[serviceName]; ok {
					return internalName
				}
				return serviceName
			}
		}
	}

	// FIFTH: Check for S3 service by looking for S3-specific indicators
	// S3 uses virtual-hosted-style bucket addressing:
	// - bucket-name.s3.infraspec.sh or bucket-name.s3.localhost (virtual-hosted)
	// - s3.infraspec.sh or s3.localhost (base S3 endpoint)
	if IsS3Request(host) {
		return "s3"
	}

	// Check for Query Protocol services by looking at form data for Action parameter
	if req.Method == "POST" && strings.Contains(req.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		body, err := io.ReadAll(req.Body)
		if err == nil {
			// Restore the body so it can be read again by the handler
			req.Body = io.NopCloser(bytes.NewBuffer(body))

			values, err := url.ParseQuery(string(body))
			if err == nil {
				if action := values.Get("Action"); action != "" {
					// Look up the service from the action-to-service map
					// Services register their actions via the ActionProvider interface
					if serviceName, exists := r.actionToSvc[action]; exists {
						return serviceName
					}
				}
			}
		}
	}

	// Fallback: extract service from host or path
	if host != "" {
		// Remove port from host if present
		hostWithoutPort := strings.Split(host, ":")[0]
		parts := strings.Split(hostWithoutPort, ".")
		if len(parts) > 0 && parts[0] != "localhost" && parts[0] != "127" {
			return parts[0]
		}
	}

	path := req.URL.Path
	if strings.HasPrefix(path, "/") {
		pathParts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(pathParts) > 0 && pathParts[0] != "" {
			return pathParts[0]
		}
	}

	return ""
}

func (r *Router) GetServices() []Service {
	services := make([]Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}
	return services
}
