package saver

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"go/ast"
	"go/token"

	"bitspark.dev/go-tree/pkg/typesys"
)

// Test constants
const testModulePath = "github.com/example/testmodule"

// Helper function to create a simple test module
func createTestModule(t *testing.T) *typesys.Module {
	t.Helper()

	// Create a temporary directory for the module
	tempDir, err := os.MkdirTemp("", "saver-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a module
	module := typesys.NewModule(tempDir)
	module.Path = testModulePath
	module.GoVersion = "1.18"

	return module
}

// Helper function to create a test package with file
func addTestPackage(t *testing.T, module *typesys.Module, pkgName, relPath string) *typesys.Package {
	t.Helper()

	importPath := module.Path
	if relPath != "" {
		importPath = module.Path + "/" + relPath
	}

	// Create package
	pkg := &typesys.Package{
		Module:     module,
		Name:       pkgName,
		ImportPath: importPath,
		Files:      make(map[string]*typesys.File),
	}

	module.Packages[importPath] = pkg
	return pkg
}

// Helper function to add a file to a package
func addTestFile(t *testing.T, pkg *typesys.Package, fileName string) *typesys.File {
	t.Helper()

	filePath := filepath.Join(pkg.Module.Dir, filepath.Base(pkg.ImportPath), fileName)

	// Create file
	file := &typesys.File{
		Path:    filePath,
		Name:    fileName,
		Package: pkg,
		Symbols: make([]*typesys.Symbol, 0),
	}

	pkg.Files[filePath] = file
	return file
}

// Helper function to add a function symbol to a file
func addFunctionSymbol(t *testing.T, file *typesys.File, name string) *typesys.Symbol {
	t.Helper()

	symbol := &typesys.Symbol{
		ID:       name + "ID",
		Name:     name,
		Kind:     typesys.KindFunction,
		Exported: len(name) > 0 && unicode.IsUpper(rune(name[0])), // Exported if starts with uppercase
		Package:  file.Package,
		File:     file,
	}

	file.Symbols = append(file.Symbols, symbol)
	return symbol
}

// Helper function to add a type symbol to a file
func addTypeSymbol(t *testing.T, file *typesys.File, name string) *typesys.Symbol {
	t.Helper()

	symbol := &typesys.Symbol{
		ID:       name + "ID",
		Name:     name,
		Kind:     typesys.KindType,
		Exported: len(name) > 0 && unicode.IsUpper(rune(name[0])), // Exported if starts with uppercase
		Package:  file.Package,
		File:     file,
	}

	file.Symbols = append(file.Symbols, symbol)
	return symbol
}

// Test SaveOptions Default values
func TestDefaultSaveOptions(t *testing.T) {
	options := DefaultSaveOptions()

	if !options.Format {
		t.Error("Default Format should be true")
	}

	if !options.OrganizeImports {
		t.Error("Default OrganizeImports should be true")
	}

	if options.ASTMode != SmartMerge {
		t.Errorf("Default ASTMode should be SmartMerge, got %v", options.ASTMode)
	}
}

// Test GoModuleSaver creation
func TestNewGoModuleSaver(t *testing.T) {
	saver := NewGoModuleSaver()

	if saver == nil {
		t.Fatal("NewGoModuleSaver returned nil")
	}

	if saver.generator == nil {
		t.Error("Saver should have a generator")
	}
}

// Test simple module saving
func TestGoModuleSaver_SaveTo(t *testing.T) {
	// Create a test module
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	// Add a package
	pkg := addTestPackage(t, module, "main", "")

	// Add a file
	file := addTestFile(t, pkg, "main.go")

	// Add symbols
	addFunctionSymbol(t, file, "main")
	addTypeSymbol(t, file, "Config")

	// Create output directory
	outDir, err := os.MkdirTemp("", "saver-output-*")
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Logf("Failed to remove output directory: %v", err)
		}
	})

	// Create saver
	saver := NewGoModuleSaver()

	// Save the module
	err = saver.SaveTo(module, outDir)
	if err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	// Check that go.mod was created
	goModPath := filepath.Join(outDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("go.mod file was not created")
	}

	// Check that main.go was created
	mainGoPath := filepath.Join(outDir, "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("main.go file was not created")
	}

	// Read the content of main.go
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	// Check that the content contains expected elements
	contentStr := string(content)
	if !strings.Contains(contentStr, "package main") {
		t.Error("main.go does not contain 'package main'")
	}

	if !strings.Contains(contentStr, "func main") {
		t.Error("main.go does not contain 'func main'")
	}

	if !strings.Contains(contentStr, "type Config") {
		t.Error("main.go does not contain 'type Config'")
	}
}

