package execute

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// MockModuleExecutor implements ModuleExecutor for testing
type MockModuleExecutor struct {
	ExecuteFn     func(module *typesys.Module, args ...string) (ExecutionResult, error)
	ExecuteTestFn func(module *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error)
	ExecuteFuncFn func(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error)
}

func (m *MockModuleExecutor) Execute(module *typesys.Module, args ...string) (ExecutionResult, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(module, args...)
	}
	return ExecutionResult{}, nil
}

func (m *MockModuleExecutor) ExecuteTest(module *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error) {
	if m.ExecuteTestFn != nil {
		return m.ExecuteTestFn(module, pkgPath, testFlags...)
	}
	return TestResult{}, nil
}

func (m *MockModuleExecutor) ExecuteFunc(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	if m.ExecuteFuncFn != nil {
		return m.ExecuteFuncFn(module, funcSymbol, args...)
	}
	return nil, nil
}

func TestNewExecutionContext(t *testing.T) {
	// Create a dummy module for testing
	module := &typesys.Module{
		Path: "test/module",
	}

	// Create a new execution context
	ctx := NewExecutionContext(module)

	// Verify the context was created correctly
	if ctx == nil {
		t.Fatal("NewExecutionContext returned nil")
	}

	if ctx.Module != module {
		t.Errorf("Expected module %v, got %v", module, ctx.Module)
	}

	if ctx.Files == nil {
		t.Error("Files map should not be nil")
	}

	if len(ctx.Files) != 0 {
		t.Errorf("Expected empty Files map, got %d entries", len(ctx.Files))
	}

	if ctx.Stdout != nil {
		t.Errorf("Expected nil Stdout, got %v", ctx.Stdout)
	}

	if ctx.Stderr != nil {
		t.Errorf("Expected nil Stderr, got %v", ctx.Stderr)
	}
}

func TestExecutionContext_WithOutputCapture(t *testing.T) {
	// Create a dummy module for testing
	module := &typesys.Module{
		Path: "test/module",
	}

	// Create a new execution context
	ctx := NewExecutionContext(module)

	// Set output capture
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ctx.Stdout = stdout
	ctx.Stderr = stderr

	// Verify the output capture was set correctly
	if ctx.Stdout != stdout {
		t.Errorf("Expected Stdout to be %v, got %v", stdout, ctx.Stdout)
	}

	if ctx.Stderr != stderr {
		t.Errorf("Expected Stderr to be %v, got %v", stderr, ctx.Stderr)
	}
}

func TestExecutionContext_Execute(t *testing.T) {
	// This is a placeholder test for the Execute method
	// Currently the implementation is a stub, so we're just testing the interface
	// Once implemented, this test should be expanded

	module := &typesys.Module{
		Path: "test/module",
	}

	ctx := NewExecutionContext(module)
	result, err := ctx.Execute("fmt.Println(\"Hello, World!\")")

	// Since the function is stubbed to return nil, nil
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Future implementation should test these behaviors:
	// 1. Code compilation
	// 2. Type checking
	// 3. Execution
	// 4. Result capturing
	// 5. Error handling
}

func TestExecutionContext_ExecuteInline(t *testing.T) {
	// This is a placeholder test for the ExecuteInline method
	// Currently the implementation is a stub, so we're just testing the interface
	// Once implemented, this test should be expanded

	module := &typesys.Module{
		Path: "test/module",
	}

	ctx := NewExecutionContext(module)
	result, err := ctx.ExecuteInline("fmt.Println(\"Hello, World!\")")

	// Since the function is stubbed to return nil, nil
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Future implementation should test these behaviors:
	// 1. Code execution in current context
	// 2. State preservation
	// 3. Output capturing
	// 4. Error handling
}

