package execute

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"bitspark.dev/go-tree/pkg/core/module"
)

// GoExecutor implements ModuleExecutor for Go modules
type GoExecutor struct {
	// EnableCGO determines whether CGO is enabled during execution
	EnableCGO bool

	// AdditionalEnv contains additional environment variables
	AdditionalEnv []string

	// WorkingDir specifies a custom working directory (defaults to module directory)
	WorkingDir string
}

// NewGoExecutor creates a new Go executor
func NewGoExecutor() *GoExecutor {
	return &GoExecutor{
		EnableCGO: true,
	}
}

// Execute runs a go command in the module's directory
func (g *GoExecutor) Execute(module *module.Module, args ...string) (ExecutionResult, error) {
	if module == nil {
		return ExecutionResult{}, errors.New("module cannot be nil")
	}

	// Prepare command
	cmd := exec.Command("go", args...)

	// Set working directory
	workDir := g.WorkingDir
	if workDir == "" {
		workDir = module.Dir
	}
	cmd.Dir = workDir

	// Set environment
	env := os.Environ()
	if !g.EnableCGO {
		env = append(env, "CGO_ENABLED=0")
	}
	env = append(env, g.AdditionalEnv...)
	cmd.Env = env

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	// Create result
	result := ExecutionResult{
		Command:  "go " + strings.Join(args, " "),
		StdOut:   stdout.String(),
		StdErr:   stderr.String(),
		ExitCode: 0,
		Error:    nil,
	}

	// Handle error and exit code
	if err != nil {
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	return result, nil
}

// ExecuteTest runs tests for a package in the module
func (g *GoExecutor) ExecuteTest(module *module.Module, pkgPath string, testFlags ...string) (TestResult, error) {
	if module == nil {
		return TestResult{}, errors.New("module cannot be nil")
	}

	// Determine the package to test
	targetPkg := pkgPath
	if targetPkg == "" {
		targetPkg = "./..."
	}

	// Prepare test command
	args := append([]string{"test"}, testFlags...)
	args = append(args, targetPkg)

	// Run the test command
	execResult, err := g.Execute(module, args...)

	// Parse test results
	result := TestResult{
		Package: targetPkg,
		Output:  execResult.StdOut + execResult.StdErr,
		Error:   err,
	}

	// Count passed/failed tests
	result.Tests = parseTestNames(execResult.StdOut)

	// If we have verbose output, count passed/failed from output
	if containsFlag(testFlags, "-v") || containsFlag(testFlags, "-json") {
		passed, failed := countTestResults(execResult.StdOut)
		result.Passed = passed
		result.Failed = failed
	} else {
		// Without verbose output, we have to infer from error code
		if err == nil {
			result.Passed = len(result.Tests)
			result.Failed = 0
		} else {
			// At least one test failed, but we don't know which ones
			result.Failed = 1
			result.Passed = len(result.Tests) - result.Failed
		}
	}

	return result, nil
}

// ExecuteFunc calls a specific function in the module
func (g *GoExecutor) ExecuteFunc(module *module.Module, funcPath string, args ...interface{}) (interface{}, error) {
	// NOTE: This is a placeholder implementation, as properly executing a function
	// from a module requires runtime reflection or code generation techniques
	// that are beyond this initial implementation. In a full implementation,
	// this would:
	// 1. Generate a small program that imports the module
	// 2. Call the function with the provided arguments
	// 3. Serialize and return the results

	return nil, fmt.Errorf("function execution not implemented: %s", funcPath)
}

// Helper functions

// parseTestNames extracts test names from go test output
func parseTestNames(output string) []string {
	// Simple regex to match "--- PASS: TestName" or "--- FAIL: TestName"
	re := regexp.MustCompile(`--- (PASS|FAIL): (Test\w+)`)
	matches := re.FindAllStringSubmatch(output, -1)

	tests := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			tests = append(tests, match[2])
		}
	}

	return tests
}

// countTestResults counts passed and failed tests from output
func countTestResults(output string) (passed, failed int) {
	passRe := regexp.MustCompile(`--- PASS: `)
	failRe := regexp.MustCompile(`--- FAIL: `)

	passed = len(passRe.FindAllString(output, -1))
	failed = len(failRe.FindAllString(output, -1))

	return passed, failed
}

// containsFlag checks if a flag is present in the arguments
func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}