// Test DefaultFileContentGenerator
func TestDefaultFileContentGenerator_GenerateFileContent(t *testing.T) {
	// Create a test module and package
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	pkg := addTestPackage(t, module, "example", "pkg")
	file := addTestFile(t, pkg, "example.go")

	// Add a function and type
	addFunctionSymbol(t, file, "ExampleFunc")
	addTypeSymbol(t, file, "ExampleType")

	// Create generator
	generator := NewDefaultFileContentGenerator()

	// Generate content
	content, err := generator.GenerateFileContent(file, DefaultSaveOptions())
	if err != nil {
		t.Fatalf("GenerateFileContent failed: %v", err)
	}

	// Check content
	contentStr := string(content)
	if !strings.Contains(contentStr, "package example") {
		t.Error("Content does not contain 'package example'")
	}

	if !strings.Contains(contentStr, "func ExampleFunc") {
		t.Error("Content does not contain 'func ExampleFunc'")
	}

	if !strings.Contains(contentStr, "type ExampleType") {
		t.Error("Content does not contain 'type ExampleType'")
	}
}

// Test Symbol Writers
func TestSymbolWriters(t *testing.T) {
	// Test scenarios for each writer
	tests := []struct {
		name     string
		kind     typesys.SymbolKind
		writer   SymbolWriter
		expected string
	}{
		{
			name:     "FunctionWriter",
			kind:     typesys.KindFunction,
			writer:   &FunctionWriter{},
			expected: "func TestFunc",
		},
		{
			name:     "TypeWriter",
			kind:     typesys.KindType,
			writer:   &TypeWriter{},
			expected: "type TestType",
		},
		{
			name:     "VarWriter",
			kind:     typesys.KindVariable,
			writer:   &VarWriter{},
			expected: "var TestVar",
		},
		{
			name:     "ConstWriter",
			kind:     typesys.KindConstant,
			writer:   &ConstWriter{},
			expected: "const TestConst",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a symbol with appropriate kind and name
			symbolName := "Test" + strings.TrimSuffix(strings.TrimPrefix(tc.name, ""), "Writer")

			symbol := &typesys.Symbol{
				Name: symbolName,
				Kind: tc.kind,
			}

			// Create buffer and write symbol
			var buf bytes.Buffer
			err := tc.writer.WriteSymbol(symbol, &buf)

			// Check result
			if err != nil {
				t.Fatalf("WriteSymbol failed: %v", err)
			}

			result := buf.String()
			if !strings.Contains(result, tc.expected) {
				t.Errorf("Expected result to contain '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// Test ModificationTracker
func TestModificationTracker(t *testing.T) {
	// Create a test module structure
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	pkg := addTestPackage(t, module, "tracker", "")
	file := addTestFile(t, pkg, "tracker.go")
	sym := addFunctionSymbol(t, file, "TestFunc")

	// Create tracker
	tracker := NewDefaultModificationTracker()

	// Check that nothing is modified initially
	if tracker.IsModified(sym) {
		t.Error("Symbol should not be modified initially")
	}

	if tracker.IsModified(file) {
		t.Error("File should not be modified initially")
	}

	// Mark symbol as modified
	tracker.MarkModified(sym)

	// Check that symbol and containing elements are marked
	if !tracker.IsModified(sym) {
		t.Error("Symbol should be marked as modified")
	}

	if !tracker.IsModified(file) {
		t.Error("File should be marked as modified when symbol is modified")
	}

	if !tracker.IsModified(pkg) {
		t.Error("Package should be marked as modified when symbol is modified")
	}

	// Clear modification
	tracker.ClearModified(sym)

	if tracker.IsModified(sym) {
		t.Error("Symbol should not be modified after clearing")
	}

	// Parent elements should still be marked
	if !tracker.IsModified(file) {
		t.Error("File should still be marked as modified")
	}

	// Clear all
	tracker.ClearAll()

	if tracker.IsModified(file) {
		t.Error("File should not be modified after ClearAll")
	}

	if tracker.IsModified(pkg) {
		t.Error("Package should not be modified after ClearAll")
	}
}

// Test relativePath function
func TestRelativePath(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		modPath    string
		expected   string
	}{
		{
			name:       "Empty module path",
			importPath: "github.com/example/pkg",
			modPath:    "",
			expected:   "github.com/example/pkg",
		},
		{
			name:       "Root package",
			importPath: "github.com/example/pkg",
			modPath:    "github.com/example/pkg",
			expected:   "",
		},
		{
			name:       "Subpackage",
			importPath: "github.com/example/pkg/subpkg",
			modPath:    "github.com/example/pkg",
			expected:   "subpkg",
		},
		{
			name:       "Nested subpackage",
			importPath: "github.com/example/pkg/subpkg/nested",
			modPath:    "github.com/example/pkg",
			expected:   "subpkg/nested",
		},
		{
			name:       "Unrelated package",
			importPath: "github.com/example/other",
			modPath:    "github.com/example/pkg",
			expected:   "github.com/example/other",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := relativePath(tc.importPath, tc.modPath)
			if result != tc.expected {
				t.Errorf("relativePath(%s, %s) = %s, expected %s",
					tc.importPath, tc.modPath, result, tc.expected)
			}
		})
	}
}

// Test ModificationsAnalyzer
func TestModificationsAnalyzer(t *testing.T) {
	// Create a test module
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	// Add two packages
	pkg1 := addTestPackage(t, module, "pkg1", "pkg1")
	pkg2 := addTestPackage(t, module, "pkg2", "pkg2")

	// Add files to packages
	file1 := addTestFile(t, pkg1, "file1.go")
	file2 := addTestFile(t, pkg1, "file2.go")
	file3 := addTestFile(t, pkg2, "file3.go")

	// Add symbols to files
	sym1 := addFunctionSymbol(t, file1, "Function1")
	sym2 := addTypeSymbol(t, file1, "Type1")
	sym3 := addFunctionSymbol(t, file2, "Function2")
	sym4 := addTypeSymbol(t, file3, "Type2")

	// Create tracker and analyzer
	tracker := NewDefaultModificationTracker()
	analyzer := NewModificationsAnalyzer(tracker)

	// Initially, nothing should be modified
	modFiles := analyzer.GetModifiedFiles(module)
	if len(modFiles) != 0 {
		t.Errorf("Expected 0 modified files initially, got %d", len(modFiles))
	}

	// Mark a symbol as modified
	tracker.MarkModified(sym1)

	// Check that file1 is now modified
	modFiles = analyzer.GetModifiedFiles(module)
	if len(modFiles) != 1 || modFiles[0] != file1 {
		t.Errorf("Expected only file1 to be modified, got %v", modFiles)
	}

	// Check modified symbols in file1
	modSymbols := analyzer.GetModifiedSymbols(file1)
	if len(modSymbols) != 1 || modSymbols[0] != sym1 {
		t.Errorf("Expected only sym1 to be modified, got %v", modSymbols)
	}

	// Mark another symbol in the same file
	tracker.MarkModified(sym2)

	// Check that we still have only one modified file
	modFiles = analyzer.GetModifiedFiles(module)
	if len(modFiles) != 1 {
		t.Errorf("Expected 1 modified file, got %d", len(modFiles))
	}

	// Check that we now have two modified symbols in file1
	modSymbols = analyzer.GetModifiedSymbols(file1)
	if len(modSymbols) != 2 {
		t.Errorf("Expected 2 modified symbols in file1, got %d", len(modSymbols))
	}

	// Mark a symbol in another file
	tracker.MarkModified(sym3)

	// Check that we now have two modified files
	modFiles = analyzer.GetModifiedFiles(module)
	if len(modFiles) != 2 {
		t.Errorf("Expected 2 modified files, got %d", len(modFiles))
	}

	// Check that sym4 is not modified
	if tracker.IsModified(sym4) {
		t.Errorf("Expected sym4 to not be modified")
	}

	// Mark the file directly
	tracker.MarkModified(file3)

	// Check that we now have three modified files
	modFiles = analyzer.GetModifiedFiles(module)
	if len(modFiles) != 3 {
		t.Errorf("Expected 3 modified files, got %d", len(modFiles))
	}

	// Clear all modifications
	tracker.ClearAll()

	// Check that no files are modified now
	modFiles = analyzer.GetModifiedFiles(module)
	if len(modFiles) != 0 {
		t.Errorf("Expected 0 modified files after clearing, got %d", len(modFiles))
	}
}

// Test helper functions in symbolgen.go
func TestSymbolGenHelpers(t *testing.T) {
	// Test writeDocComment
	t.Run("writeDocComment", func(t *testing.T) {
		tests := []struct {
			name     string
			doc      string
			expected string
		}{
			{
				name:     "Single line comment",
				doc:      "This is a comment",
				expected: "// This is a comment\n",
			},
			{
				name:     "Multi-line comment",
				doc:      "Line 1\nLine 2\nLine 3",
				expected: "// Line 1\n// Line 2\n// Line 3\n",
			},
			{
				name:     "Empty comment",
				doc:      "",
				expected: "// \n",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var buf bytes.Buffer
				writeDocComment(tc.doc, &buf)
				result := buf.String()
				if result != tc.expected {
					t.Errorf("writeDocComment(%s) = %q, expected %q", tc.doc, result, tc.expected)
				}
			})
		}
	})

	// Test indentCode
	t.Run("indentCode", func(t *testing.T) {
		tests := []struct {
			name     string
			code     string
			indent   string
			expected string
		}{
			{
				name:     "Single line with tab",
				code:     "func main() {}",
				indent:   "\t",
				expected: "\tfunc main() {}",
			},
			{
				name:     "Multi-line with tab",
				code:     "func main() {\n    fmt.Println(\"Hello\")\n}",
				indent:   "\t",
				expected: "\tfunc main() {\n\t    fmt.Println(\"Hello\")\n\t}",
			},
			{
				name:     "With spaces",
				code:     "func main() {\nfmt.Println(\"Hello\")\n}",
				indent:   "  ",
				expected: "  func main() {\n  fmt.Println(\"Hello\")\n  }",
			},
			{
				name:     "Empty lines",
				code:     "func main() {\n\nfmt.Println(\"Hello\")\n\n}",
				indent:   "\t",
				expected: "\tfunc main() {\n\n\tfmt.Println(\"Hello\")\n\n\t}",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				result := indentCode(tc.code, tc.indent)
				if result != tc.expected {
					t.Errorf("indentCode(%q, %q) = %q, expected %q", tc.code, tc.indent, result, tc.expected)
				}
			})
		}
	})
}

// Test savePackage function
func TestSavePackage(t *testing.T) {
	// Create a test module
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	// Add a package
	pkg := addTestPackage(t, module, "testpkg", "testpkg")

	// Add a file to the package
	file := addTestFile(t, pkg, "testfile.go")

	// Add symbols
	addFunctionSymbol(t, file, "TestFunc")
	addTypeSymbol(t, file, "TestType")

	// Create output directory
	outDir, err := os.MkdirTemp("", "saver-pkg-test-*")
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Logf("Failed to remove output directory: %v", err)
		}
	})

	// Create saver
	saver := NewGoModuleSaver()

	// Save the package
	err = saver.savePackage(pkg, outDir, pkg.ImportPath, module.Path, DefaultSaveOptions())
	if err != nil {
		t.Fatalf("savePackage failed: %v", err)
	}

	// Check that file was created in the right place
	expectedFilePath := filepath.Join(outDir, "testpkg", "testfile.go")
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", expectedFilePath)
	}

	// Read the content to verify
	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check content basics
	contentStr := string(content)
	if !strings.Contains(contentStr, "package testpkg") {
		t.Error("File content doesn't contain package declaration")
	}
	if !strings.Contains(contentStr, "func TestFunc") {
		t.Error("File content doesn't contain function")
	}
	if !strings.Contains(contentStr, "type TestType") {
		t.Error("File content doesn't contain type")
	}
}

