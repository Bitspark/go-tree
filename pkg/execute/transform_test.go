package execute

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/loader"
	"bitspark.dev/go-tree/pkg/core/module"
)

func TestLoadTransformExecute(t *testing.T) {
	// Step 1: Load the module from testdata
	l := loader.NewGoModuleLoader()
	options := loader.DefaultLoadOptions()
	options.IncludeTests = true

	mod, err := l.LoadWithOptions("../../testdata", options)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	t.Logf("Loaded module: %s with %d packages", mod.Path, len(mod.Packages))

	// Step 2: Add a test file to the samplepackage since it doesn't have one
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Sample package not found in loaded module")
	}

	// Create a test file for the NewUser function
	testFile := module.NewFile("", "functions_test.go", true)
	testFile.SourceCode = `package samplepackage

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	user := NewUser("testuser")
	
	if user.Name != "testuser" {
		t.Errorf("Expected name to be 'testuser', got %q", user.Name)
	}
	
	if user.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got %q", user.Username)
	}
}

func TestUser_Login(t *testing.T) {
	user := NewUser("testuser")
	user.Password = "password123"
	
	// Test successful login
	success, err := user.Login("testuser", "password123")
	if !success || err != nil {
		t.Errorf("Expected successful login, got success=%v, err=%v", success, err)
	}
}
`
	// Add test file to the package
	samplePkg.AddFile(testFile)

	// Step 3: Transform the module - add debug output to all non-test functions
	transformedMod := transformModule(t, mod)

	// Step 4: Execute tests on the transformed module
	executor := NewTmpExecutor()
	// Uncomment to keep generated files for inspection
	// executor.KeepTempFiles = true

	// Run tests specifically for the samplepackage
	result, err := executor.ExecuteTest(transformedMod, "./samplepackage", "-v")
	if err != nil {
		t.Fatalf("Failed to execute tests: %v", err)
	}

	// Verify test results
	t.Logf("Test results: %d tests, %d failures", len(result.Tests), result.Failed)

	if result.Failed > 0 {
		t.Errorf("Expected all tests to pass, got %d failures", result.Failed)
	}

	// Verify the expected tests ran
	expectedTests := []string{
		"TestNewUser",
		"TestUser_Login",
	}

	for _, testName := range expectedTests {
		found := false
		for _, ran := range result.Tests {
			if strings.Contains(ran, testName) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected test %q to run but it was not found in results", testName)
		}
	}

	// Verify transformation effects are visible in output
	if !strings.Contains(result.Output, "[DEBUG] Executing NewUser") {
		t.Errorf("Expected to see debug output from transformed NewUser function")
	}
}

// transformModule adds debug print statements to all functions in the module
func transformModule(t *testing.T, mod *module.Module) *module.Module {
	t.Log("Transforming module...")

	// For each package in the module
	for pkgPath, pkg := range mod.Packages {
		t.Logf("Transforming package: %s", pkgPath)

		// For each function, add a debug print statement at the beginning
		for funcName, fn := range pkg.Functions {
			// Skip methods and test functions
			if fn.IsMethod || strings.HasPrefix(funcName, "Test") {
				continue
			}

			t.Logf("  Transforming function: %s", funcName)

			// Find the file containing this function
			for _, file := range pkg.Files {
				if file.IsTest {
					continue
				}

				// Check if this file contains our function's source code
				if strings.Contains(file.SourceCode, "func "+funcName) {
					// Parse the source code
					lines := strings.Split(file.SourceCode, "\n")

					// Find the function definition line
					startLine := -1
					openBraceIndex := -1

					for i, line := range lines {
						if strings.Contains(line, "func "+funcName) {
							startLine = i
						}

						// Find the opening brace after function signature
						if startLine != -1 && i > startLine && strings.Contains(line, "{") {
							openBraceIndex = i
							break
						}

						// If we find the brace on the same line as the func declaration
						if startLine != -1 && i == startLine && strings.Contains(line, "{") {
							openBraceIndex = i
							break
						}
					}

					// Insert debug statement after the opening brace
					if openBraceIndex != -1 {
						// Find position after the opening brace
						pos := strings.Index(lines[openBraceIndex], "{")

						// If we have a position, insert after the brace
						if pos != -1 {
							indent := strings.Repeat(" ", pos+2) // Indent plus 2 spaces
							debugLine := indent + `fmt.Println("[DEBUG] Executing ` + funcName + `")`

							// Add import if needed
							if !strings.Contains(file.SourceCode, `import "fmt"`) && !strings.Contains(file.SourceCode, `import (`) {
								// Add import at the top, after package declaration
								for i, line := range lines {
									if strings.HasPrefix(line, "package ") {
										lines = append(lines[:i+1], append([]string{"", `import "fmt"`}, lines[i+1:]...)...)
										break
									}
								}
							} else if !strings.Contains(file.SourceCode, `"fmt"`) && strings.Contains(file.SourceCode, `import (`) {
								// Find import block and add fmt
								for i, line := range lines {
									if strings.Contains(line, "import (") {
										// Find the closing parenthesis
										for j := i + 1; j < len(lines); j++ {
											if strings.Contains(lines[j], ")") {
												// Insert before closing parenthesis
												indent := strings.Repeat(" ", strings.Index(lines[i+1], strings.TrimSpace(lines[i+1])))
												lines = append(lines[:j], append([]string{indent + `"fmt"`}, lines[j:]...)...)
												break
											}
										}
										break
									}
								}
							}

							// Insert debug line after opening brace
							parts := strings.SplitN(lines[openBraceIndex], "{", 2)
							if len(parts) == 2 {
								lines[openBraceIndex] = parts[0] + "{" + "\n" + debugLine
								if parts[1] != "" {
									lines[openBraceIndex] += "\n" + indent + parts[1]
								}
							} else {
								// Just add after the line with the brace
								newLines := make([]string, 0, len(lines)+1)
								newLines = append(newLines, lines[:openBraceIndex+1]...)
								newLines = append(newLines, debugLine)
								newLines = append(newLines, lines[openBraceIndex+1:]...)
								lines = newLines
							}

							// Update the source code
							file.SourceCode = strings.Join(lines, "\n")

							t.Logf("    Added debug output to %s", funcName)
						}
					}
				}
			}
		}
	}

	return mod
}
