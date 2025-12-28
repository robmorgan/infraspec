package emulator

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockActionProviderService implements both Service and ActionProvider interfaces
type mockActionProviderService struct {
	name    string
	actions []string
}

func (s *mockActionProviderService) ServiceName() string {
	return s.name
}

func (s *mockActionProviderService) HandleRequest(ctx context.Context, req *AWSRequest) (*AWSResponse, error) {
	return &AWSResponse{StatusCode: 200}, nil
}

func (s *mockActionProviderService) SupportedActions() []string {
	return s.actions
}

// mockBasicService implements only Service (not ActionProvider)
type mockBasicService struct {
	name string
}

func (s *mockBasicService) ServiceName() string {
	return s.name
}

func (s *mockBasicService) HandleRequest(ctx context.Context, req *AWSRequest) (*AWSResponse, error) {
	return &AWSResponse{StatusCode: 200}, nil
}

func TestRouter_ActionProviderRegistration(t *testing.T) {
	router := NewRouter()

	// Register a service that implements ActionProvider
	iamService := &mockActionProviderService{
		name:    "iam",
		actions: []string{"CreateRole", "GetRole", "DeleteRole"},
	}

	err := router.RegisterService(iamService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Verify actions are registered in the map
	if len(router.actionToSvc) != 3 {
		t.Errorf("Expected 3 actions registered, got %d", len(router.actionToSvc))
	}

	// Verify each action maps to the correct service
	for _, action := range iamService.actions {
		if svc, exists := router.actionToSvc[action]; !exists || svc != "iam" {
			t.Errorf("Action %s not properly registered, got service=%q, exists=%v", action, svc, exists)
		}
	}
}

func TestRouter_ActionProviderDuplicateActionRejected(t *testing.T) {
	router := NewRouter()

	// Register first service with CreateRole action
	service1 := &mockActionProviderService{
		name:    "iam",
		actions: []string{"CreateRole"},
	}
	if err := router.RegisterService(service1); err != nil {
		t.Fatalf("Failed to register first service: %v", err)
	}

	// Try to register second service with same action - should fail
	service2 := &mockActionProviderService{
		name:    "other-service",
		actions: []string{"CreateRole"},
	}
	err := router.RegisterService(service2)
	if err == nil {
		t.Error("Expected error when registering duplicate action, got nil")
	}
}

func TestRouter_BasicServiceWithoutActionProvider(t *testing.T) {
	router := NewRouter()

	// Register a service that doesn't implement ActionProvider
	basicService := &mockBasicService{name: "s3"}

	err := router.RegisterService(basicService)
	if err != nil {
		t.Fatalf("Failed to register basic service: %v", err)
	}

	// Verify no actions are registered for this service
	if len(router.actionToSvc) != 0 {
		t.Errorf("Expected 0 actions registered for basic service, got %d", len(router.actionToSvc))
	}

	// Verify service is still registered
	if _, exists := router.services["s3"]; !exists {
		t.Error("Expected s3 service to be registered")
	}
}

func TestRouter_ActionBasedRouting(t *testing.T) {
	router := NewRouter()

	// Register IAM service with some actions
	iamService := &mockActionProviderService{
		name:    "iam",
		actions: []string{"CreateRole", "GetRole"},
	}
	if err := router.RegisterService(iamService); err != nil {
		t.Fatalf("Failed to register IAM service: %v", err)
	}

	// Register EC2 service with some actions
	ec2Service := &mockActionProviderService{
		name:    "ec2",
		actions: []string{"RunInstances", "DescribeInstances"},
	}
	if err := router.RegisterService(ec2Service); err != nil {
		t.Fatalf("Failed to register EC2 service: %v", err)
	}

	// Test routing CreateRole to IAM
	body := "Action=CreateRole&RoleName=TestRole"
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	service, err := router.Route(req)
	if err != nil {
		t.Fatalf("Failed to route CreateRole request: %v", err)
	}
	if service.ServiceName() != "iam" {
		t.Errorf("Expected IAM service for CreateRole, got %s", service.ServiceName())
	}

	// Test routing RunInstances to EC2
	body = "Action=RunInstances&ImageId=ami-12345"
	req = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	service, err = router.Route(req)
	if err != nil {
		t.Fatalf("Failed to route RunInstances request: %v", err)
	}
	if service.ServiceName() != "ec2" {
		t.Errorf("Expected EC2 service for RunInstances, got %s", service.ServiceName())
	}
}

func TestRouter_SubdomainRoutingTakesPrecedence(t *testing.T) {
	router := NewRouter()

	// Register IAM service
	iamService := &mockActionProviderService{
		name:    "iam",
		actions: []string{"CreateRole"},
	}
	if err := router.RegisterService(iamService); err != nil {
		t.Fatalf("Failed to register IAM service: %v", err)
	}

	// Test that subdomain routing works (iam.infraspec.sh)
	body := "Action=CreateRole&RoleName=TestRole"
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	req.Host = "iam.infraspec.sh"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	service, err := router.Route(req)
	if err != nil {
		t.Fatalf("Failed to route request with subdomain: %v", err)
	}
	if service.ServiceName() != "iam" {
		t.Errorf("Expected IAM service for subdomain routing, got %s", service.ServiceName())
	}
}

func TestRouter_MultipleServicesRegistration(t *testing.T) {
	router := NewRouter()

	services := []*mockActionProviderService{
		{name: "iam", actions: []string{"CreateRole", "GetRole"}},
		{name: "ec2", actions: []string{"RunInstances"}},
		{name: "rds", actions: []string{"CreateDBInstance"}},
		{name: "sqs", actions: []string{"CreateQueue", "SendMessage"}},
		{name: "sts", actions: []string{"GetCallerIdentity"}},
	}

	for _, svc := range services {
		if err := router.RegisterService(svc); err != nil {
			t.Fatalf("Failed to register %s service: %v", svc.name, err)
		}
	}

	// Verify all services are registered
	if len(router.services) != 5 {
		t.Errorf("Expected 5 services, got %d", len(router.services))
	}

	// Verify total actions count
	expectedActions := 2 + 1 + 1 + 2 + 1 // 7 total
	if len(router.actionToSvc) != expectedActions {
		t.Errorf("Expected %d actions, got %d", expectedActions, len(router.actionToSvc))
	}
}

func TestRouter_DuplicateServiceRejected(t *testing.T) {
	router := NewRouter()

	service1 := &mockActionProviderService{name: "iam", actions: []string{"CreateRole"}}
	if err := router.RegisterService(service1); err != nil {
		t.Fatalf("Failed to register first service: %v", err)
	}

	// Try to register same service name again
	service2 := &mockActionProviderService{name: "iam", actions: []string{"GetRole"}}
	err := router.RegisterService(service2)
	if err == nil {
		t.Error("Expected error when registering duplicate service name, got nil")
	}
}

// Verify mock services implement the expected interfaces
var _ Service = (*mockActionProviderService)(nil)
var _ ActionProvider = (*mockActionProviderService)(nil)
var _ Service = (*mockBasicService)(nil)

// Ensure mockBasicService does NOT implement ActionProvider
func TestMockBasicServiceNotActionProvider(t *testing.T) {
	var svc Service = &mockBasicService{name: "test"}
	if _, ok := svc.(ActionProvider); ok {
		t.Error("mockBasicService should not implement ActionProvider")
	}
}

func createTestRequest(method, host, body string) *http.Request {
	req := httptest.NewRequest(method, "/", bytes.NewBufferString(body))
	if host != "" {
		req.Host = host
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req
}
