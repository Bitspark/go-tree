package loader

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/typesys"

	"golang.org/x/tools/go/packages"
)

// loadPackages loads all Go packages in the module directory.
func loadPackages(module *typesys.Module, opts *typesys.LoadOptions) error {
	// Configuration for package loading
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax,
		Dir:        module.Dir,
		Tests:      opts.IncludeTests,
		Fset:       module.FileSet,
		ParseFile:  nil, // Use default parser
		BuildFlags: []string{},
	}

	// Determine the package pattern
	pattern := "./..." // Simple recursive pattern

	tracef(opts, "Loading packages from directory: %s with pattern %s\n", module.Dir, pattern)

	// Load packages
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	tracef(opts, "Loaded %d packages\n", len(pkgs))

	// Debug any package errors
	var pkgsWithErrors int
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			pkgsWithErrors++
			tracef(opts, "Package %s has %d errors:\n", pkg.PkgPath, len(pkg.Errors))
			for _, err := range pkg.Errors {
				tracef(opts, "  - %v\n", err)
			}
		}
	}

	if pkgsWithErrors > 0 {
		tracef(opts, "%d packages had errors\n", pkgsWithErrors)
	}

	// Process loaded packages
	processedPkgs := 0
	for _, pkg := range pkgs {
		// Skip packages with errors
		if len(pkg.Errors) > 0 {
			continue
		}

		// Process the package
		if err := processPackage(module, pkg, opts); err != nil {
			errorf(opts, "Error processing package %s: %v\n", pkg.PkgPath, err)
			continue // Don't fail completely, just skip this package
		}
		processedPkgs++
	}

	tracef(opts, "Successfully processed %d packages\n", processedPkgs)

	// Extract module path and Go version from go.mod if available
	if err := extractModuleInfo(module); err != nil {
		warnf(opts, "Failed to extract module info: %v\n", err)
	}

	return nil
}

// processPackage processes a loaded package and adds it to the module.
func processPackage(module *typesys.Module, pkg *packages.Package, opts *typesys.LoadOptions) error {
	// Skip test packages unless explicitly requested
	if !opts.IncludeTests && strings.HasSuffix(pkg.PkgPath, ".test") {
		return nil
	}

	// Create a new package
	p := typesys.NewPackage(module, pkg.Name, pkg.PkgPath)
	p.TypesPackage = pkg.Types
	p.TypesInfo = pkg.TypesInfo

	// Set the package directory - prefer real filesystem path if available
	if len(pkg.GoFiles) > 0 {
		p.Dir = normalizePath(filepath.Dir(pkg.GoFiles[0]))
	} else {
		p.Dir = pkg.PkgPath
	}

	// Cache the package for later use
	module.CachePackage(pkg.PkgPath, pkg)

	// Add package to module
	module.Packages[pkg.PkgPath] = p

	// Build a comprehensive map of files for reliable path resolution
	// Map both by full path and by basename for robust lookups
	filePathMap := make(map[string]string)     // filename -> full path
	fileBaseMap := make(map[string]string)     // basename -> full path
	fileIdentMap := make(map[*ast.File]string) // AST file -> full path

	// Add all known Go files to our maps with normalized paths
	for _, path := range pkg.GoFiles {
		normalizedPath := normalizePath(path)
		base := filepath.Base(normalizedPath)
		filePathMap[normalizedPath] = normalizedPath
		fileBaseMap[base] = normalizedPath
	}

	for _, path := range pkg.CompiledGoFiles {
		normalizedPath := normalizePath(path)
		base := filepath.Base(normalizedPath)
		filePathMap[normalizedPath] = normalizedPath
		fileBaseMap[base] = normalizedPath
	}

	// First pass: Try to establish a direct mapping between AST files and file paths
	for i, astFile := range pkg.Syntax {
		if i < len(pkg.CompiledGoFiles) {
			fileIdentMap[astFile] = normalizePath(pkg.CompiledGoFiles[i])
		}
	}

	// Track processed files for debugging
	processedFiles := 0

	// Process files with improved path resolution
	for _, astFile := range pkg.Syntax {
		var filePath string

		// Try using our pre-computed map first
		if path, ok := fileIdentMap[astFile]; ok {
			filePath = path
		} else if astFile.Name != nil {
			// Fall back to looking up by filename
			filename := astFile.Name.Name
			if filename != "" {
				// Try with .go extension
				possibleName := filename + ".go"
				if path, ok := fileBaseMap[possibleName]; ok {
					filePath = path
				} else {
					// Look for partial matches as a last resort
					for base, path := range fileBaseMap {
						if strings.HasPrefix(base, filename) {
							filePath = path
							break
						}
					}
				}
			}
		}

		// If we still don't have a path, use position info from FileSet
		if filePath == "" && module.FileSet != nil {
			position := module.FileSet.Position(astFile.Pos())
			if position.IsValid() && position.Filename != "" {
				filePath = normalizePath(position.Filename)
			}
		}

		// If we still don't have a path, skip this file
		if filePath == "" {
			warnf(opts, "Could not determine file path for AST file in package %s\n", pkg.PkgPath)
			continue
		}

		// Ensure the path is absolute for consistency
		filePath = ensureAbsolutePath(filePath)

		// Create a new file
		file := typesys.NewFile(filePath, p)
		file.AST = astFile
		file.FileSet = module.FileSet

		// Add file to package
		p.AddFile(file)

		// Process imports
		processImports(file, astFile)

		processedFiles++
	}

	tracef(opts, "Processed %d/%d files for package %s\n", processedFiles, len(pkg.Syntax), pkg.PkgPath)
	if processedFiles < len(pkg.Syntax) {
		warnf(opts, "Not all files were processed for package %s\n", pkg.PkgPath)
	}

	// Process symbols (now that all files are loaded)
	processedSymbols := 0
	for _, file := range p.Files {
		beforeCount := len(p.Symbols)
		if err := processSymbols(p, file, opts); err != nil {
			errorf(opts, "Error processing symbols in file %s: %v\n", file.Path, err)
			continue // Don't fail completely, just skip this file
		}
		processedSymbols += len(p.Symbols) - beforeCount
	}

	if processedSymbols > 0 {
		tracef(opts, "Extracted %d symbols from package %s\n", processedSymbols, pkg.PkgPath)
	}

	return nil
}

// processImports processes imports in a file.
func processImports(file *typesys.File, astFile *ast.File) {
	for _, importSpec := range astFile.Imports {
		// Extract import path (removing quotes)
		path := strings.Trim(importSpec.Path.Value, "\"")

		// Create import
		imp := &typesys.Import{
			Path: path,
			File: file,
			Pos:  importSpec.Pos(),
			End:  importSpec.End(),
		}

		// Get local name if specified
		if importSpec.Name != nil {
			imp.Name = importSpec.Name.Name
		}

		// Add import to file
		file.AddImport(imp)
	}
}