// Test different ASTReconstructionMode values
func TestASTReconstructionModes(t *testing.T) {
	// Create save options with different modes
	modes := []struct {
		name string
		mode ASTReconstructionMode
	}{
		{"PreserveOriginal", PreserveOriginal},
		{"ReformatAll", ReformatAll},
		{"SmartMerge", SmartMerge},
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			options := DefaultSaveOptions()
			options.ASTMode = m.mode

			if options.ASTMode != m.mode {
				t.Errorf("Expected ASTMode to be %v, got %v", m.mode, options.ASTMode)
			}
		})
	}
}

// Test saveGoMod function
func TestSaveGoMod(t *testing.T) {
	// Create a test module
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	// Create output directory
	outDir, err := os.MkdirTemp("", "saver-gomod-test-*")
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Logf("Failed to remove output directory: %v", err)
		}
	})

	// Create saver
	saver := NewGoModuleSaver()

	// Save the go.mod file
	err = saver.saveGoMod(module, outDir)
	if err != nil {
		t.Fatalf("saveGoMod failed: %v", err)
	}

	// Check that go.mod was created
	goModPath := filepath.Join(outDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("go.mod file was not created")
	}

	// Read the content to verify
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Check content basics
	contentStr := string(content)
	expectedContent := fmt.Sprintf("module %s\n\ngo %s\n", module.Path, module.GoVersion)
	if contentStr != expectedContent {
		t.Errorf("go.mod content doesn't match expected.\nGot: %q\nExpected: %q", contentStr, expectedContent)
	}
}

