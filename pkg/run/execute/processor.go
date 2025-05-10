package execute

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// JsonResultProcessor processes function results encoded as JSON
type JsonResultProcessor struct{}

// NewJsonResultProcessor creates a new JSON result processor
func NewJsonResultProcessor() *JsonResultProcessor {
	return &JsonResultProcessor{}
}

// ProcessFunctionResult processes the raw execution result into a typed value
func (p *JsonResultProcessor) ProcessFunctionResult(
	result *ExecutionResult,
	funcSymbol *typesys.Symbol) (interface{}, error) {

	if result == nil {
		return nil, fmt.Errorf("result cannot be nil")
	}

	// If execution failed, return the error
	if result.Error != nil {
		return nil, fmt.Errorf("execution failed: %w", result.Error)
	}

	// Parse the stdout as JSON
	jsonOutput := strings.TrimSpace(result.Stdout)
	if jsonOutput == "" {
		return nil, fmt.Errorf("empty result")
	}

	// Handle special "success" response
	if jsonOutput == `{"success":true}` {
		return nil, nil // Function returned void
	}

	// Unmarshal the JSON into a generic interface{}
	var value interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// For more advanced cases, we could use funcSymbol to determine the expected return type
	// and convert the result accordingly, but for now we'll return the generic value

	return value, nil
}

// ProcessTestResult processes test execution results
func (p *JsonResultProcessor) ProcessTestResult(
	result *ExecutionResult,
	testSymbol *typesys.Symbol) (*TestResult, error) {

	if result == nil {
		return nil, fmt.Errorf("result cannot be nil")
	}

	// Create a basic test result
	testResult := &TestResult{
		Package: "", // Will be populated below
		Tests:   []string{},
		Passed:  0,
		Failed:  0,
		Output:  result.Stdout + result.Stderr,
		Error:   result.Error,
	}

	// Extract test information from output
	testResult.Tests = extractTestNames(result.Stdout)
	testResult.Passed, testResult.Failed = countPassFail(result.Stdout)

	// Extract package name
	pkgName := extractPackageName(result.Stdout)
	if pkgName != "" {
		testResult.Package = pkgName
	} else if testSymbol != nil && testSymbol.Package != nil {
		testResult.Package = testSymbol.Package.ImportPath
	}

	// Extract coverage information
	testResult.Coverage = extractCoverage(result.Stdout)

	// If test symbol is provided, add it to the tested symbols
	if testSymbol != nil {
		testResult.TestedSymbols = []*typesys.Symbol{testSymbol}
	}

	return testResult, nil
}

// Helper functions

// extractTestNames extracts test names from Go test output
func extractTestNames(output string) []string {
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

// countPassFail counts passed and failed tests
func countPassFail(output string) (int, int) {
	passRe := regexp.MustCompile(`(?m)^--- PASS: Test\w+`)
	failRe := regexp.MustCompile(`(?m)^--- FAIL: Test\w+`)

	passed := len(passRe.FindAllString(output, -1))
	failed := len(failRe.FindAllString(output, -1))

	return passed, failed
}

// extractPackageName extracts the package name from test output
func extractPackageName(output string) string {
	re := regexp.MustCompile(`(?m)^ok\s+(\S+)`)
	match := re.FindStringSubmatch(output)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// extractCoverage extracts the code coverage percentage
func extractCoverage(output string) float64 {
	re := regexp.MustCompile(`(?m)coverage: (\d+\.\d+)% of statements`)
	match := re.FindStringSubmatch(output)
	if len(match) >= 2 {
		var coverage float64
		fmt.Sscanf(match[1], "%f", &coverage)
		return coverage
	}
	return 0.0
}
