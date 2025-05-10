package execute

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// GoExecutor executes Go commands
type GoExecutor struct {
	// Enable CGO during execution
	EnableCGO bool

	// Environment variables for command execution
	EnvVars map[string]string

	// Working directory override
	WorkingDir string

	// Security policy for execution
	Security SecurityPolicy

	// Timeout for command execution in seconds (0 means no timeout)
	Timeout int
}

// NewGoExecutor creates a new Go executor with default settings
func NewGoExecutor() *GoExecutor {
	return &GoExecutor{
		EnableCGO: true,
		EnvVars:   make(map[string]string),
		Timeout:   30, // 30 second default timeout
	}
}

// WithSecurity sets the security policy
func (e *GoExecutor) WithSecurity(security SecurityPolicy) *GoExecutor {
	e.Security = security
	return e
}

// WithTimeout sets the execution timeout
func (e *GoExecutor) WithTimeout(seconds int) *GoExecutor {
	e.Timeout = seconds
	return e
}

// Execute runs a command in the given environment
func (e *GoExecutor) Execute(env Environment, command []string) (*ExecutionResult, error) {
	if env == nil {
		return nil, errors.New("environment cannot be nil")
	}

	if len(command) == 0 {
		return nil, errors.New("command cannot be empty")
	}

	// Apply security policy to command if available
	if e.Security != nil {
		command = e.Security.ApplyToExecution(command)
	}

	// Prepare the command
	cmd := exec.Command(command[0], command[1:]...)

	// Set working directory
	workDir := e.WorkingDir
	if workDir == "" {
		workDir = env.GetPath()
	}
	cmd.Dir = workDir

	// Setup environment variables
	cmd.Env = os.Environ()
	if !e.EnableCGO {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
	}

	// Apply security policy to environment
	if e.Security != nil {
		if err := e.Security.ApplyToEnvironment(env); err != nil {
			return nil, fmt.Errorf("failed to apply security policy: %w", err)
		}

		// Add security environment variables
		for k, v := range e.Security.GetEnvironmentVariables() {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Add executor environment variables
	for k, v := range e.EnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute with timeout if set
	var err error
	if e.Timeout > 0 {
		err = runWithTimeout(cmd, time.Duration(e.Timeout)*time.Second)
	} else {
		err = cmd.Run()
	}

	// Create result
	result := &ExecutionResult{
		Command:  strings.Join(command, " "),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Error:    nil,
	}

	// Handle error
	if err != nil {
		result.Error = err
		// Get exit code if available
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	return result, nil
}

// ExecuteTest runs tests in a package
func (e *GoExecutor) ExecuteTest(env Environment, module *typesys.Module, pkgPath string,
	testFlags ...string) (*TestResult, error) {

	if env == nil || module == nil {
		return nil, errors.New("environment and module cannot be nil")
	}

	// Prepare the command
	args := []string{"go", "test"}
	args = append(args, testFlags...)

	// Create the test result
	result := &TestResult{
		Package: pkgPath,
		Tests:   []string{},
		Passed:  0,
		Failed:  0,
		Output:  "",
	}

	// Execute the command
	execResult, err := e.Execute(env, args)
	if err != nil {
		result.Error = err
		return result, err
	}

	// Populate the result
	result.Output = execResult.Stdout + execResult.Stderr

	// Parse test output to count passes and failures
	if strings.Contains(execResult.Stdout, "ok") || strings.Contains(execResult.Stdout, "PASS") {
		// Tests passed
		result.Passed = countTests(execResult.Stdout)
	} else if strings.Contains(execResult.Stdout, "FAIL") {
		// Some tests failed
		result.Passed, result.Failed = parseTestResults(execResult.Stdout)
	}

	// Parse the test names
	result.Tests = parseTestNames(execResult.Stdout)

	// Set error if tests failed
	if result.Failed > 0 {
		result.Error = fmt.Errorf("%d tests failed", result.Failed)
	}

	return result, nil
}

// ExecuteFunc executes a function in the given environment
func (e *GoExecutor) ExecuteFunc(env Environment, module *typesys.Module,
	funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {

	if env == nil || module == nil || funcSymbol == nil {
		return nil, errors.New("environment, module, and function symbol cannot be nil")
	}

	// Create a code generator
	generator := NewTypeAwareGenerator()

	// Generate the wrapper code
	code, err := generator.GenerateFunctionWrapper(module, funcSymbol, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate wrapper code: %w", err)
	}

	// Create a temporary file for the wrapper
	wrapperFile := filepath.Join(env.GetPath(), "wrapper.go")
	if err := os.WriteFile(wrapperFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write wrapper file: %w", err)
	}
	defer os.Remove(wrapperFile)

	// Execute the wrapper
	execResult, err := e.Execute(env, []string{"go", "run", wrapperFile})
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %w", err)
	}

	// Process the result
	processor := NewJsonResultProcessor()
	result, err := processor.ProcessFunctionResult(execResult, funcSymbol)
	if err != nil {
		return nil, fmt.Errorf("failed to process result: %w", err)
	}

	return result, nil
}

// Helper functions

// runWithTimeout runs a command with a timeout
func runWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	if timeout <= 0 {
		return cmd.Run()
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Create a channel for the process to finish
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for the process to finish or timeout
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process after timeout: %w", err)
		}
		return fmt.Errorf("process killed after timeout of %v", timeout)
	}
}

// countTests counts the number of tests in the output
func countTests(output string) int {
	re := regexp.MustCompile(`(?m)^--- PASS: Test\w+`)
	return len(re.FindAllString(output, -1))
}

// parseTestResults parses the test output to count passes and failures
func parseTestResults(output string) (passed, failed int) {
	passRe := regexp.MustCompile(`(?m)^--- PASS: Test\w+`)
	failRe := regexp.MustCompile(`(?m)^--- FAIL: Test\w+`)

	passed = len(passRe.FindAllString(output, -1))
	failed = len(failRe.FindAllString(output, -1))

	return passed, failed
}

// parseTestNames extracts test names from output
func parseTestNames(output string) []string {
	re := regexp.MustCompile(`(?m)^--- (PASS|FAIL): (Test\w+)`)
	matches := re.FindAllStringSubmatch(output, -1)

	tests := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			tests = append(tests, match[2])
		}
	}

	return tests
}
