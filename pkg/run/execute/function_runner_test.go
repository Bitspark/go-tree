package execute

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
)

// MockRegistry implements a simple mock of the registry interface
type MockRegistry struct {
	modules      map[string]*MockRegistryModule
	queriedPaths map[string]bool
}

// MockRegistryModule represents a module in the mock registry
type MockRegistryModule struct {
	ImportPath     string
	FilesystemPath string
	IsLocal        bool
	Module         *typesys.Module
}

// NewMockRegistry creates a new mock registry
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		modules:      make(map[string]*MockRegistryModule),
		queriedPaths: make(map[string]bool),
	}
}

// RegisterModule adds a module to the mock registry
func (r *MockRegistry) RegisterModule(importPath, fsPath string, isLocal bool) error {
	r.modules[importPath] = &MockRegistryModule{
		ImportPath:     importPath,
		FilesystemPath: fsPath,
		IsLocal:        isLocal,
	}
	return nil
}

// FindModule checks if a module exists in the registry by import path
func (r *MockRegistry) FindModule(importPath string) (interface{}, bool) {
	r.queriedPaths[importPath] = true
	module, ok := r.modules[importPath]
	return module, ok
}

// FindByPath checks if a module exists in the registry by filesystem path
func (r *MockRegistry) FindByPath(fsPath string) (interface{}, bool) {
	// Simple implementation for mock - just check all modules
	for _, mod := range r.modules {
		if mod.FilesystemPath == fsPath {
			r.queriedPaths[mod.ImportPath] = true
			return mod, true
		}
	}
	return nil, false
}

// WasQueried checks if a path was queried during tests
func (r *MockRegistry) WasQueried(path string) bool {
	return r.queriedPaths[path]
}

// GetImportPath returns the import path
func (m *MockRegistryModule) GetImportPath() string {
	return m.ImportPath
}

// GetFilesystemPath returns the filesystem path
func (m *MockRegistryModule) GetFilesystemPath() string {
	return m.FilesystemPath
}

// GetModule returns the module
func (m *MockRegistryModule) GetModule() *typesys.Module {
	return m.Module
}

// MockResolver is a mock implementation of ModuleResolver
type MockResolver struct {
	Modules  map[string]*typesys.Module
	Registry *MockRegistry
}

func (r *MockResolver) ResolveModule(path, version string, opts interface{}) (*typesys.Module, error) {
	// First try the registry if available
	if r.Registry != nil {
		if module, ok := r.Registry.FindModule(path); ok {
			if mockModule, ok := module.(*MockRegistryModule); ok && mockModule.Module != nil {
				return mockModule.Module, nil
			}
		}
	}

	// Fall back to direct lookup
	module, ok := r.Modules[path]
	if !ok {
		return createFunctionRunnerMockModule(), nil // Return a default module if not found
	}
	return module, nil
}

func (r *MockResolver) ResolveDependencies(module *typesys.Module, depth int) error {
	return nil
}

// GetRegistry returns the registry if available
func (r *MockResolver) GetRegistry() interface{} {
	return r.Registry
}

// Additional methods required by the resolve.Resolver interface
func (r *MockResolver) AddDependency(from, to *typesys.Module) error {
	return nil
}

// MockMaterializer is a mock implementation of ModuleMaterializer
type MockMaterializer struct{}

func (m *MockMaterializer) MaterializeMultipleModules(modules []*typesys.Module, opts materialize.MaterializeOptions) (*materialize.Environment, error) {
	env := materialize.NewEnvironment("test-dir", false)
	for _, module := range modules {
		env.ModulePaths[module.Path] = "test-dir/" + module.Path
	}
	return env, nil
}

// Additional methods required by the materialize.Materializer interface
func (m *MockMaterializer) Materialize(module *typesys.Module, opts materialize.MaterializeOptions) (*materialize.Environment, error) {
	env := materialize.NewEnvironment("test-dir", false)
	env.ModulePaths[module.Path] = "test-dir/" + module.Path
	return env, nil
}

// MockExecutor is a mock implementation of Executor interface
type MockExecutor struct {
	ExecuteResult *ExecutionResult
	TestResult    *TestResult
	LastEnvVars   map[string]string
	LastCommand   []string
}

func (e *MockExecutor) Execute(env *materialize.Environment, command []string) (*ExecutionResult, error) {
	// Track the last environment and command for assertions
	e.LastCommand = command
	e.LastEnvVars = make(map[string]string)

	// Copy environment variables for testing
	for k, v := range env.EnvVars {
		e.LastEnvVars[k] = v
	}

	return e.ExecuteResult, nil
}

func (e *MockExecutor) ExecuteTest(env *materialize.Environment, module *typesys.Module, pkgPath string, testFlags ...string) (*TestResult, error) {
	return e.TestResult, nil
}

func (e *MockExecutor) ExecuteFunc(env *materialize.Environment, module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	return 42, nil // Always return 42 for tests
}

// MockProcessor implements the ResultProcessor interface for testing
type MockProcessor struct {
	ProcessResult interface{}
	ProcessError  error
}