// Test SaveWithOptions and SaveToWithOptions
func TestSaveWithOptions(t *testing.T) {
	// Create a test module
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	// Add a package
	pkg := addTestPackage(t, module, "main", "")
	file := addTestFile(t, pkg, "main.go")
	addFunctionSymbol(t, file, "main")

	// Create custom options
	options := DefaultSaveOptions()
	options.CreateBackups = true
	options.Format = false

	// Create saver
	saver := NewGoModuleSaver()

	// Create output directory
	outDir, err := os.MkdirTemp("", "saver-options-test-*")
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Logf("Failed to remove output directory: %v", err)
		}
	})

	// Test SaveToWithOptions
	err = saver.SaveToWithOptions(module, outDir, options)
	if err != nil {
		t.Fatalf("SaveToWithOptions failed: %v", err)
	}

	// Check that files were created
	goModPath := filepath.Join(outDir, "go.mod")
	mainGoPath := filepath.Join(outDir, "main.go")

	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("go.mod file was not created")
	}

	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("main.go file was not created")
	}

	// Now test SaveWithOptions to same directory
	module.Dir = outDir // Set the module dir to our output dir

	// First modify the main.go file to have some content
	err = os.WriteFile(mainGoPath, []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write to main.go: %v", err)
	}

	// Now save the module which should create a backup
	err = saver.SaveWithOptions(module, options)
	if err != nil {
		t.Fatalf("SaveWithOptions failed: %v", err)
	}

	// Check that backup was created
	backupPath := mainGoPath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
}

