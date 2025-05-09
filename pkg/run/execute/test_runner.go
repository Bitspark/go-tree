package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
	"bitspark.dev/go-tree/pkg/io/resolve"
)

// TestRunner executes tests
type TestRunner struct {
	Resolver     ModuleResolver
	Materializer ModuleMaterializer
	Executor     Executor
	Generator    CodeGenerator
	Processor    ResultProcessor
}

// NewTestRunner creates a new test runner with default components
func NewTestRunner(resolver ModuleResolver, materializer ModuleMaterializer) *TestRunner {
	return &TestRunner{
		Resolver:     resolver,
		Materializer: materializer,
		Executor:     NewGoExecutor(),
		Generator:    NewTypeAwareGenerator(),
		Processor:    NewJsonResultProcessor(),
	}
}

// WithExecutor sets the executor to use
func (r *TestRunner) WithExecutor(executor Executor) *TestRunner {
	r.Executor = executor
	return r
}

// WithGenerator sets the code generator to use
func (r *TestRunner) WithGenerator(generator CodeGenerator) *TestRunner {
	r.Generator = generator
	return r
}

// WithProcessor sets the result processor to use
func (r *TestRunner) WithProcessor(processor ResultProcessor) *TestRunner {
	r.Processor = processor
	return r
}

// ExecuteModuleTests runs all tests in a module
func (r *TestRunner) ExecuteModuleTests(
	module *typesys.Module,
	testFlags ...string) (*TestResult, error) {

	if module == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	// Use materializer to create an execution environment
	opts := materialize.MaterializeOptions{
		DependencyPolicy: materialize.DirectDependenciesOnly,
		ReplaceStrategy:  materialize.RelativeReplace,
		LayoutStrategy:   materialize.FlatLayout,
		RunGoModTidy:     true,
		EnvironmentVars:  make(map[string]string),
	}

	// Create a materialized environment
	// Instead of calling a specific method on the materializer, we'll create an environment
	// and let the executor handle the module
	env := materialize.NewEnvironment(filepath.Join(os.TempDir(), module.Path), false)
	for k, v := range opts.EnvironmentVars {
		env.SetEnvVar(k, v)
	}

	// Execute tests in the environment
	result, err := r.Executor.ExecuteTest(env, module, "", testFlags...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tests: %w", err)
	}

	return result, nil
}

// ExecutePackageTests runs all tests in a specific package
func (r *TestRunner) ExecutePackageTests(
	module *typesys.Module,
	pkgPath string,
	testFlags ...string) (*TestResult, error) {

	if module == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	// Check if the package exists
	if _, ok := module.Packages[pkgPath]; !ok {
		return nil, fmt.Errorf("package %s not found in module %s", pkgPath, module.Path)
	}

	// Create a materialized environment
	env := materialize.NewEnvironment(filepath.Join(os.TempDir(), module.Path), false)

	// Execute tests in the specific package
	result, err := r.Executor.ExecuteTest(env, module, pkgPath, testFlags...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tests: %w", err)
	}

	return result, nil
}

// ExecuteSpecificTest runs a specific test function
func (r *TestRunner) ExecuteSpecificTest(
	module *typesys.Module,
	pkgPath string,
	testName string) (*TestResult, error) {

	if module == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	// Check if the package exists
	pkg, ok := module.Packages[pkgPath]
	if !ok {
		return nil, fmt.Errorf("package %s not found in module %s", pkgPath, module.Path)
	}

	// Find the test symbol
	var testSymbol *typesys.Symbol
	for _, sym := range pkg.Symbols {
		if sym.Kind == typesys.KindFunction && strings.HasPrefix(sym.Name, "Test") && sym.Name == testName {
			testSymbol = sym
			break
		}
	}

	if testSymbol == nil {
		return nil, fmt.Errorf("test function %s not found in package %s", testName, pkgPath)
	}

	// Create a materialized environment
	env := materialize.NewEnvironment(filepath.Join(os.TempDir(), module.Path), false)

	// Prepare test flags to run only the specific test
	testFlags := []string{"-v", "-run", "^" + testName + "$"}

	// Execute the specific test
	result, err := r.Executor.ExecuteTest(env, module, pkgPath, testFlags...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute test: %w", err)
	}

	return result, nil
}

// ResolveAndExecuteModuleTests resolves a module and runs all its tests
func (r *TestRunner) ResolveAndExecuteModuleTests(
	modulePath string,
	testFlags ...string) (*TestResult, error) {

	// Use resolver to get the module
	module, err := r.Resolver.ResolveModule(modulePath, "", resolve.ResolveOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module: %w", err)
	}

	// Resolve dependencies
	if err := r.Resolver.ResolveDependencies(module, 1); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Execute tests for the resolved module
	return r.ExecuteModuleTests(module, testFlags...)
}

// ResolveAndExecutePackageTests resolves a module and runs tests for a specific package
func (r *TestRunner) ResolveAndExecutePackageTests(
	modulePath string,
	pkgPath string,
	testFlags ...string) (*TestResult, error) {

	// Use resolver to get the module
	module, err := r.Resolver.ResolveModule(modulePath, "", resolve.ResolveOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module: %w", err)
	}

	// Resolve dependencies
	if err := r.Resolver.ResolveDependencies(module, 1); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Execute tests for the resolved package
	return r.ExecutePackageTests(module, pkgPath, testFlags...)
}
