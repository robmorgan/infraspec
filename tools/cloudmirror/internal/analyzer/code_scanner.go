package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
)

// CodeScanner scans InfraSpec service implementations
type CodeScanner struct {
	servicesPath string
}

// NewCodeScanner creates a new code scanner
func NewCodeScanner(servicesPath string) *CodeScanner {
	return &CodeScanner{servicesPath: servicesPath}
}

// ScanService scans a service implementation for supported operations
func (s *CodeScanner) ScanService(serviceName string) (*models.ServiceImplementation, error) {
	servicePath := filepath.Join(s.servicesPath, serviceName)

	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		return nil, nil // Service not implemented, return nil without error
	}

	impl := &models.ServiceImplementation{
		Name:       serviceName,
		Path:       servicePath,
		Operations: make(map[string]*models.ImplementedOperation),
	}

	// Parse all Go files in the service directory
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, servicePath, func(info os.FileInfo) bool {
		// Skip test files
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			s.scanFile(impl, fset, file, filename)
		}
	}

	return impl, nil
}

func (s *CodeScanner) scanFile(impl *models.ServiceImplementation, fset *token.FileSet, file *ast.File, filename string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Look for HandleRequest method
			if node.Name.Name == "HandleRequest" {
				s.scanHandleRequest(impl, fset, node, filename)
			}
		}
		return true
	})
}

func (s *CodeScanner) scanHandleRequest(impl *models.ServiceImplementation, fset *token.FileSet, fn *ast.FuncDecl, filename string) {
	if fn.Body == nil {
		return
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if switchStmt, ok := n.(*ast.SwitchStmt); ok {
			s.scanSwitchStmt(impl, fset, switchStmt, filename)
		}
		return true
	})
}

func (s *CodeScanner) scanSwitchStmt(impl *models.ServiceImplementation, fset *token.FileSet, switchStmt *ast.SwitchStmt, filename string) {
	// Check if this is switching on action
	isActionSwitch := false

	switch tag := switchStmt.Tag.(type) {
	case *ast.Ident:
		// Check for common action variable names
		actionNames := []string{"action", "operation", "op", "apiAction"}
		for _, name := range actionNames {
			if tag.Name == name {
				isActionSwitch = true
				break
			}
		}
	case *ast.CallExpr:
		// Check if this is a function that extracts an action
		// Only accept calls to functions with "action" or "operation" in the name
		funcName := s.extractFunctionName(tag)
		funcNameLower := strings.ToLower(funcName)
		if strings.Contains(funcNameLower, "action") ||
			strings.Contains(funcNameLower, "operation") ||
			strings.Contains(funcNameLower, "getaction") ||
			strings.Contains(funcNameLower, "extractaction") {
			isActionSwitch = true
		}
	}

	if !isActionSwitch {
		return
	}

	for _, clause := range switchStmt.Body.List {
		caseClause, ok := clause.(*ast.CaseClause)
		if !ok {
			continue
		}

		for _, expr := range caseClause.List {
			if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				actionName := strings.Trim(lit.Value, "\"")

				// Find the handler function being called
				handlerName := s.findHandlerCall(caseClause.Body)

				pos := fset.Position(caseClause.Pos())
				impl.Operations[actionName] = &models.ImplementedOperation{
					Name:        actionName,
					Handler:     handlerName,
					File:        filepath.Base(filename),
					Line:        pos.Line,
					Implemented: true,
				}
			}
		}
	}
}

func (s *CodeScanner) findHandlerCall(stmts []ast.Stmt) string {
	for _, stmt := range stmts {
		switch st := stmt.(type) {
		case *ast.ReturnStmt:
			for _, result := range st.Results {
				if call, ok := result.(*ast.CallExpr); ok {
					return s.extractFunctionName(call)
				}
			}
		case *ast.ExprStmt:
			if call, ok := st.X.(*ast.CallExpr); ok {
				return s.extractFunctionName(call)
			}
		case *ast.AssignStmt:
			for _, rhs := range st.Rhs {
				if call, ok := rhs.(*ast.CallExpr); ok {
					return s.extractFunctionName(call)
				}
			}
		}
	}
	return ""
}

func (s *CodeScanner) extractFunctionName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		return fn.Sel.Name
	case *ast.Ident:
		return fn.Name
	}
	return ""
}

// ScanAllServices scans all services in the services directory
func (s *CodeScanner) ScanAllServices() (map[string]*models.ServiceImplementation, error) {
	services := make(map[string]*models.ServiceImplementation)

	entries, err := os.ReadDir(s.servicesPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		serviceName := entry.Name()
		impl, err := s.ScanService(serviceName)
		if err != nil {
			// Log but continue
			continue
		}

		if impl != nil {
			services[serviceName] = impl
		}
	}

	return services, nil
}

// GetImplementedServiceNames returns the names of all implemented services
func (s *CodeScanner) GetImplementedServiceNames() ([]string, error) {
	var names []string

	entries, err := os.ReadDir(s.servicesPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has Go files
			servicePath := filepath.Join(s.servicesPath, entry.Name())
			files, _ := filepath.Glob(filepath.Join(servicePath, "*.go"))
			if len(files) > 0 {
				names = append(names, entry.Name())
			}
		}
	}

	return names, nil
}
