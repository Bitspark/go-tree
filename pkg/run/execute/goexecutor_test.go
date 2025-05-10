package execute

import (
	"bitspark.dev/go-tree/pkg/env"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoExecutor_Execute(t *testing.T) {
	// Create a temporary directory and write a test file
	tmpDir, err := os.MkdirTemp("", "executor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	code := `package main
import "fmt"
func main() { fmt.Println("Hello, world!") }`

	err = os.WriteFile(mainFile, []byte(code), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a real environment
	env := env.NewEnvironment(tmpDir, false)

	// Create executor and execute
	executor := NewGoExecutor()
	result, err := executor.Execute(env, []string{"go", "run", mainFile})

	// Verify results
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	if !strings.Contains(result.StdOut, "Hello, world!") {
		t.Errorf("Expected output to contain 'Hello, world!', got: %s", result.StdOut)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
}

func TestGoExecutor_ExecuteWithError(t *testing.T) {
	// Create a temporary directory and write an invalid Go file
	tmpDir, err := os.MkdirTemp("", "executor-test-error-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	invalidCode := `package main
func main() { undefinedFunction() }`

	err = os.WriteFile(mainFile, []byte(invalidCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a real environment
	env := env.NewEnvironment(tmpDir, false)

	// Create executor and execute
	executor := NewGoExecutor()
	result, err := executor.Execute(env, []string{"go", "run", mainFile})

	// We expect compilation error, but not a func execution error
	if err != nil {
		t.Fatalf("Execute should not return an error, but got: %v", err)
	}

	// Check for compilation error in the output
	if !strings.Contains(result.StdErr, "undefined") {
		t.Errorf("Expected output to contain compilation error, got: %s", result.StdErr)
	}

	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code, got: %d", result.ExitCode)
	}
}

func TestGoExecutor_WithSecurity(t *testing.T) {
	// Create a temporary directory and write a test file
	tmpDir, err := os.MkdirTemp("", "executor-test-security-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	code := `package main
import (
	"fmt"
	"os"
)
func main() { 
	fmt.Println("Security test")
	fmt.Println("SANDBOX_NETWORK:", os.Getenv("SANDBOX_NETWORK"))
}`

	err = os.WriteFile(mainFile, []byte(code), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a real environment
	env := env.NewEnvironment(tmpDir, false)

	// Create security policy
	security := NewStandardSecurityPolicy().WithAllowNetwork(false)

	// Create executor with security and execute
	executor := NewGoExecutor().WithSecurity(security)
	result, err := executor.Execute(env, []string{"go", "run", mainFile})

	// Verify results
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check that the security environment variable was set
	if !strings.Contains(result.StdOut, "SANDBOX_NETWORK: disabled") {
		t.Errorf("Expected output to contain SANDBOX_NETWORK: disabled, got: %s", result.StdOut)
	}
}

func TestGoExecutor_WithTimeout(t *testing.T) {
	// Create a temporary directory and write a test file with an infinite loop
	tmpDir, err := os.MkdirTemp("", "executor-test-timeout-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	code := `package main
import "time"
func main() { 
	// Infinite loop
	for {
		time.Sleep(100 * time.Millisecond)
	}
}`

	err = os.WriteFile(mainFile, []byte(code), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a real environment
	env := env.NewEnvironment(tmpDir, false)

	// Create executor with a short timeout and execute
	executor := NewGoExecutor().WithTimeout(1) // 1 second timeout
	result, err := executor.Execute(env, []string{"go", "run", mainFile})

	// We expect the command to be killed due to timeout
	if err != nil {
		t.Fatalf("Execute should not return an error, but got: %v", err)
	}

	if result.Error == nil || !strings.Contains(result.Error.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", result.Error)
	}
}
