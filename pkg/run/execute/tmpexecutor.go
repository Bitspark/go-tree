package execute

import (
	saver2 "bitspark.dev/go-tree/pkg/io/saver"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
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
func (e *TmpExecutor) Execute(mod *typesys.Module, args ...string) (ExecutionResult, error) {
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
func (e *TmpExecutor) ExecuteTest(mod *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error) {
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

// ExecuteFunc calls a specific function in the module with type checking after saving to a temp directory
func (e *TmpExecutor) ExecuteFunc(mod *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
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

	// Find the equivalent function symbol in the saved module
	var savedFuncSymbol *typesys.Symbol
	if pkg := findPackage(tmpModule, funcSymbol.Package.ImportPath); pkg != nil {
		// Look for the function in the saved package
		for _, file := range pkg.Files {
			for _, sym := range file.Symbols {
				if sym.Kind == typesys.KindFunction && sym.Name == funcSymbol.Name {
					savedFuncSymbol = sym
					break
				}
			}
			if savedFuncSymbol != nil {
				break
			}
		}
	}

	if savedFuncSymbol == nil {
		return nil, fmt.Errorf("could not find function %s in saved module", funcSymbol.Name)
	}

	// Execute function using the underlying executor
	return e.executor.ExecuteFunc(tmpModule, savedFuncSymbol, args...)
}

// Helper methods

// createTempDir creates a temporary directory for the module
func (e *TmpExecutor) createTempDir(mod *typesys.Module) (string, error) {
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
func (e *TmpExecutor) saveToTemp(mod *typesys.Module, tempDir string) (*typesys.Module, error) {
	// Use the saver package to write the entire module
	moduleSaver := saver2.NewGoModuleSaver()

	// Configure options for temporary directory use
	options := saver2.DefaultSaveOptions()
	options.CreateBackups = false // No backups in temp dir

	// Save the entire module to the temporary directory
	if err := moduleSaver.SaveToWithOptions(mod, tempDir, options); err != nil {
		return nil, fmt.Errorf("failed to save module to temp directory: %w", err)
	}

	// Create a new module reference that points to the saved location
	tmpModule := typesys.NewModule(tempDir)
	tmpModule.Path = mod.Path
	tmpModule.GoVersion = mod.GoVersion

	// Recreate the package structure
	for importPath, pkg := range mod.Packages {
		// Skip the root package if needed
		if importPath == mod.Path {
			continue
		}

		// Calculate relative path for the package
		relPath := relativePath(importPath, mod.Path)
		pkgDir := filepath.Join(tempDir, relPath)

		// Create a package in the temp module with the same metadata
		tmpPkg := &typesys.Package{
			Module:     tmpModule,
			Name:       pkg.Name,
			ImportPath: importPath,
			Files:      make(map[string]*typesys.File),
		}
		tmpModule.Packages[importPath] = tmpPkg

		// Link each file saved by the saver to the temporary module's structure
		// We need to do this to maintain the right references for later operations
		for filePath, file := range pkg.Files {
			fileName := filepath.Base(filePath)
			newFilePath := filepath.Join(pkgDir, fileName)

			// Create a file reference in the temp module
			tmpFile := &typesys.File{
				Path:    newFilePath,
				Name:    fileName,
				Package: tmpPkg,
				Symbols: make([]*typesys.Symbol, 0),
			}
			tmpPkg.Files[newFilePath] = tmpFile

			// Copy symbols with updated references
			for _, symbol := range file.Symbols {
				tmpSymbol := &typesys.Symbol{
					ID:       symbol.ID,
					Name:     symbol.Name,
					Kind:     symbol.Kind,
					Exported: symbol.Exported,
					Package:  tmpPkg,
					File:     tmpFile,
					Pos:      symbol.Pos,
					End:      symbol.End,
				}
				tmpFile.Symbols = append(tmpFile.Symbols, tmpSymbol)
			}
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