func (p *MockProcessor) ProcessFunctionResult(result *ExecutionResult, funcSymbol *typesys.Symbol) (interface{}, error) {
	return p.ProcessResult, p.ProcessError
}

func (p *MockProcessor) ProcessTestResult(result *ExecutionResult, testSymbol *typesys.Symbol) (*TestResult, error) {
	return &TestResult{}, p.ProcessError
}

// TestFunctionRunner tests using the mock runner
func TestFunctionRunner(t *testing.T) {
	// Create a mock resolver with registry support
	registry := NewMockRegistry()
	resolver := &MockResolver{
		Modules:  make(map[string]*typesys.Module),
		Registry: registry,
	}

	// Create mock module
	module := createFunctionRunnerMockModule()
	resolver.Modules["github.com/test/simplemath"] = module

	// Register in the registry
	registry.RegisterModule("github.com/test/simplemath", "test-dir/simplemath", true)

	// Set up the resolver to return our module
	registry.modules["github.com/test/simplemath"].Module = module

	// Create a function runner
	runner := NewFunctionRunner(resolver, &MockMaterializer{})

	// Use mocks for execution and processing
	executor := &MockExecutor{
		ExecuteResult: &ExecutionResult{
			StdOut:   `{"result": 8}`,
			StdErr:   "",
			ExitCode: 0,
		},
	}

	processor := &MockProcessor{
		ProcessResult: float64(8),
	}

	runner.WithExecutor(executor)
	runner.WithProcessor(processor)

	// Add security policy
	runner.WithSecurity(NewStandardSecurityPolicy())

	// Test execution
	result, err := runner.ResolveAndExecuteFunc(
		"github.com/test/simplemath",
		"github.com/test/simplemath",
		"Add", 5, 3)

	// Validate results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != float64(8) {
		t.Errorf("Expected result 8, got: %v", result)
	}

	// Verify registry was queried
	if !registry.WasQueried("github.com/test/simplemath") {
		t.Error("Registry was not queried")
	}
}

// TestFunctionRunner_ExecuteFunc tests executing a function directly
func TestFunctionRunner_ExecuteFunc(t *testing.T) {
	// Create mocks
	resolver := &MockResolver{
		Modules: map[string]*typesys.Module{},
	}
	materializer := &MockMaterializer{}

	// Create a function runner with the mocks
	runner := NewFunctionRunner(resolver, materializer)

	// Use a mock executor that returns a known result
	mockExecutor := &MockExecutor{
		ExecuteResult: &ExecutionResult{
			StdOut:   "42",
			StdErr:   "",
			ExitCode: 0,
		},
	}
	runner.WithExecutor(mockExecutor)

	// Get a mock module and function symbol
	module := createFunctionRunnerMockModule()

	// The symbol should be directly accessible by key
	funcSymbol := module.Packages["github.com/test/simplemath"].Symbols["Add"]

	if funcSymbol == nil {
		t.Fatal("Failed to find Add function in mock module")
	}

	// Execute the function
	result, err := runner.ExecuteFunc(module, funcSymbol, 5, 3)

	// Check the result
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The mock processor will convert the string "42" to a float64
	if result != float64(42) {
		t.Errorf("Expected result 42, got: %v", result)
	}
}

// TestFunctionRunner_ResolveAndExecuteFunc tests resolving and executing a function by name
func TestFunctionRunner_ResolveAndExecuteFunc(t *testing.T) {
	// Create a mock module and add it to the resolver
	module := createFunctionRunnerMockModule()
	resolver := &MockResolver{
		Modules: map[string]*typesys.Module{
			"github.com/test/simplemath": module,
		},
	}
	materializer := &MockMaterializer{}

	// Create a function runner with the mocks
	runner := NewFunctionRunner(resolver, materializer)

	// Use a mock executor that returns a known result
	mockExecutor := &MockExecutor{
		ExecuteResult: &ExecutionResult{
			StdOut:   "42",
			StdErr:   "",
			ExitCode: 0,
		},
	}
	runner.WithExecutor(mockExecutor)

	// Resolve and execute the function
	result, err := runner.ResolveAndExecuteFunc(
		"github.com/test/simplemath",
		"github.com/test/simplemath",
		"Add",
		5, 3)

	// Check the result
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The mock processor will convert the string "42" to a float64
	if result != float64(42) {
		t.Errorf("Expected result 42, got: %v", result)
	}
}

// Helper function to create a mock module for testing
func createFunctionRunnerMockModule() *typesys.Module {
	module := typesys.NewModule("test-dir/simplemath")
	module.Path = "github.com/test/simplemath"

	// Create a package
	pkg := typesys.NewPackage(module, "simplemath", "github.com/test/simplemath")
	module.Packages["github.com/test/simplemath"] = pkg

	// Create an Add function symbol
	addFunc := &typesys.Symbol{
		Name:    "Add",
		Kind:    typesys.KindFunction,
		Package: pkg,
		// Description removed as it's not in the struct
	}

	// Add to package's symbol map with a unique key
	pkg.Symbols["Add"] = addFunc

	return module
}