// Test error cases for saver functions
func TestSaverErrorCases(t *testing.T) {
	// Create saver
	saver := NewGoModuleSaver()

	// Try to save a nil module
	err := saver.SaveToWithOptions(nil, "some/dir", DefaultSaveOptions())
	if err == nil {
		t.Error("Expected error when saving nil module, got nil")
	}

	// Create a module with empty Dir
	module := typesys.NewModule("")

	// Try to save a module with empty Dir
	err = saver.SaveWithOptions(module, DefaultSaveOptions())
	if err == nil {
		t.Error("Expected error when saving module with empty Dir, got nil")
	}
}

// Test FileFilter
func TestGoModuleSaverFileFilter(t *testing.T) {
	// Create a test module
	module := createTestModule(t)
	t.Cleanup(func() {
		if err := os.RemoveAll(module.Dir); err != nil {
			t.Logf("Failed to remove module directory: %v", err)
		}
	})

	// Add a package with two files
	pkg := addTestPackage(t, module, "main", "")
	file1 := addTestFile(t, pkg, "main.go")
	file2 := addTestFile(t, pkg, "helper.go")
	addFunctionSymbol(t, file1, "main")
	addFunctionSymbol(t, file2, "helper")

	// Create output directory
	outDir, err := os.MkdirTemp("", "saver-filter-test-*")
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Logf("Failed to remove output directory: %v", err)
		}
	})

	// Create saver with filter that only includes main.go
	saver := NewGoModuleSaver()
	saver.FileFilter = func(file *typesys.File) bool {
		return file.Name == "main.go"
	}

	// Save the module
	err = saver.SaveTo(module, outDir)
	if err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	// Check that main.go was created but helper.go was not
	mainGoPath := filepath.Join(outDir, "main.go")
	helperGoPath := filepath.Join(outDir, "helper.go")

	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("main.go file was not created")
	}

	if _, err := os.Stat(helperGoPath); !os.IsNotExist(err) {
		t.Error("helper.go file was created despite filter")
	}
}

