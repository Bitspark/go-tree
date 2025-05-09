package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/io/materialize"
)

// CodeEvaluator evaluates arbitrary code
type CodeEvaluator struct {
	Materializer ModuleMaterializer // Changed to use the simplified interface
	Executor     Executor
	Security     SecurityPolicy
}

// NewCodeEvaluator creates a new code evaluator with default components
func NewCodeEvaluator(materializer ModuleMaterializer) *CodeEvaluator {
	return &CodeEvaluator{
		Materializer: materializer,
		Executor:     NewGoExecutor(),
		Security:     NewStandardSecurityPolicy(),
	}
}

// WithExecutor sets the executor to use
func (e *CodeEvaluator) WithExecutor(executor Executor) *CodeEvaluator {
	e.Executor = executor
	return e
}

// WithSecurity sets the security policy to use
func (e *CodeEvaluator) WithSecurity(security SecurityPolicy) *CodeEvaluator {
	e.Security = security
	return e
}

// EvaluateGoCode evaluates arbitrary Go code in a sandboxed environment
func (e *CodeEvaluator) EvaluateGoCode(code string) (*ExecutionResult, error) {
	// Create a temporary directory for the code
	tmpDir, err := os.MkdirTemp("", "go-eval-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the code to a temporary file
	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code to file: %w", err)
	}

	// Create a materialized environment
	env := materialize.NewEnvironment(tmpDir, false)

	// Apply security policy
	if e.Security != nil {
		if err := e.Security.ApplyToEnvironment(env); err != nil {
			return nil, fmt.Errorf("failed to apply security policy: %w", err)
		}
	}

	// Execute the code
	result, err := e.Executor.Execute(env, []string{"go", "run", mainFile})
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	return result, nil
}

// EvaluateGoPackage evaluates a complete Go package in a sandboxed environment
func (e *CodeEvaluator) EvaluateGoPackage(packageDir string, mainFile string) (*ExecutionResult, error) {
	// Check if the package directory exists
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("package directory does not exist: %s", packageDir)
	}

	// Create a materialized environment
	env := materialize.NewEnvironment(packageDir, false)

	// Apply security policy
	if e.Security != nil {
		if err := e.Security.ApplyToEnvironment(env); err != nil {
			return nil, fmt.Errorf("failed to apply security policy: %w", err)
		}
	}

	// Execute the main file in the package
	mainPath := filepath.Join(packageDir, mainFile)
	result, err := e.Executor.Execute(env, []string{"go", "run", mainPath})
	if err != nil {
		return nil, fmt.Errorf("failed to execute package: %w", err)
	}

	return result, nil
}

// EvaluateGoScript runs a Go script (single file with dependencies)
func (e *CodeEvaluator) EvaluateGoScript(scriptPath string, args ...string) (*ExecutionResult, error) {
	// Check if the script file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script file does not exist: %s", scriptPath)
	}

	// Get the directory containing the script
	scriptDir := filepath.Dir(scriptPath)

	// Create a materialized environment
	env := materialize.NewEnvironment(scriptDir, false)

	// Apply security policy
	if e.Security != nil {
		if err := e.Security.ApplyToEnvironment(env); err != nil {
			return nil, fmt.Errorf("failed to apply security policy: %w", err)
		}
	}

	// Prepare the command with arguments
	cmdArgs := append([]string{"go", "run", scriptPath}, args...)

	// Execute the script
	result, err := e.Executor.Execute(env, cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	return result, nil
}
