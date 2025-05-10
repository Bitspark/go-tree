package execute

import (
	"fmt"
	"os"
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

	// Create a temporary directory for the wrapper
	wrapperDir, err := os.MkdirTemp("", "go-tree-wrapper-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(wrapperDir)

	// Create wrapper module path
	wrapperModulePath := module.Path + "_wrapper"

	// Make sure we have absolute paths for replacements
	moduleAbsDir, err := filepath.Abs(module.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for module: %w", err)
	}

	// Create an explicit go.mod with the replacement directive for the target module
	goModContent := fmt.Sprintf(`module %s

go 1.19

require %s v0.0.0

replace %s => %s
`,
		wrapperModulePath,
		funcSymbol.Package.ImportPath,
		funcSymbol.Package.ImportPath,
		moduleAbsDir)

	// Write the go.mod and main.go files directly
	if err := os.WriteFile(filepath.Join(wrapperDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write go.mod: %w", err)
	}

	if err := os.WriteFile(filepath.Join(wrapperDir, "main.go"), []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write main.go: %w", err)
	}

	// Create a temporary environment for execution
	env := &materialize.Environment{
		RootDir: wrapperDir, // Use wrapper dir as root
		ModulePaths: map[string]string{
			wrapperModulePath: wrapperDir,
			module.Path:       moduleAbsDir,
		},
		IsTemporary: true,
		EnvVars:     make(map[string]string),
	}

	// Apply security policy to environment
	if r.Security != nil {
		for k, v := range r.Security.GetEnvironmentVariables() {
			env.EnvVars[k] = v
		}
	}

	// Save generated code for debugging
	debugCode := fmt.Sprintf("\n--- Generated wrapper code ---\n%s\n--- go.mod ---\n%s\n",
		code, goModContent)
	os.WriteFile(filepath.Join(wrapperDir, "debug.txt"), []byte(debugCode), 0644)

	// Set the working directory for the executor if it's a GoExecutor
	if goExec, ok := r.Executor.(*GoExecutor); ok {
		goExec.WorkingDir = wrapperDir
	}

	// Execute in the materialized environment with proper working directory
	execResult, err := r.Executor.Execute(env, []string{"go", "run", "."})
	if err != nil {
		// If execution fails, try to read the debug file for more information
		return nil, fmt.Errorf("failed to execute function: %w\nworking dir: %s", err, wrapperDir)
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

// createTempModule creates a temporary module with a simple main.go and go.mod file
// that explicitly requires the target module at a placeholder version (v0.0.0)
func createTempModule(basePath string, mainCode string, dependencies ...string) (*typesys.Module, error) {
	// Create a module with a name that won't conflict
	wrapperModulePath := basePath + "_wrapper"

	// Create the module
	module := typesys.NewModule("")
	module.Path = wrapperModulePath

	// The content will be written to disk by the materializer
	// and the go.mod will be created when we call writeWrapperFiles
	// in ExecuteFunc

	// Create a package for the wrapper
	pkg := typesys.NewPackage(module, "main", wrapperModulePath)
	module.Packages[wrapperModulePath] = pkg

	// Create main.go file
	mainFile := typesys.NewFile("main.go", pkg)
	pkg.Files["main.go"] = mainFile

	// Create go.mod file
	goModFile := typesys.NewFile("go.mod", pkg)
	pkg.Files["go.mod"] = goModFile

	return module, nil
}

// writeWrapperFiles writes the wrapper files to disk
func writeWrapperFiles(dir string, mainCode string, modulePath string, dependencyPath string, replacementPath string) error {
	// Create go.mod content
	goModContent := fmt.Sprintf("module %s\n\ngo 1.16\n\n", modulePath)

	// Add requires for dependencies
	goModContent += "require (\n"
	goModContent += fmt.Sprintf("\t%s v0.0.0\n", dependencyPath)
	goModContent += ")\n\n"

	// Add replace directive for the dependency
	goModContent += "replace (\n"
	goModContent += fmt.Sprintf("\t%s => %s\n", dependencyPath, replacementPath)
	goModContent += ")\n"

	// Write main.go
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainCode), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	// Write go.mod
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	return nil
}