// Test WriteTo function
func TestWriteTo(t *testing.T) {
	// Test successful writing
	t.Run("successful write", func(t *testing.T) {
		content := []byte("test content")
		var buf bytes.Buffer

		err := WriteTo(content, &buf)
		if err != nil {
			t.Errorf("WriteTo should not return error: %v", err)
		}

		if buf.String() != "test content" {
			t.Errorf("WriteTo did not write correct content. Got %q, expected %q", buf.String(), "test content")
		}
	})

	// Test error handling with a failing writer
	t.Run("error handling", func(t *testing.T) {
		content := []byte("test content")
		w := &errorWriter{}

		err := WriteTo(content, w)
		if err == nil {
			t.Errorf("WriteTo should return error with failing writer")
		}
	})
}

// MockWriter that always fails on Write
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write error")
}

// Test ASTGenerator
func TestASTGenerator(t *testing.T) {
	// We need to create minimal AST to test the generator
	fset := token.NewFileSet()
	astFile := &ast.File{
		Name: &ast.Ident{Name: "main"},
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					&ast.ImportSpec{
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"fmt\"",
						},
					},
				},
			},
			&ast.FuncDecl{
				Name: &ast.Ident{Name: "main"},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "fmt"},
									Sel: &ast.Ident{Name: "Println"},
								},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"Hello, World!\"",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create options with different configurations
	tests := []struct {
		name    string
		options SaveOptions
	}{
		{
			name: "gofmt enabled",
			options: SaveOptions{
				Gofmt:    true,
				UseTabs:  true,
				TabWidth: 8,
			},
		},
		{
			name: "custom formatting",
			options: SaveOptions{
				Gofmt:    false,
				UseTabs:  false,
				TabWidth: 4,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			generator := NewASTGenerator(tc.options)

			content, err := generator.GenerateFromAST(astFile, fset)
			if err != nil {
				t.Fatalf("GenerateFromAST failed: %v", err)
			}

			if len(content) == 0 {
				t.Error("GenerateFromAST returned empty content")
			}

			// Check basic content properties
			contentStr := string(content)
			if !strings.Contains(contentStr, "package main") {
				t.Error("Content doesn't contain package declaration")
			}
			if !strings.Contains(contentStr, "import") {
				t.Error("Content doesn't contain import")
			}
			if !strings.Contains(contentStr, "fmt") {
				t.Error("Content doesn't contain imported package")
			}
			if !strings.Contains(contentStr, "func main") {
				t.Error("Content doesn't contain main function")
			}
		})
	}

	// Test error cases
	t.Run("nil inputs", func(t *testing.T) {
		generator := NewASTGenerator(DefaultSaveOptions())

		// Test with nil AST
		_, err := generator.GenerateFromAST(nil, fset)
		if err == nil {
			t.Error("GenerateFromAST should return error with nil AST")
		}

		// Test with nil FileSet
		_, err = generator.GenerateFromAST(astFile, nil)
		if err == nil {
			t.Error("GenerateFromAST should return error with nil FileSet")
		}
	})
}