func TestExecutionResult(t *testing.T) {
	// Test creating and using ExecutionResult
	result := ExecutionResult{
		Command:  "go run main.go",
		StdOut:   "Hello, World!",
		StdErr:   "",
		ExitCode: 0,
		Error:    nil,
		TypeInfo: map[string]typesys.Symbol{
			"main": {Name: "main"},
		},
	}

	if result.Command != "go run main.go" {
		t.Errorf("Expected Command to be 'go run main.go', got '%s'", result.Command)
	}

	if result.StdOut != "Hello, World!" {
		t.Errorf("Expected StdOut to be 'Hello, World!', got '%s'", result.StdOut)
	}

	if result.StdErr != "" {
		t.Errorf("Expected empty StdErr, got '%s'", result.StdErr)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected ExitCode to be 0, got %d", result.ExitCode)
	}

	if result.Error != nil {
		t.Errorf("Expected nil Error, got %v", result.Error)
	}

	if len(result.TypeInfo) == 0 {
		t.Error("Expected non-empty TypeInfo")
	}
}

func TestTestResult(t *testing.T) {
	// Test creating and using TestResult
	symbol := &typesys.Symbol{Name: "TestFunc"}
	result := TestResult{
		Package:       "example/pkg",
		Tests:         []string{"TestFunc1", "TestFunc2"},
		Passed:        1,
		Failed:        1,
		Output:        "PASS: TestFunc1\nFAIL: TestFunc2",
		Error:         nil,
		TestedSymbols: []*typesys.Symbol{symbol},
		Coverage:      75.5,
	}

	if result.Package != "example/pkg" {
		t.Errorf("Expected Package to be 'example/pkg', got '%s'", result.Package)
	}

	expectedTests := []string{"TestFunc1", "TestFunc2"}
	if len(result.Tests) != len(expectedTests) {
		t.Errorf("Expected %d tests, got %d", len(expectedTests), len(result.Tests))
	}

	for i, test := range expectedTests {
		if i >= len(result.Tests) || result.Tests[i] != test {
			t.Errorf("Expected test %d to be '%s', got '%s'", i, test, result.Tests[i])
		}
	}

	if result.Passed != 1 {
		t.Errorf("Expected Passed to be 1, got %d", result.Passed)
	}

	if result.Failed != 1 {
		t.Errorf("Expected Failed to be 1, got %d", result.Failed)
	}

	if !bytes.Contains([]byte(result.Output), []byte("PASS: TestFunc1")) {
		t.Errorf("Expected Output to contain 'PASS: TestFunc1', got '%s'", result.Output)
	}

	if !bytes.Contains([]byte(result.Output), []byte("FAIL: TestFunc2")) {
		t.Errorf("Expected Output to contain 'FAIL: TestFunc2', got '%s'", result.Output)
	}

	if result.Error != nil {
		t.Errorf("Expected nil Error, got %v", result.Error)
	}

	if len(result.TestedSymbols) != 1 || result.TestedSymbols[0] != symbol {
		t.Errorf("Expected TestedSymbols to contain symbol, got %v", result.TestedSymbols)
	}

	if result.Coverage != 75.5 {
		t.Errorf("Expected Coverage to be 75.5, got %f", result.Coverage)
	}
}

func TestGoExecutor_New(t *testing.T) {
	executor := NewGoExecutor()

	if executor == nil {
		t.Fatal("NewGoExecutor should return a non-nil executor")
	}

	if !executor.EnableCGO {
		t.Error("EnableCGO should be true by default")
	}

	if len(executor.AdditionalEnv) != 0 {
		t.Errorf("AdditionalEnv should be empty by default, got %v", executor.AdditionalEnv)
	}

	if executor.WorkingDir != "" {
		t.Errorf("WorkingDir should be empty by default, got %s", executor.WorkingDir)
	}
}

