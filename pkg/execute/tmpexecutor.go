package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/core/module"
)

// TmpExecutor is an executor that saves in-memory modules to a temporary
// directory before executing them with the Go toolchain.
type TmpExecutor struct {
	// Underlying executor to use after saving to temp directory
	executor ModuleExecutor

	// TempBaseDir is the base directory for creating temporary module directories
	// If empty, os.TempDir() will be used
	TempBaseDir string

	// KeepTempFiles determines whether temporary files are kept after execution
	KeepTempFiles bool
}

// NewTmpExecutor creates a new temporary directory executor
func NewTmpExecutor() *TmpExecutor {
	return &TmpExecutor{
		executor:      NewGoExecutor(),
		KeepTempFiles: false,
	}
}

// Execute runs a command on a module by first saving it to a temporary directory
func (e *TmpExecutor) Execute(mod *module.Module, args ...string) (ExecutionResult, error) {
	// Create temporary directory
	tempDir, err := e.createTempDir(mod)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clean up temporary directory unless KeepTempFiles is true
	if !e.KeepTempFiles {
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory %s: %v\n", tempDir, err)
			}
		}()
	}

	// Save module to temporary directory
	tmpModule, err := e.saveToTemp(mod, tempDir)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to save module to temp directory: %w", err)
	}

	// Set working directory explicitly
	if goExec, ok := e.executor.(*GoExecutor); ok {
		goExec.WorkingDir = tempDir
	}

	// Execute using the underlying executor
	return e.executor.Execute(tmpModule, args...)
}

// ExecuteTest runs tests in a module by first saving it to a temporary directory
func (e *TmpExecutor) ExecuteTest(mod *module.Module, pkgPath string, testFlags ...string) (TestResult, error) {
	// Create temporary directory
	tempDir, err := e.createTempDir(mod)
	if err != nil {
		return TestResult{}, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clean up temporary directory unless KeepTempFiles is true
	if !e.KeepTempFiles {
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory %s: %v\n", tempDir, err)
			}
		}()
	}

	// Save module to temporary directory
	tmpModule, err := e.saveToTemp(mod, tempDir)
	if err != nil {
		return TestResult{}, fmt.Errorf("failed to save module to temp directory: %w", err)
	}

	// Explicitly set working directory in the executor
	if goExec, ok := e.executor.(*GoExecutor); ok {
		goExec.WorkingDir = tempDir
	}

	// Execute test using the underlying executor
	return e.executor.ExecuteTest(tmpModule, pkgPath, testFlags...)
}

// ExecuteFunc calls a specific function in the module after saving to a temp directory
func (e *TmpExecutor) ExecuteFunc(mod *module.Module, funcPath string, args ...interface{}) (interface{}, error) {
	// Create temporary directory
	tempDir, err := e.createTempDir(mod)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clean up temporary directory unless KeepTempFiles is true
	if !e.KeepTempFiles {
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory %s: %v\n", tempDir, err)
			}
		}()
	}

	// Save module to temporary directory
	tmpModule, err := e.saveToTemp(mod, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to save module to temp directory: %w", err)
	}

	// Explicitly set working directory in the executor
	if goExec, ok := e.executor.(*GoExecutor); ok {
		goExec.WorkingDir = tempDir
	}

	// Execute function using the underlying executor
	return e.executor.ExecuteFunc(tmpModule, funcPath, args...)
}

// Helper methods

// createTempDir creates a temporary directory for the module
func (e *TmpExecutor) createTempDir(mod *module.Module) (string, error) {
	baseDir := e.TempBaseDir
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// Create a unique module directory name based on the module path
	moduleNameSafe := filepath.Base(mod.Path)
	tempDir, err := os.MkdirTemp(baseDir, fmt.Sprintf("gotree-%s-", moduleNameSafe))
	if err != nil {
		return "", err
	}

	return tempDir, nil
}

// saveToTemp saves the module to the temporary directory and returns a new Module
// instance that points to the temporary location
func (e *TmpExecutor) saveToTemp(mod *module.Module, tempDir string) (*module.Module, error) {
	// First, ensure the go.mod file is created correctly
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := fmt.Sprintf("module %s\n\ngo %s\n", mod.Path, mod.GoVersion)

	err := os.WriteFile(goModPath, []byte(goModContent), 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Create directories and files for each package
	for importPath, pkg := range mod.Packages {
		if importPath == mod.Path {
			// Skip the root package, we already created go.mod
			continue
		}

		// Create package directory
		relPath := relativePath(importPath, mod.Path)
		pkgDir := filepath.Join(tempDir, relPath)

		if err := os.MkdirAll(pkgDir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create package directory %s: %w", pkgDir, err)
		}

		// Write each file
		for _, file := range pkg.Files {
			filePath := filepath.Join(pkgDir, file.Name)

			if err := os.WriteFile(filePath, []byte(file.SourceCode), 0600); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", filePath, err)
			}
		}
	}

	// Create a new module instance with updated paths
	tmpModule := module.NewModule(mod.Path, tempDir)
	tmpModule.Version = mod.Version
	tmpModule.GoVersion = mod.GoVersion
	tmpModule.Dependencies = mod.Dependencies
	tmpModule.Replace = mod.Replace
	tmpModule.BuildFlags = mod.BuildFlags
	tmpModule.BuildTags = mod.BuildTags
	tmpModule.GoMod = goModPath

	// Create new package references that point to the temp directory
	for importPath, pkg := range mod.Packages {
		// Skip the root package
		if importPath == mod.Path {
			continue
		}

		relPath := relativePath(importPath, mod.Path)
		pkgDir := filepath.Join(tempDir, relPath)

		// Create new package in the temp directory
		tmpPkg := module.NewPackage(pkg.Name, importPath, pkgDir)
		tmpModule.AddPackage(tmpPkg)

		// Create files with proper paths
		for _, file := range pkg.Files {
			tmpFile := module.NewFile(
				filepath.Join(pkgDir, file.Name),
				file.Name,
				file.IsTest,
			)
			tmpPkg.AddFile(tmpFile)
		}
	}

	return tmpModule, nil
}

// relativePath returns a path relative to the module path
// For example, if importPath is "github.com/user/repo/pkg" and modPath is "github.com/user/repo",
// it returns "pkg"
func relativePath(importPath, modPath string) string {
	// If the import path doesn't start with the module path, return it as is
	if !strings.HasPrefix(importPath, modPath) {
		return importPath
	}

	// Get the relative path
	relPath := strings.TrimPrefix(importPath, modPath)

	// Remove leading slash if present
	relPath = strings.TrimPrefix(relPath, "/")

	// If empty (root package), return empty string
	if relPath == "" {
		return ""
	}

	return relPath
}
