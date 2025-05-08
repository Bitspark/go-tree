package execute

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"bitspark.dev/go-tree/pkg/typesys"
)

// GoExecutor implements ModuleExecutor for Go modules with type awareness
type GoExecutor struct {
	// EnableCGO determines whether CGO is enabled during execution
	EnableCGO bool

	// AdditionalEnv contains additional environment variables
	AdditionalEnv []string

	// WorkingDir specifies a custom working directory (defaults to module directory)
	WorkingDir string
}

// NewGoExecutor creates a new type-aware Go executor
func NewGoExecutor() *GoExecutor {
	return &GoExecutor{
		EnableCGO: true,
	}
}

// Execute runs a go command in the module's directory
func (g *GoExecutor) Execute(module *typesys.Module, args ...string) (ExecutionResult, error) {
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
func (g *GoExecutor) ExecuteTest(module *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error) {
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

	// Enhance with type system information - will be implemented further with type-aware system
	if module != nil && pkgPath != "" {
		pkg := findPackage(module, pkgPath)
		if pkg != nil {
			result.TestedSymbols = findTestedSymbols(pkg, result.Tests)
		}
	}

	return result, nil
}

// ExecuteFunc calls a specific function in the module with type checking
func (g *GoExecutor) ExecuteFunc(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	if module == nil {
		return nil, errors.New("module cannot be nil")
	}

	if funcSymbol == nil {
		return nil, errors.New("function symbol cannot be nil")
	}

	// This will be implemented in the TypeAwareCodeGenerator
	// For now, return a placeholder error
	return nil, fmt.Errorf("type-aware function execution not yet implemented for: %s", funcSymbol.Name)
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

// findPackage finds a package in the module by path
func findPackage(module *typesys.Module, pkgPath string) *typesys.Package {
	// Handle relative paths like "./..."
	if strings.HasPrefix(pkgPath, "./") {
		// Try to find the package by checking all packages
		for _, pkg := range module.Packages {
			relativePath := strings.TrimPrefix(pkg.ImportPath, module.Path+"/")
			if strings.HasPrefix(relativePath, strings.TrimPrefix(pkgPath, "./")) {
				return pkg
			}
		}
		return nil
	}

	// Direct package lookup
	pkg, ok := module.Packages[pkgPath]
	if ok {
		return pkg
	}

	// Try with module path prefix
	fullPath := module.Path
	if pkgPath != "" {
		fullPath = module.Path + "/" + pkgPath
	}
	return module.Packages[fullPath]
}

// findTestedSymbols finds symbols being tested
func findTestedSymbols(pkg *typesys.Package, testNames []string) []*typesys.Symbol {
	symbols := make([]*typesys.Symbol, 0)

	// This naive implementation assumes test names are in the format TestXxx where Xxx is the function name
	// We'll improve this with the analyzer later
	for _, test := range testNames {
		if len(test) <= 4 {
			continue // "Test" is 4 characters, so we need more than that
		}

		// Extract the function name being tested
		funcName := test[4:] // Remove "Test" prefix

		// Look for symbols that match this name
		for _, file := range pkg.Files {
			for _, symbol := range file.Symbols {
				if symbol.Kind == typesys.KindFunction && symbol.Name == funcName {
					symbols = append(symbols, symbol)
					break
				}
			}
		}
	}

	return symbols
}