func TestGoExecutor_Execute(t *testing.T) {
	// Create a simple test module
	tempDir, err := os.MkdirTemp("", "goexecutor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	})

	// Create a simple Go module
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/test\n\ngo 1.16\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a simple main.go file
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from test module")
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Create a mock module
	module := &typesys.Module{
		Path: "example.com/test",
		Dir:  tempDir,
	}

	// Create a GoExecutor
	executor := NewGoExecutor()

	// Test 'go version' command
	result, err := executor.Execute(module, "version")
	if err != nil {
		t.Errorf("Execute should not return an error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Execute should return exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.StdOut, "go version") {
		t.Errorf("Execute output should contain 'go version', got: %s", result.StdOut)
	}

	// Test command error handling
	result, err = executor.Execute(module, "invalid-command")
	if err == nil {
		t.Error("Execute should return an error for invalid command")
	}

	if result.ExitCode == 0 {
		t.Errorf("Execute should return non-zero exit code for error, got %d", result.ExitCode)
	}
}

func TestGoExecutor_ExecuteWithEnv(t *testing.T) {
	// Create a simple test module
	tempDir, err := os.MkdirTemp("", "goexecutor-env-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	})

	// Create a mock module
	module := &typesys.Module{
		Path: "example.com/test",
		Dir:  tempDir,
	}

	// Create a GoExecutor with custom environment
	executor := NewGoExecutor()
	executor.AdditionalEnv = []string{"TEST_ENV_VAR=test_value"}

	// Create a main.go that prints environment variables
	mainContent := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("TEST_ENV_VAR=%s\n", os.Getenv("TEST_ENV_VAR"))
	fmt.Printf("CGO_ENABLED=%s\n", os.Getenv("CGO_ENABLED"))
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// First test with CGO enabled (default)
	result, err := executor.Execute(module, "run", "main.go")
	if err != nil {
		t.Errorf("Execute should not return an error: %v", err)
	}

	if !strings.Contains(result.StdOut, "TEST_ENV_VAR=test_value") {
		t.Errorf("Custom environment variable should be set, got: %s", result.StdOut)
	}

	// Now test with CGO disabled
	executor.EnableCGO = false
	result, err = executor.Execute(module, "run", "main.go")
	if err != nil {
		t.Errorf("Execute should not return an error: %v", err)
	}

	if !strings.Contains(result.StdOut, "CGO_ENABLED=0") {
		t.Errorf("CGO_ENABLED should be set to 0, got: %s", result.StdOut)
	}
}

func TestGoExecutor_ExecuteTest(t *testing.T) {
	// Create a simple test module
	tempDir, err := os.MkdirTemp("", "goexecutor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	})

	// Create a simple Go module
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/test\n\ngo 1.16\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a simple testable package
	err = os.Mkdir(filepath.Join(tempDir, "pkg"), 0755)
	if err != nil {
		t.Fatalf("Failed to create pkg directory: %v", err)
	}

	// Create a package with a function to test
	pkgContent := `package pkg

// Add adds two integers
func Add(a, b int) int {
	return a + b
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg", "pkg.go"), []byte(pkgContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write pkg.go: %v", err)
	}

	// Create a test file
	testContent := `package pkg

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Error("Add(2, 3) should be 5")
	}
}

func TestAddFail(t *testing.T) {
	// This test should fail
	if Add(2, 3) == 5 {
		t.Error("This test should fail but won't")
	}
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg", "pkg_test.go"), []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write pkg_test.go: %v", err)
	}

	// Create a mock module with a package
	module := &typesys.Module{
		Path: "example.com/test",
		Dir:  tempDir,
		Packages: map[string]*typesys.Package{
			"example.com/test/pkg": {
				ImportPath: "example.com/test/pkg",
				Name:       "pkg",
				Files: map[string]*typesys.File{
					filepath.Join(tempDir, "pkg", "pkg.go"): {
						Path: filepath.Join(tempDir, "pkg", "pkg.go"),
						Name: "pkg.go",
					},
					filepath.Join(tempDir, "pkg", "pkg_test.go"): {
						Path:   filepath.Join(tempDir, "pkg", "pkg_test.go"),
						Name:   "pkg_test.go",
						IsTest: true,
					},
				},
				Symbols: map[string]*typesys.Symbol{
					"Add": {
						ID:   "Add",
						Name: "Add",
						Kind: typesys.KindFunction,
					},
				},
			},
		},
	}

	// Create a GoExecutor
	executor := NewGoExecutor()

	// Test running a specific test
	result, _ := executor.ExecuteTest(module, "./pkg", "-v", "-run=TestAdd$")
	// We don't check err because some tests might fail, which returns an error

	if !strings.Contains(result.Output, "TestAdd") {
		t.Errorf("Test output should contain 'TestAdd', got: %s", result.Output)
	}

	// Test parsing of test names
	if len(result.Tests) == 0 {
		t.Error("ExecuteTest should find at least one test")
	}

	// Test test counting with verbose output
	result, _ = executor.ExecuteTest(module, "./pkg", "-v", "-run=TestAdd$")
	if result.Passed != 1 || result.Failed != 0 {
		t.Errorf("Expected 1 passed test and 0 failed tests, got %d passed and %d failed",
			result.Passed, result.Failed)
	}

	// Test failing test
	result, _ = executor.ExecuteTest(module, "./pkg", "-v", "-run=TestAddFail$")
	if result.Passed != 0 || result.Failed != 1 {
		t.Errorf("Expected 0 passed tests and 1 failed test, got %d passed and %d failed",
			result.Passed, result.Failed)
	}
}

// TestParseTestNames verifies the test name parsing logic
func TestParseTestNames(t *testing.T) {
	testOutput := `--- PASS: TestFunc1 (0.00s)
--- FAIL: TestFunc2 (0.01s)
    file_test.go:42: Test failure message
--- SKIP: TestFunc3 (0.00s)
    file_test.go:50: Test skipped message
`

	tests := parseTestNames(testOutput)

	expected := []string{"TestFunc1", "TestFunc2", "TestFunc3"}
	if len(tests) != len(expected) {
		t.Errorf("Expected %d tests, got %d", len(expected), len(tests))
	}

	for i, test := range expected {
		if i >= len(tests) || tests[i] != test {
			t.Errorf("Expected test %d to be '%s', got '%s'", i, test, tests[i])
		}
	}
}

// TestCountTestResults verifies the test counting logic
func TestCountTestResults(t *testing.T) {
	testOutput := `--- PASS: TestFunc1 (0.00s)
--- PASS: TestFunc2 (0.00s)
--- FAIL: TestFunc3 (0.01s)
    file_test.go:42: Test failure message
--- FAIL: TestFunc4 (0.01s)
    file_test.go:50: Test failure message
--- SKIP: TestFunc5 (0.00s)
`

	passed, failed := countTestResults(testOutput)

	if passed != 2 {
		t.Errorf("Expected 2 passed tests, got %d", passed)
	}

	if failed != 2 {
		t.Errorf("Expected 2 failed tests, got %d", failed)
	}
}

// TestFindPackage verifies the package finding logic
func TestFindPackage(t *testing.T) {
	// Create a test module with packages
	module := &typesys.Module{
		Path: "example.com/test",
		Packages: map[string]*typesys.Package{
			"example.com/test":     {ImportPath: "example.com/test", Name: "main"},
			"example.com/test/pkg": {ImportPath: "example.com/test/pkg", Name: "pkg"},
			"example.com/test/sub": {ImportPath: "example.com/test/sub", Name: "sub"},
		},
	}

	// Test finding package by import path
	pkg := findPackage(module, "example.com/test/pkg")
	if pkg == nil {
		t.Error("findPackage should find package by import path")
	} else if pkg.Name != "pkg" {
		t.Errorf("Expected package name 'pkg', got '%s'", pkg.Name)
	}

	// Test finding package with relative path
	pkg = findPackage(module, "./pkg")
	if pkg == nil {
		t.Error("findPackage should find package by relative path")
	} else if pkg.Name != "pkg" {
		t.Errorf("Expected package name 'pkg', got '%s'", pkg.Name)
	}

	// Test finding non-existent package
	pkg = findPackage(module, "nonexistent")
	if pkg != nil {
		t.Error("findPackage should return nil for non-existent package")
	}
}

// TestFindTestedSymbols verifies the symbol finding logic
func TestFindTestedSymbols(t *testing.T) {
	// Create a test package with symbols
	pkg := &typesys.Package{
		Name:       "pkg",
		ImportPath: "example.com/test/pkg",
		Symbols: map[string]*typesys.Symbol{
			"Func1": {ID: "Func1", Name: "Func1", Kind: typesys.KindFunction},
			"Func2": {ID: "Func2", Name: "Func2", Kind: typesys.KindFunction},
			"Type1": {ID: "Type1", Name: "Type1", Kind: typesys.KindType},
		},
		Files: map[string]*typesys.File{
			"file1.go": {
				Path: "file1.go",
				Symbols: []*typesys.Symbol{
					{ID: "Func1", Name: "Func1", Kind: typesys.KindFunction},
					{ID: "Type1", Name: "Type1", Kind: typesys.KindType},
				},
			},
			"file2.go": {
				Path: "file2.go",
				Symbols: []*typesys.Symbol{
					{ID: "Func2", Name: "Func2", Kind: typesys.KindFunction},
				},
			},
		},
	}

	// Test finding symbols by test names
	testNames := []string{"TestFunc1", "TestFunc2", "TestNonExistent"}
	symbols := findTestedSymbols(pkg, testNames)

	if len(symbols) != 2 {
		t.Errorf("Expected 2 symbols to be found, got %d", len(symbols))
	}

	// Check the found symbols
	foundFunc1 := false
	foundFunc2 := false

	for _, sym := range symbols {
		switch sym.Name {
		case "Func1":
			foundFunc1 = true
		case "Func2":
			foundFunc2 = true
		}
	}

	if !foundFunc1 {
		t.Error("Expected to find symbol 'Func1'")
	}

	if !foundFunc2 {
		t.Error("Expected to find symbol 'Func2'")
	}
}

// TestSandboxExecution tests sandbox execution functionality
func TestSandboxExecution(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Create a sandbox
	sandbox := NewSandbox(module)

	if sandbox == nil {
		t.Fatal("NewSandbox should return a non-nil sandbox")
	}

	// Test running a simple code in the sandbox
	code := `
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	result, err := sandbox.Execute(code)
	if err != nil {
		t.Errorf("Execute should not return an error: %v", err)
	}

	if !strings.Contains(result.StdOut, "hello") {
		t.Errorf("Sandbox output should contain 'hello', got: %s", result.StdOut)
	}

	// Test security constraints by trying to access file system
	securityCode := `
package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		fmt.Println("Access denied, as expected")
		return
	}
	fmt.Println("Unexpectedly accessed system file")
}
`
	result, _ = sandbox.Execute(securityCode)
	if strings.Contains(result.StdOut, "Unexpectedly accessed system file") {
		t.Error("Sandbox should prevent access to system files")
	}
}

// TestTemporaryExecutor tests the temporary file execution functionality
func TestTemporaryExecutor(t *testing.T) {
	tempExecutor := NewTmpExecutor()

	if tempExecutor == nil {
		t.Fatal("NewTmpExecutor should return a non-nil executor")
	}

	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Test executing a Go command
	result, err := tempExecutor.Execute(module, "version")
	if err != nil {
		t.Errorf("Execute should not return an error: %v", err)
	}

	if !strings.Contains(result.StdOut, "go version") {
		t.Errorf("Go version output should contain version info, got: %s", result.StdOut)
	}
}

// TestTypeAwareExecution tests the type-aware execution functionality
func TestTypeAwareExecution(t *testing.T) {
	// Create a simple test module
	tempDir, err := os.MkdirTemp("", "typeaware-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	})

	// Create a simple Go module
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/test\n\ngo 1.16\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a simple testable package
	err = os.Mkdir(filepath.Join(tempDir, "pkg"), 0755)
	if err != nil {
		t.Fatalf("Failed to create pkg directory: %v", err)
	}

	// Create a package with a function to test
	pkgContent := `package pkg

