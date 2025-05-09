package execute

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
)

// MockResolver is a mock implementation of ModuleResolver
type MockResolver struct {
	Modules map[string]*typesys.Module
}

func (r *MockResolver) ResolveModule(path, version string, opts interface{}) (*typesys.Module, error) {
	module, ok := r.Modules[path]
	if !ok {
		return createMockModule(), nil // Return a default module if not found
	}
	return module, nil
}

func (r *MockResolver) ResolveDependencies(module *typesys.Module, depth int) error {
	return nil
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

// TestFunctionRunner tests using the mock runner
func TestFunctionRunner(t *testing.T) {
	// Skip this test for now since we're still developing the interface
	t.Skip("Skipping TestFunctionRunner until interfaces are stable")
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
	module := createMockModule()
	var funcSymbol *typesys.Symbol
	for _, sym := range module.Packages["github.com/test/simplemath"].Symbols {
		if sym.Name == "Add" && sym.Kind == typesys.KindFunction {
			funcSymbol = sym
			break
		}
	}

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
	module := createMockModule()
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

// MockExecutor is a mock implementation of Executor interface
type MockExecutor struct {
	ExecuteResult *ExecutionResult
	TestResult    *TestResult
}

func (e *MockExecutor) Execute(env *materialize.Environment, command []string) (*ExecutionResult, error) {
	return e.ExecuteResult, nil
}

func (e *MockExecutor) ExecuteTest(env *materialize.Environment, module *typesys.Module, pkgPath string, testFlags ...string) (*TestResult, error) {
	return e.TestResult, nil
}

func (e *MockExecutor) ExecuteFunc(env *materialize.Environment, module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	return 42, nil // Always return 42 for tests
}
