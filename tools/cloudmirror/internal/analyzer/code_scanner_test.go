package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

func TestCodeScanner_ScanService_NotImplemented(t *testing.T) {
	// Create a temporary directory without any service implementation
	tmpDir := t.TempDir()
	scanner := NewCodeScanner(tmpDir)

	impl, err := scanner.ScanService("nonexistent")
	if err != nil {
		t.Errorf("Expected no error for non-existent service, got: %v", err)
	}
	if impl != nil {
		t.Errorf("Expected nil implementation for non-existent service, got: %+v", impl)
	}
}

func TestCodeScanner_ScanService_EmptyService(t *testing.T) {
	// Create a temporary directory with an empty service directory
	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "testservice")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatalf("Failed to create service directory: %v", err)
	}

	// Create an empty Go file
	goFile := filepath.Join(serviceDir, "service.go")
	if err := os.WriteFile(goFile, []byte("package testservice\n"), 0o644); err != nil {
		t.Fatalf("Failed to create Go file: %v", err)
	}

	scanner := NewCodeScanner(tmpDir)
	impl, err := scanner.ScanService("testservice")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if impl == nil {
		t.Fatal("Expected non-nil implementation")
	}
	if len(impl.Operations) != 0 {
		t.Errorf("Expected 0 operations, got %d", len(impl.Operations))
	}
}

func TestCodeScanner_ScanService_WithHandleRequest(t *testing.T) {
	// Create a temporary directory with a service that has HandleRequest
	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "testservice")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatalf("Failed to create service directory: %v", err)
	}

	// Create a Go file with HandleRequest method and action switch
	goContent := `package testservice

type TestService struct {}

func (s *TestService) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	action := req.Action
	switch action {
	case "CreateResource":
		return s.handleCreate(ctx, req)
	case "DeleteResource":
		return s.handleDelete(ctx, req)
	case "ListResources":
		return s.handleList(ctx, req)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (s *TestService) handleCreate(ctx context.Context, req *Request) (*Response, error) {
	return nil, nil
}

func (s *TestService) handleDelete(ctx context.Context, req *Request) (*Response, error) {
	return nil, nil
}

func (s *TestService) handleList(ctx context.Context, req *Request) (*Response, error) {
	return nil, nil
}
`
	goFile := filepath.Join(serviceDir, "service.go")
	if err := os.WriteFile(goFile, []byte(goContent), 0o644); err != nil {
		t.Fatalf("Failed to create Go file: %v", err)
	}

	scanner := NewCodeScanner(tmpDir)
	impl, err := scanner.ScanService("testservice")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if impl == nil {
		t.Fatal("Expected non-nil implementation")
	}

	expectedOps := []string{"CreateResource", "DeleteResource", "ListResources"}
	if len(impl.Operations) != len(expectedOps) {
		t.Errorf("Expected %d operations, got %d", len(expectedOps), len(impl.Operations))
	}

	for _, op := range expectedOps {
		if _, exists := impl.Operations[op]; !exists {
			t.Errorf("Expected operation %q to exist", op)
		}
	}
}