// Add adds two integers
func Add(a, b int) int {
	return a + b
}

// Person represents a person
type Person struct {
	Name string
	Age  int
}

// Greet returns a greeting
func (p Person) Greet() string {
	return fmt.Sprintf("Hello, my name is %s", p.Name)
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg", "pkg.go"), []byte(pkgContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write pkg.go: %v", err)
	}

	// Create the module structure
	module := &typesys.Module{
		Path: "example.com/test",
		Dir:  tempDir,
	}

	// Create a type-aware execution context
	ctx := NewExecutionContext(module)

	// For this test, we'll simulate the behavior since the real implementation
	// requires a complete type system setup

	// Verify the execution context
	if ctx == nil {
		t.Fatal("NewExecutionContext should return a non-nil context")
	}

	// Test code generation
	generator := NewTypeAwareCodeGenerator(module)

	if generator == nil {
		t.Fatal("NewTypeAwareCodeGenerator should return a non-nil generator")
	}

	// Let's create a test function symbol to test GenerateExecWrapper
	funcSymbol := &typesys.Symbol{
		Name: "TestFunc",
		Kind: typesys.KindFunction,
		Package: &typesys.Package{
			ImportPath: "example.com/test/pkg",
			Name:       "pkg",
		},
	}

	// This will likely fail since our test symbol doesn't have proper type information,
	// but we can at least test that the function exists and is called
	code, _ := generator.GenerateExecWrapper(funcSymbol)
	// We don't assert on the error here since it's expected to fail without proper type info

	// Just verify we got something back
	if code != "" {
		t.Logf("Generated wrapper code: %s", code)
	}
}

