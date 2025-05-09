package execute

import (
	"fmt"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
	"bitspark.dev/go-tree/pkg/io/resolve"
)

// ModuleResolver defines a minimal interface for resolving modules
type ModuleResolver interface {
	// ResolveModule resolves a module by path and version
	ResolveModule(path, version string, opts interface{}) (*typesys.Module, error)

	// ResolveDependencies resolves dependencies for a module
	ResolveDependencies(module *typesys.Module, depth int) error
}

// ModuleMaterializer defines a minimal interface for materializing modules
type ModuleMaterializer interface {
	// MaterializeMultipleModules materializes multiple modules into an environment
	MaterializeMultipleModules(modules []*typesys.Module, opts materialize.MaterializeOptions) (*materialize.Environment, error)
}

// FunctionRunner executes individual functions
type FunctionRunner struct {
	Resolver     ModuleResolver
	Materializer ModuleMaterializer
	Executor     Executor
	Generator    CodeGenerator
	Processor    ResultProcessor
	Security     SecurityPolicy
}

// NewFunctionRunner creates a new function runner with default components
func NewFunctionRunner(resolver ModuleResolver, materializer ModuleMaterializer) *FunctionRunner {
	return &FunctionRunner{
		Resolver:     resolver,
		Materializer: materializer,
		Executor:     NewGoExecutor(),
		Generator:    NewTypeAwareGenerator(),
		Processor:    NewJsonResultProcessor(),
		Security:     NewStandardSecurityPolicy(),
	}
}

// WithExecutor sets the executor to use
func (r *FunctionRunner) WithExecutor(executor Executor) *FunctionRunner {
	r.Executor = executor
	return r
}

// WithGenerator sets the code generator to use
func (r *FunctionRunner) WithGenerator(generator CodeGenerator) *FunctionRunner {
	r.Generator = generator
	return r
}

// WithProcessor sets the result processor to use
func (r *FunctionRunner) WithProcessor(processor ResultProcessor) *FunctionRunner {
	r.Processor = processor
	return r
}

// WithSecurity sets the security policy to use
func (r *FunctionRunner) WithSecurity(security SecurityPolicy) *FunctionRunner {
	r.Security = security
	return r
}

// ExecuteFunc executes a function using materialization
func (r *FunctionRunner) ExecuteFunc(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	args ...interface{}) (interface{}, error) {

	if module == nil || funcSymbol == nil {
		return nil, fmt.Errorf("module and function symbol cannot be nil")
	}

	// Generate wrapper code
	code, err := r.Generator.GenerateFunctionWrapper(module, funcSymbol, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate wrapper code: %w", err)
	}

	// Create a temporary module
	tmpModule, err := createTempModule(module.Path, code)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary module: %w", err)
	}

	// Use materializer to create an execution environment
	opts := materialize.MaterializeOptions{
		DependencyPolicy: materialize.DirectDependenciesOnly,
		ReplaceStrategy:  materialize.RelativeReplace,
		LayoutStrategy:   materialize.FlatLayout,
		RunGoModTidy:     true,
		EnvironmentVars:  make(map[string]string),
	}

	// Apply security policy to environment options
	if r.Security != nil {
		for k, v := range r.Security.GetEnvironmentVariables() {
			opts.EnvironmentVars[k] = v
		}
	}

	// Materialize the environment with the main module and dependencies
	env, err := r.Materializer.MaterializeMultipleModules(
		[]*typesys.Module{tmpModule, module}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to materialize environment: %w", err)
	}
	defer env.Cleanup()

	// Execute in the materialized environment
	mainFile := filepath.Join(env.ModulePaths[tmpModule.Path], "main.go")
	execResult, err := r.Executor.Execute(env, []string{"go", "run", mainFile})
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %w", err)
	}

	// Process the result
	result, err := r.Processor.ProcessFunctionResult(execResult, funcSymbol)
	if err != nil {
		return nil, fmt.Errorf("failed to process result: %w", err)
	}

	return result, nil
}

// ResolveAndExecuteFunc resolves a function by name and executes it
func (r *FunctionRunner) ResolveAndExecuteFunc(
	modulePath string,
	pkgPath string,
	funcName string,
	args ...interface{}) (interface{}, error) {

	// Use resolver to get the module
	module, err := r.Resolver.ResolveModule(modulePath, "", resolve.ResolveOptions{
		IncludeTests:   false,
		IncludePrivate: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module: %w", err)
	}

	// Resolve dependencies
	if err := r.Resolver.ResolveDependencies(module, 1); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Find the function symbol
	pkg, ok := module.Packages[pkgPath]
	if !ok {
		return nil, fmt.Errorf("package %s not found", pkgPath)
	}

	var funcSymbol *typesys.Symbol
	for _, sym := range pkg.Symbols {
		if sym.Kind == typesys.KindFunction && sym.Name == funcName {
			funcSymbol = sym
			break
		}
	}

	if funcSymbol == nil {
		return nil, fmt.Errorf("function %s not found in package %s", funcName, pkgPath)
	}

	// Execute the resolved function
	return r.ExecuteFunc(module, funcSymbol, args...)
}

// Helper functions

// createTempModule creates a temporary module with a single main.go file
func createTempModule(basePath string, code string) (*typesys.Module, error) {
	// Create a module with a name that won't conflict
	wrapperModulePath := basePath + "_wrapper"

	// Create the module
	module := typesys.NewModule("")
	module.Path = wrapperModulePath

	// Create a package for the wrapper
	pkg := typesys.NewPackage(module, "main", wrapperModulePath)
	module.Packages[wrapperModulePath] = pkg

	// Create a file for the wrapper
	// Note: We're assuming File has fields Path and Package.
	// The actual file content will be written to disk by the materializer.
	file := &typesys.File{
		Path:    "main.go",
		Package: pkg,
	}

	// Store the code separately as we'll need it later
	// The materializer will need to write this content to the filesystem
	pkg.Files["main.go"] = file

	return module, nil
}