// Test GenerateSourceFile function
func TestGenerateSourceFile(t *testing.T) {
	// Create a basic file with AST
	file := &typesys.File{
		Name: "test.go",
		Package: &typesys.Package{
			Name: "test",
		},
		AST:     &ast.File{Name: &ast.Ident{Name: "test"}},
		FileSet: token.NewFileSet(),
	}

	// Test with different ASTMode options
	t.Run("preserve original", func(t *testing.T) {
		options := DefaultSaveOptions()
		options.ASTMode = PreserveOriginal

		_, err := GenerateSourceFile(file, options)
		if err != nil {
			t.Errorf("GenerateSourceFile with PreserveOriginal should not fail: %v", err)
		}
	})

	// Test error cases
	t.Run("missing AST", func(t *testing.T) {
		fileWithoutAST := &typesys.File{
			Name:    "test.go",
			Package: &typesys.Package{Name: "test"},
			// No AST or FileSet
		}

		_, err := GenerateSourceFile(fileWithoutAST, DefaultSaveOptions())
		if err == nil {
			t.Error("GenerateSourceFile should fail with missing AST")
		}
	})
}

// Test DefaultFileContentGenerator's error cases
func TestGenerateFileContentErrors(t *testing.T) {
	generator := NewDefaultFileContentGenerator()

	// Test with nil file
	_, err := generator.GenerateFileContent(nil, DefaultSaveOptions())
	if err == nil {
		t.Error("GenerateFileContent should return error with nil file")
	}

	// Test with missing symbol writer
	customGenerator := &DefaultFileContentGenerator{
		symbolWriters: make(map[typesys.SymbolKind]SymbolWriter),
	}
	file := &typesys.File{
		Name:    "test.go",
		Package: &typesys.Package{Name: "test"},
		Symbols: []*typesys.Symbol{
			{
				Name: "TestFunc",
				Kind: typesys.KindFunction,
			},
		},
	}

	_, err = customGenerator.generateFromSymbols(file, DefaultSaveOptions())
	// This shouldn't return an error, it should just skip the symbol
	if err != nil {
		t.Errorf("generateFromSymbols should not return error with missing symbol writer: %v", err)
	}
}

// Test symbol writers error cases
func TestSymbolWritersErrors(t *testing.T) {
	writers := []struct {
		name   string
		writer SymbolWriter
		kind   typesys.SymbolKind
	}{
		{"FunctionWriter", &FunctionWriter{}, typesys.KindFunction},
		{"TypeWriter", &TypeWriter{}, typesys.KindType},
		{"VarWriter", &VarWriter{}, typesys.KindVariable},
		{"ConstWriter", &ConstWriter{}, typesys.KindConstant},
	}

	for _, w := range writers {
		t.Run(w.name+" nil symbol", func(t *testing.T) {
			var buf bytes.Buffer
			err := w.writer.WriteSymbol(nil, &buf)
			if err == nil {
				t.Errorf("%s.WriteSymbol should return error with nil symbol", w.name)
			}
		})

		t.Run(w.name+" wrong kind", func(t *testing.T) {
			// Create symbol with wrong kind
			wrongKind := typesys.KindConstant
			if w.kind == typesys.KindConstant {
				wrongKind = typesys.KindFunction
			}

			sym := &typesys.Symbol{
				Name: "TestSymbol",
				Kind: wrongKind,
			}

			var buf bytes.Buffer
			err := w.writer.WriteSymbol(sym, &buf)
			if err == nil {
				t.Errorf("%s.WriteSymbol should return error with wrong symbol kind", w.name)
			}
		})
	}
}