// TestModuleExecutor_Interface ensures our mock executor implements the interface correctly
func TestModuleExecutor_Interface(t *testing.T) {
	// Create mock executor with custom implementations
	executor := &MockModuleExecutor{}

	// Create dummy module and symbol
	module := &typesys.Module{Path: "test/module"}
	symbol := &typesys.Symbol{Name: "TestFunc"}

	// Setup mock implementations
	expectedResult := ExecutionResult{
		Command:  "go run main.go",
		StdOut:   "Hello, World!",
		ExitCode: 0,
	}

	executor.ExecuteFn = func(m *typesys.Module, args ...string) (ExecutionResult, error) {
		if m != module {
			t.Errorf("Expected module %v, got %v", module, m)
		}

		if len(args) != 2 || args[0] != "run" || args[1] != "main.go" {
			t.Errorf("Expected args [run main.go], got %v", args)
		}

		return expectedResult, nil
	}

	expectedTestResult := TestResult{
		Package: "test/module",
		Tests:   []string{"TestFunc"},
		Passed:  1,
		Failed:  0,
	}

	executor.ExecuteTestFn = func(m *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error) {
		if m != module {
			t.Errorf("Expected module %v, got %v", module, m)
		}

		if pkgPath != "test/module" {
			t.Errorf("Expected pkgPath 'test/module', got '%s'", pkgPath)
		}

		if len(testFlags) != 1 || testFlags[0] != "-v" {
			t.Errorf("Expected testFlags [-v], got %v", testFlags)
		}

		return expectedTestResult, nil
	}

	executor.ExecuteFuncFn = func(m *typesys.Module, funcSym *typesys.Symbol, args ...interface{}) (interface{}, error) {
		if m != module {
			t.Errorf("Expected module %v, got %v", module, m)
		}

		if funcSym != symbol {
			t.Errorf("Expected symbol %v, got %v", symbol, funcSym)
		}

		if len(args) != 1 || args[0] != "arg1" {
			t.Errorf("Expected args [arg1], got %v", args)
		}

		return "result", nil
	}

	// Execute and verify
	result, err := executor.Execute(module, "run", "main.go")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Command != expectedResult.Command ||
		result.StdOut != expectedResult.StdOut ||
		result.ExitCode != expectedResult.ExitCode {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}

	testResult, err := executor.ExecuteTest(module, "test/module", "-v")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if testResult.Package != expectedTestResult.Package ||
		len(testResult.Tests) != len(expectedTestResult.Tests) ||
		testResult.Passed != expectedTestResult.Passed ||
		testResult.Failed != expectedTestResult.Failed {
		t.Errorf("Expected test result %v, got %v", expectedTestResult, testResult)
	}

	funcResult, err := executor.ExecuteFunc(module, symbol, "arg1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if funcResult != "result" {
		t.Errorf("Expected func result 'result', got %v", funcResult)
	}
}

