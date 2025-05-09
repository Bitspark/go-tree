package execute

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// Sandbox provides a secure environment for executing code
type Sandbox struct {
	// Configuration options
	AllowNetwork bool
	AllowFileIO  bool
	MemoryLimit  int64
	TimeLimit    int // In seconds

	// Module being executed
	Module *typesys.Module

	// Base directory for temporary files
	TempDir string

	// Keep temporary files for debugging
	KeepTempFiles bool

	// Code generator for type-aware execution
	generator *TypeAwareCodeGenerator
}

// NewSandbox creates a new sandbox for the given module
func NewSandbox(module *typesys.Module) *Sandbox {
	return &Sandbox{
		AllowNetwork:  false,
		AllowFileIO:   false,
		MemoryLimit:   102400000, // 100MB
		TimeLimit:     10,        // 10 seconds
		Module:        module,
		KeepTempFiles: false,
		generator:     NewTypeAwareCodeGenerator(module),
	}
}

// Execute runs code in the sandbox with type checking
func (s *Sandbox) Execute(code string) (*ExecutionResult, error) {
	// Create a temporary directory
	tempDir, createErr := s.createTempDir()
	if createErr != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", createErr)
	}

	// Clean up temporary directory unless KeepTempFiles is true
	if !s.KeepTempFiles {
		defer func() {
			if cleanErr := os.RemoveAll(tempDir); cleanErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory %s: %v\n", tempDir, cleanErr)
			}
		}()
	}

	// Create a temp file for the code
	mainFile := filepath.Join(tempDir, "main.go")
	if writeErr := os.WriteFile(mainFile, []byte(code), 0600); writeErr != nil {
		return nil, fmt.Errorf("failed to write temporary code file: %w", writeErr)
	}

	// Check if the code imports from the module - simple check for module name in imports
	needsModule := s.Module != nil && strings.Contains(code, s.Module.Path)

	// Create an appropriate go.mod file
	var goModContent string
	if needsModule {
		// Create a go.mod file with a replace directive for the module
		goModContent = fmt.Sprintf(`module sandbox

go 1.18

require %s v0.0.0
replace %s => %s
`, s.Module.Path, s.Module.Path, s.Module.Dir)
	} else {
		// Create a simple go.mod for standalone code
		goModContent = `module sandbox

go 1.18
`
	}

	goModFile := filepath.Join(tempDir, "go.mod")
	if writeErr := os.WriteFile(goModFile, []byte(goModContent), 0600); writeErr != nil {
		return nil, fmt.Errorf("failed to write go.mod file: %w", writeErr)
	}

	// Execute the code
	// Validate mainFile to prevent command injection by ensuring it's within our tempDir
	mainFileAbs, pathErr1 := filepath.Abs(mainFile)
	tempDirAbs, pathErr2 := filepath.Abs(tempDir)
	if pathErr1 != nil || pathErr2 != nil || !strings.HasPrefix(mainFileAbs, tempDirAbs) {
		return nil, fmt.Errorf("invalid file path: must be within sandbox directory")
	}

	cmd := exec.Command("go", "run", mainFile) // #nosec G204 - mainFile is validated as being within our controlled temp directory
	cmd.Dir = tempDir

	// Set up sandbox restrictions
	env := os.Environ()

	// Add memory limit if supported on the platform
	// Note: This is very platform-specific and may not work everywhere
	if s.MemoryLimit > 0 {
		env = append(env, fmt.Sprintf("GOMEMLIMIT=%d", s.MemoryLimit))
	}

	// Disable network if not allowed
	if !s.AllowNetwork {
		// On some platforms, you might set up network namespaces or other restrictions
		// For simplicity, we'll just set an environment variable and rely on the code
		// to respect it
		env = append(env, "SANDBOX_NETWORK=disabled")
	}

	// Disable file I/O if not allowed
	if !s.AllowFileIO {
		// Similar to network restrictions, this is platform-specific
		env = append(env, "SANDBOX_FILEIO=disabled")
	}

	cmd.Env = env

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set up a timeout
	runChan := make(chan error, 1)
	go func() {
		runChan <- cmd.Run()
	}()

	// Wait for completion or timeout
	var err error
	select {
	case err = <-runChan:
		// Command completed normally
	case <-time.After(time.Duration(s.TimeLimit) * time.Second):
		// Command timed out
		if cmd.Process != nil {
			if killErr := cmd.Process.Kill(); killErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to kill timed out process: %v\n", killErr)
			}
		}
		err = fmt.Errorf("execution timed out after %d seconds", s.TimeLimit)
	}

	// Create execution result
	result := &ExecutionResult{
		Command:  "go run " + mainFile,
		StdOut:   stdout.String(),
		StdErr:   stderr.String(),
		ExitCode: 0,
		Error:    err,
	}

	// Parse the exit code if available
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}

	return result, nil
}

// ExecuteFunction runs a specific function in the sandbox
func (s *Sandbox) ExecuteFunction(funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	if funcSymbol == nil {
		return nil, fmt.Errorf("function symbol cannot be nil")
	}

	// Generate wrapper code
	wrapperCode, genErr := s.generator.GenerateExecWrapper(funcSymbol, args...)
	if genErr != nil {
		return nil, fmt.Errorf("failed to generate execution wrapper: %w", genErr)
	}

	// Execute the generated code
	result, execErr := s.Execute(wrapperCode)
	if execErr != nil {
		return nil, fmt.Errorf("execution failed: %w", execErr)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("function execution failed with exit code %d: %s",
			result.ExitCode, result.StdErr)
	}

	// The result is in the stdout as JSON
	// In a real implementation, we'd parse the JSON and convert it back to Go objects
	// For this simplified implementation, we'll just return the raw stdout
	return strings.TrimSpace(result.StdOut), nil
}

// createTempDir creates a temporary directory for sandbox execution
func (s *Sandbox) createTempDir() (string, error) {
	baseDir := s.TempDir
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	return os.MkdirTemp(baseDir, "gosandbox-")
}
