package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
	"bitspark.dev/go-tree/pkg/io/resolve"
	"bitspark.dev/go-tree/pkg/run/execute"
	"bitspark.dev/go-tree/pkg/run/testing/common"
)

// UnifiedTestRunner provides unified test execution functionality
type UnifiedTestRunner struct {
	Executor  execute.Executor
	Generator execute.CodeGenerator
	Processor execute.ResultProcessor
}

// NewUnifiedTestRunner creates a new unified test runner
func NewUnifiedTestRunner(executor execute.Executor, generator execute.CodeGenerator, processor execute.ResultProcessor) *UnifiedTestRunner {
	if executor == nil {
		executor = execute.NewGoExecutor()
	}

	if generator == nil {
		generator = execute.NewTypeAwareGenerator()
	}

	if processor == nil {
		processor = execute.NewJsonResultProcessor()
	}

	return &UnifiedTestRunner{
		Executor:  executor,
		Generator: generator,
		Processor: processor,
	}
}

// ExecuteTest runs tests for a given module and package path
// This replaces the execute.Executor.ExecuteTest method
func (r *UnifiedTestRunner) ExecuteTest(env *materialize.Environment, module *typesys.Module,
	pkgPath string, testFlags ...string) (*common.TestResult, error) {
	// Create environment if none provided
	if env == nil {
		env = materialize.NewEnvironment(filepath.Join(os.TempDir(), module.Path), false)
	}

	// Prepare test command
	cmd := append([]string{"go", "test"}, testFlags...)
	if pkgPath != "" {
		cmd = append(cmd, pkgPath)
	}

	// Use the core executor to run the test command
	execResult, err := r.Executor.Execute(env, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tests: %w", err)
	}

	// Process test-specific output
	result := r.processTestOutput(execResult, module, pkgPath)
	return result, nil
}

// processTestOutput parses the output from 'go test' and extracts test results
func (r *UnifiedTestRunner) processTestOutput(result *execute.ExecutionResult, module *typesys.Module, pkgPath string) *common.TestResult {
	// Initialize test result
	testResult := &common.TestResult{
		Package:       pkgPath,
		Tests:         []string{},
		Passed:        0,
		Failed:        0,
		Output:        result.StdOut + result.StdErr,
		Error:         result.Error,
		TestedSymbols: []*typesys.Symbol{},
		Coverage:      0.0,
	}

	// Parse test output to identify tests and results
	testNameRegex := regexp.MustCompile(`--- (PASS|FAIL): (Test\w+) \(`)
	testMatches := testNameRegex.FindAllStringSubmatch(testResult.Output, -1)

	for _, match := range testMatches {
		status := match[1]
		testName := match[2]

		// Add test to the list
		testResult.Tests = append(testResult.Tests, testName)

		// Update pass/fail counts
		if status == "PASS" {
			testResult.Passed++
		} else {
			testResult.Failed++
		}

		// Try to find corresponding symbol
		if module != nil && pkgPath != "" {
			pkg, ok := module.Packages[pkgPath]
			if ok {
				// Find function being tested by parsing test name
				// TestXxx typically tests function Xxx
				funcName := strings.TrimPrefix(testName, "Test")
				for _, sym := range pkg.Symbols {
					if sym.Name == funcName && sym.Kind == typesys.KindFunction {
						testResult.TestedSymbols = append(testResult.TestedSymbols, sym)
						break
					}
				}
			}
		}
	}

	// Try to extract coverage information if present
	coverageRegex := regexp.MustCompile(`coverage: (\d+\.\d+)% of statements`)
	coverageMatch := coverageRegex.FindStringSubmatch(testResult.Output)
	if len(coverageMatch) > 1 {
		coverage, err := strconv.ParseFloat(coverageMatch[1], 64)
		if err == nil {
			testResult.Coverage = coverage
		}
	}

	return testResult
}

// ExecuteModuleTests runs all tests in a module
func (r *UnifiedTestRunner) ExecuteModuleTests(
	module *typesys.Module,
	testFlags ...string) (*common.TestResult, error) {

	if module == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	// Create a materialized environment
	env := materialize.NewEnvironment(filepath.Join(os.TempDir(), module.Path), false)

	// Execute tests in the environment
	return r.ExecuteTest(env, module, "", testFlags...)
}

// ExecutePackageTests runs all tests in a specific package
func (r *UnifiedTestRunner) ExecutePackageTests(
	module *typesys.Module,
	pkgPath string,
	testFlags ...string) (*common.TestResult, error) {

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
	return r.ExecuteTest(env, module, pkgPath, testFlags...)
}

// ExecuteSpecificTest runs a specific test function
func (r *UnifiedTestRunner) ExecuteSpecificTest(
	module *typesys.Module,
	pkgPath string,
	testName string) (*common.TestResult, error) {

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
	return r.ExecuteTest(env, module, pkgPath, testFlags...)
}

// ResolveAndExecuteModuleTests resolves a module and runs all its tests
func (r *UnifiedTestRunner) ResolveAndExecuteModuleTests(
	modulePath string,
	resolver execute.ModuleResolver,
	testFlags ...string) (*common.TestResult, error) {

	// Use resolver to get the module
	module, err := resolver.ResolveModule(modulePath, "", resolve.ResolveOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module: %w", err)
	}

	// Resolve dependencies
	if err := resolver.ResolveDependencies(module, 1); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Execute tests for the resolved module
	return r.ExecuteModuleTests(module, testFlags...)
}

// ResolveAndExecutePackageTests resolves a module and runs tests for a specific package
func (r *UnifiedTestRunner) ResolveAndExecutePackageTests(
	modulePath string,
	resolver execute.ModuleResolver,
	pkgPath string,
	testFlags ...string) (*common.TestResult, error) {

	// Use resolver to get the module
	module, err := resolver.ResolveModule(modulePath, "", resolve.ResolveOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module: %w", err)
	}

	// Resolve dependencies
	if err := resolver.ResolveDependencies(module, 1); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Execute tests for the resolved package
	return r.ExecutePackageTests(module, pkgPath, testFlags...)
}