func TestCodeScanner_GetImplementedServiceNames(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a few service directories with Go files
	services := []string{"rds", "s3", "dynamodb"}
	for _, svc := range services {
		svcDir := filepath.Join(tmpDir, svc)
		if err := os.MkdirAll(svcDir, 0o755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		goFile := filepath.Join(svcDir, "service.go")
		if err := os.WriteFile(goFile, []byte("package "+svc+"\n"), 0o644); err != nil {
			t.Fatalf("Failed to create Go file: %v", err)
		}
	}

	// Create a directory without Go files (should not be included)
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	scanner := NewCodeScanner(tmpDir)
	names, err := scanner.GetImplementedServiceNames()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(names) != len(services) {
		t.Errorf("Expected %d services, got %d", len(services), len(names))
	}
}

func TestCodeScanner_ExtractFunctionName(t *testing.T) {
	scanner := &CodeScanner{}

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "Selector expression (method call)",
			code:     "s.handleCreate(ctx, req)",
			expected: "handleCreate",
		},
		{
			name:     "Identifier (function call)",
			code:     "doSomething(arg)",
			expected: "doSomething",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the expression
			expr, err := parser.ParseExpr(tt.code)
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			call, ok := expr.(*ast.CallExpr)
			if !ok {
				t.Fatalf("Expected CallExpr, got %T", expr)
			}

			got := scanner.extractFunctionName(call)
			if got != tt.expected {
				t.Errorf("extractFunctionName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCodeScanner_ScanSwitchStmt_ActionIdentifier(t *testing.T) {
	scanner := &CodeScanner{}
	impl := &models.ServiceImplementation{
		Name:       "test",
		Operations: make(map[string]*models.ImplementedOperation),
	}
	fset := token.NewFileSet()

	// Create a switch statement on "action" variable
	code := `package test
func test() {
	switch action {
	case "CreateResource":
		handleCreate()
	case "DeleteResource":
		handleDelete()
	}
}
`
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find the switch statement
	ast.Inspect(file, func(n ast.Node) bool {
		if switchStmt, ok := n.(*ast.SwitchStmt); ok {
			scanner.scanSwitchStmt(impl, fset, switchStmt, "test.go")
			return false
		}
		return true
	})

	if len(impl.Operations) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(impl.Operations))
	}

	if _, exists := impl.Operations["CreateResource"]; !exists {
		t.Error("Expected 'CreateResource' operation to exist")
	}
	if _, exists := impl.Operations["DeleteResource"]; !exists {
		t.Error("Expected 'DeleteResource' operation to exist")
	}
}

func TestCodeScanner_ScanSwitchStmt_NonActionIdentifier(t *testing.T) {
	scanner := &CodeScanner{}
	impl := &models.ServiceImplementation{
		Name:       "test",
		Operations: make(map[string]*models.ImplementedOperation),
	}
	fset := token.NewFileSet()

	// Create a switch statement on a non-action variable (should be ignored)
	code := `package test
func test() {
	switch someVar {
	case "Value1":
		doThing1()
	case "Value2":
		doThing2()
	}
}
`
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find the switch statement
	ast.Inspect(file, func(n ast.Node) bool {
		if switchStmt, ok := n.(*ast.SwitchStmt); ok {
			scanner.scanSwitchStmt(impl, fset, switchStmt, "test.go")
			return false
		}
		return true
	})

	if len(impl.Operations) != 0 {
		t.Errorf("Expected 0 operations for non-action switch, got %d", len(impl.Operations))
	}
}

func TestCodeScanner_ScanSwitchStmt_OperationVariant(t *testing.T) {
	scanner := &CodeScanner{}
	impl := &models.ServiceImplementation{
		Name:       "test",
		Operations: make(map[string]*models.ImplementedOperation),
	}
	fset := token.NewFileSet()

	// Create a switch statement on "operation" variable (should also work)
	code := `package test
func test() {
	switch operation {
	case "GetItem":
		getItem()
	case "PutItem":
		putItem()
	}
}
`
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find the switch statement
	ast.Inspect(file, func(n ast.Node) bool {
		if switchStmt, ok := n.(*ast.SwitchStmt); ok {
			scanner.scanSwitchStmt(impl, fset, switchStmt, "test.go")
			return false
		}
		return true
	})

	if len(impl.Operations) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(impl.Operations))
	}

	if _, exists := impl.Operations["GetItem"]; !exists {
		t.Error("Expected 'GetItem' operation to exist")
	}
	if _, exists := impl.Operations["PutItem"]; !exists {
		t.Error("Expected 'PutItem' operation to exist")
	}
}

func TestCodeScanner_ScanSwitchStmt_CallExprWithActionName(t *testing.T) {
	scanner := &CodeScanner{}
	impl := &models.ServiceImplementation{
		Name:       "test",
		Operations: make(map[string]*models.ImplementedOperation),
	}
	fset := token.NewFileSet()

	// Create a switch statement on a function call with "action" in name
	code := `package test
func test() {
	switch getAction() {
	case "ListBuckets":
		listBuckets()
	}
}
`
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find the switch statement
	ast.Inspect(file, func(n ast.Node) bool {
		if switchStmt, ok := n.(*ast.SwitchStmt); ok {
			scanner.scanSwitchStmt(impl, fset, switchStmt, "test.go")
			return false
		}
		return true
	})

	if len(impl.Operations) != 1 {
		t.Errorf("Expected 1 operation, got %d", len(impl.Operations))
	}
}

func TestCodeScanner_ScanSwitchStmt_CallExprWithoutActionName(t *testing.T) {
	scanner := &CodeScanner{}
	impl := &models.ServiceImplementation{
		Name:       "test",
		Operations: make(map[string]*models.ImplementedOperation),
	}
	fset := token.NewFileSet()

	// Create a switch statement on a function call without "action" in name (should be ignored)
	code := `package test
func test() {
	switch getSomething() {
	case "Value1":
		doThing1()
	}
}
`
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find the switch statement
	ast.Inspect(file, func(n ast.Node) bool {
		if switchStmt, ok := n.(*ast.SwitchStmt); ok {
			scanner.scanSwitchStmt(impl, fset, switchStmt, "test.go")
			return false
		}
		return true
	})

	if len(impl.Operations) != 0 {
		t.Errorf("Expected 0 operations for non-action function call, got %d", len(impl.Operations))
	}
}