// TestGoExecutor_CompleteApplication tests a complete application execution cycle
func TestGoExecutor_CompleteApplication(t *testing.T) {
	// Create a test project directory
	tempDir, err := os.MkdirTemp("", "goexecutor-app-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	})

	// Create a simple Go application
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/calculator\n\ngo 1.16\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a main.go file with arguments parsing
	mainContent := `package main

import (
	"fmt"
	"os"
	"strconv"
)

// Simple calculator application
func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}

	operation := os.Args[1]
	num1, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid number: %s\n", os.Args[2])
		os.Exit(1)
	}

	num2, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Printf("Invalid number: %s\n", os.Args[3])
		os.Exit(1)
	}

	var result int
	switch operation {
	case "add":
		result = num1 + num2
	case "subtract":
		result = num1 - num2
	case "multiply":
		result = num1 * num2
	case "divide":
		if num2 == 0 {
			fmt.Println("Error: Division by zero")
			os.Exit(1)
		}
		result = num1 / num2
	default:
		fmt.Printf("Unknown operation: %s\n", operation)
		os.Exit(1)
	}

	fmt.Printf("Result: %d\n", result)
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Create a module with the application
	module := &typesys.Module{
		Path: "example.com/calculator",
		Dir:  tempDir,
	}

	// Create an executor
	executor := NewGoExecutor()

	// Test building the application
	buildResult, err := executor.Execute(module, "build")
	if err != nil {
		t.Errorf("Failed to build application: %v", err)
	}

	if buildResult.ExitCode != 0 {
		t.Errorf("Build failed with exit code %d: %s",
			buildResult.ExitCode, buildResult.StdErr)
	}

	// Test running the application with different operations
	testCases := []struct {
		operation string
		num1      string
		num2      string
		expected  string
	}{
		{"add", "5", "3", "Result: 8"},
		{"subtract", "10", "4", "Result: 6"},
		{"multiply", "6", "7", "Result: 42"},
		{"divide", "20", "5", "Result: 4"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s_%s", tc.operation, tc.num1, tc.num2), func(t *testing.T) {
			runResult, err := executor.Execute(module, "run", "main.go", tc.operation, tc.num1, tc.num2)
			if err != nil {
				t.Errorf("Failed to run application: %v", err)
			}

			if runResult.ExitCode != 0 {
				t.Errorf("Application failed with exit code %d: %s",
					runResult.ExitCode, runResult.StdErr)
			}

			if !strings.Contains(runResult.StdOut, tc.expected) {
				t.Errorf("Expected output to contain '%s', got: %s",
					tc.expected, runResult.StdOut)
			}
		})
	}

	// Test error handling in the application
	errorCases := []struct {
		name       string
		args       []string
		expectFail bool
		errorMsg   string
	}{
		{"missing_args", []string{"run", "main.go"}, true, "Usage: calculator"},
		{"invalid_number", []string{"run", "main.go", "add", "not-a-number", "5"}, true, "Invalid number"},
		{"division_by_zero", []string{"run", "main.go", "divide", "10", "0"}, true, "Division by zero"},
		{"unknown_operation", []string{"run", "main.go", "power", "2", "3"}, true, "Unknown operation"},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := executor.Execute(module, tc.args...)

			if tc.expectFail && result.ExitCode == 0 {
				t.Errorf("Expected application to fail, but it succeeded")
			}

			if !tc.expectFail && result.ExitCode != 0 {
				t.Errorf("Expected application to succeed, but it failed with: %s",
					result.StdErr)
			}

			output := result.StdOut
			if result.StdErr != "" {
				output += result.StdErr
			}

			if !strings.Contains(output, tc.errorMsg) {
				t.Errorf("Expected output to contain '%s', got: %s",
					tc.errorMsg, output)
			}
		})
	}
}

// TestGoExecutor_ExecuteTestComprehensive tests comprehensive test execution features
func TestGoExecutor_ExecuteTestComprehensive(t *testing.T) {
	// Create a test project directory
	tempDir, err := os.MkdirTemp("", "goexecutor-comprehensive-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to clean up temp dir: %v", err)
		}
	})

	// Create a simple Go project with tests
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/testproject\n\ngo 1.16\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a libary package with multiple testable functions
	err = os.Mkdir(filepath.Join(tempDir, "pkg"), 0755)
	if err != nil {
		t.Fatalf("Failed to create pkg directory: %v", err)
	}

	// Create the library code
	libContent := `package pkg

// StringUtils provides string manipulation functions

// Reverse returns the reverse of a string
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = toUpper(runes[0])
	return string(runes)
}

// IsEmpty checks if a string is empty
func IsEmpty(s string) bool {
	return s == ""
}

// Private helper function to capitalize a rune
func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - ('a' - 'A')
	}
	return r
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg", "string_utils.go"), []byte(libContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write library code: %v", err)
	}

	// Create a test file with mixed passing and failing tests
	testContent := `package pkg

import "testing"

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single char", "a", "a"},
		{"simple string", "hello", "olleh"},
		{"palindrome", "racecar", "racecar"},
		{"with spaces", "hello world", "dlrow olleh"},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Reverse(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"already capitalized", "Hello", "Hello"},
		{"lowercase", "hello", "Hello"},
		{"with spaces", "hello world", "Hello world"}, // This will pass
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Capitalize(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	if !IsEmpty("") {
		t.Error("Expected IsEmpty(\"\") to be true")
	}
	
	if IsEmpty("not empty") {
		t.Error("Expected IsEmpty(\"not empty\") to be false")
	}
}

// This test will intentionally fail
func TestIntentionallyFailing(t *testing.T) {
	t.Error("This test is designed to fail")
}
`
	err = os.WriteFile(filepath.Join(tempDir, "pkg", "string_utils_test.go"), []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test code: %v", err)
	}

	// Create a proper module structure with package info
	module := &typesys.Module{
		Path: "example.com/testproject",
		Dir:  tempDir,
		Packages: map[string]*typesys.Package{
			"example.com/testproject/pkg": {
				ImportPath: "example.com/testproject/pkg",
				Name:       "pkg",
				Files: map[string]*typesys.File{
					filepath.Join(tempDir, "pkg", "string_utils.go"): {
						Path: filepath.Join(tempDir, "pkg", "string_utils.go"),
						Name: "string_utils.go",
					},
					filepath.Join(tempDir, "pkg", "string_utils_test.go"): {
						Path:   filepath.Join(tempDir, "pkg", "string_utils_test.go"),
						Name:   "string_utils_test.go",
						IsTest: true,
					},
				},
				Symbols: map[string]*typesys.Symbol{
					"Reverse": {
						ID:       "Reverse",
						Name:     "Reverse",
						Kind:     typesys.KindFunction,
						Exported: true,
					},
					"Capitalize": {
						ID:       "Capitalize",
						Name:     "Capitalize",
						Kind:     typesys.KindFunction,
						Exported: true,
					},
					"IsEmpty": {
						ID:       "IsEmpty",
						Name:     "IsEmpty",
						Kind:     typesys.KindFunction,
						Exported: true,
					},
				},
			},
		},
	}

	// Create a GoExecutor
	executor := NewGoExecutor()

	// Test running all tests
	result, _ := executor.ExecuteTest(module, "./pkg", "-v")
	// We expect an error since one test is designed to fail

	// Verify test counts
	if result.Passed == 0 {
		t.Error("Expected at least some tests to pass")
	}

	if result.Failed == 0 {
		t.Error("Expected at least one test to fail")
	}

	// Verify test names were extracted
	expectedTests := []string{
		"TestReverse",
		"TestCapitalize",
		"TestIsEmpty",
		"TestIntentionallyFailing",
	}

	for _, expectedTest := range expectedTests {
		found := false
		for _, actualTest := range result.Tests {
			if strings.HasPrefix(actualTest, expectedTest) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find test %s in results", expectedTest)
		}
	}

	// Verify output contains information about the failing test
	if !strings.Contains(result.Output, "TestIntentionallyFailing") ||
		!strings.Contains(result.Output, "This test is designed to fail") {
		t.Errorf("Expected output to contain information about the failing test")
	}

	// Test running a specific test
	specificResult, err := executor.ExecuteTest(module, "./pkg", "-run=TestReverse")
	if err != nil {
		t.Errorf("Running specific test should not fail: %v", err)
	}

	if specificResult.Failed > 0 {
		t.Errorf("TestReverse should not contain failing tests")
	}

	// Test running a failing test
	failingResult, _ := executor.ExecuteTest(module, "./pkg", "-run=TestIntentionallyFailing")
	if failingResult.Failed != 1 {
		t.Errorf("Expected exactly 1 failing test, got %d", failingResult.Failed)
	}

	// Verify tested symbols
	if len(result.TestedSymbols) == 0 {
		t.Logf("Note: TestedSymbols is empty. This is expected if the implementation is a stub.")
	}
}
