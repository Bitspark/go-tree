package typesys

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// LoadModule loads a Go module with full type checking.
func LoadModule(dir string, opts *LoadOptions) (*Module, error) {
	if opts == nil {
		opts = &LoadOptions{
			IncludeTests:   false,
			IncludePrivate: true,
		}
	}

	// Normalize and make directory path absolute
	dir = ensureAbsolutePath(normalizePath(dir))

	// Create a new module
	module := NewModule(dir)

	// Load packages
	if err := loadPackages(module, opts); err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	return module, nil
}

// loadPackages loads all Go packages in the module directory.
func loadPackages(module *Module, opts *LoadOptions) error {
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
func processPackage(module *Module, pkg *packages.Package, opts *LoadOptions) error {
	// Skip test packages unless explicitly requested
	if !opts.IncludeTests && strings.HasSuffix(pkg.PkgPath, ".test") {
		return nil
	}

	// Create a new package
	p := NewPackage(module, pkg.Name, pkg.PkgPath)
	p.TypesPackage = pkg.Types
	p.TypesInfo = pkg.TypesInfo

	// Set the package directory - prefer real filesystem path if available
	if len(pkg.GoFiles) > 0 {
		p.Dir = normalizePath(filepath.Dir(pkg.GoFiles[0]))
	} else {
		p.Dir = pkg.PkgPath
	}

	// Cache the package for later use
	module.pkgCache[pkg.PkgPath] = pkg

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
		file := NewFile(filePath, p)
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
func processImports(file *File, astFile *ast.File) {
	for _, importSpec := range astFile.Imports {
		// Extract import path (removing quotes)
		path := strings.Trim(importSpec.Path.Value, "\"")

		// Create import
		imp := &Import{
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

// processSymbols processes all symbols in a file.
func processSymbols(pkg *Package, file *File, opts *LoadOptions) error {
	// Get the AST file
	astFile := file.AST

	if astFile == nil {
		warnf(opts, "Missing AST for file %s\n", file.Path)
		return nil
	}

	tracef(opts, "Processing symbols in file: %s\n", file.Path)

	declCount := 0
	symbolCount := 0

	// Track any errors during processing
	var processingErrors []error

	// Process declarations
	for _, decl := range astFile.Decls {
		declCount++

		// Use processSafely to catch any unexpected issues
		err := processSafely(file, func() error {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if syms := processFuncDecl(pkg, file, d, opts); len(syms) > 0 {
					symbolCount += len(syms)
				}
			case *ast.GenDecl:
				if syms := processGenDecl(pkg, file, d, opts); len(syms) > 0 {
					symbolCount += len(syms)
				}
			}
			return nil
		}, opts)

		if err != nil {
			processingErrors = append(processingErrors, err)
		}
	}

	tracef(opts, "Processed %d declarations in file %s, extracted %d symbols\n",
		declCount, file.Path, symbolCount)

	if len(processingErrors) > 0 {
		tracef(opts, "Encountered %d errors during symbol processing in %s\n",
			len(processingErrors), file.Path)
	}

	return nil
}

// processFuncDecl processes a function declaration and returns extracted symbols.
func processFuncDecl(pkg *Package, file *File, funcDecl *ast.FuncDecl, opts *LoadOptions) []*Symbol {
	// Skip if invalid or should not be included
	if funcDecl.Name == nil || funcDecl.Name.Name == "" ||
		!shouldIncludeSymbol(funcDecl.Name.Name, opts) {
		return nil
	}

	// Determine if this is a method
	isMethod := funcDecl.Recv != nil

	// Create a new symbol using helper
	kind := KindFunction
	if isMethod {
		kind = KindMethod
	}

	sym := createSymbol(pkg, file, funcDecl.Name.Name, kind, funcDecl.Pos(), funcDecl.End(), nil)

	// Extract type info
	obj, typeInfo := extractTypeInfo(pkg, funcDecl.Name, nil)
	sym.TypeObj = obj
	if fn, ok := typeInfo.(*types.Signature); ok {
		sym.TypeInfo = fn
	}

	// If method, add receiver information
	if isMethod && funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		recv := funcDecl.Recv.List[0]
		if recv.Type != nil {
			// Get base type without * (pointer)
			recvTypeExpr := recv.Type
			if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
				recvTypeExpr = starExpr.X
			}

			// Get receiver type name
			recvType := exprToString(recvTypeExpr)
			if recvType != "" {
				// Find parent type
				parentSyms := pkg.SymbolByName(recvType, KindType, KindStruct, KindInterface)
				if len(parentSyms) > 0 {
					sym.Parent = parentSyms[0]
				}
			}
		}
	}

	// Add the symbol to the file
	file.AddSymbol(sym)

	return []*Symbol{sym}
}

// processGenDecl processes a general declaration (type, var, const) and returns extracted symbols.
func processGenDecl(pkg *Package, file *File, genDecl *ast.GenDecl, opts *LoadOptions) []*Symbol {
	var symbols []*Symbol

	for _, spec := range genDecl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			// Skip if invalid or should not be included
			if s.Name == nil || s.Name.Name == "" ||
				!shouldIncludeSymbol(s.Name.Name, opts) {
				continue
			}

			// Determine kind
			kind := KindType
			if _, ok := s.Type.(*ast.StructType); ok {
				kind = KindStruct
			} else if _, ok := s.Type.(*ast.InterfaceType); ok {
				kind = KindInterface
			}

			// Create symbol using helper
			sym := createSymbol(pkg, file, s.Name.Name, kind, s.Pos(), s.End(), nil)

			// Extract type information
			obj, typeInfo := extractTypeInfo(pkg, s.Name, nil)
			sym.TypeObj = obj
			sym.TypeInfo = typeInfo

			// Add the symbol to the file
			file.AddSymbol(sym)
			symbols = append(symbols, sym)

			// Process struct fields or interface methods
			switch t := s.Type.(type) {
			case *ast.StructType:
				if fieldSyms := processStructFields(pkg, file, sym, t, opts); len(fieldSyms) > 0 {
					symbols = append(symbols, fieldSyms...)
				}
			case *ast.InterfaceType:
				if methodSyms := processInterfaceMethods(pkg, file, sym, t, opts); len(methodSyms) > 0 {
					symbols = append(symbols, methodSyms...)
				}
			}

		case *ast.ValueSpec:
			// Process each name in the value spec
			for i, name := range s.Names {
				// Skip if invalid or should not be included
				if name.Name == "" || !shouldIncludeSymbol(name.Name, opts) {
					continue
				}

				// Determine kind
				kind := KindVariable
				if genDecl.Tok == token.CONST {
					kind = KindConstant
				}

				// Create symbol using helper
				sym := createSymbol(pkg, file, name.Name, kind, name.Pos(), name.End(), nil)

				// Extract type information
				obj, typeInfo := extractTypeInfo(pkg, name, nil)
				if obj != nil {
					sym.TypeObj = obj
					sym.TypeInfo = typeInfo
				} else {
					// Fall back to AST-based type inference if type checker data is unavailable
					if s.Type != nil {
						// Get type from declaration
						sym.TypeInfo = pkg.TypesInfo.TypeOf(s.Type)
					} else if i < len(s.Values) {
						// Infer type from value
						sym.TypeInfo = pkg.TypesInfo.TypeOf(s.Values[i])
					}
				}

				// Add the symbol to the file
				file.AddSymbol(sym)
				symbols = append(symbols, sym)
			}
		}
	}

	return symbols
}

// processStructFields processes fields in a struct type and returns extracted symbols.
func processStructFields(pkg *Package, file *File, structSym *Symbol, structType *ast.StructType, opts *LoadOptions) []*Symbol {
	var symbols []*Symbol

	if structType.Fields == nil {
		return nil
	}

	for _, field := range structType.Fields.List {
		// Handle embedded types (those without field names)
		if len(field.Names) == 0 {
			// Try to get the embedded type name
			typeName := exprToString(field.Type)
			if typeName != "" {
				// Create a special field symbol for the embedded type using helper
				sym := createSymbol(pkg, file, typeName, KindEmbeddedField, field.Pos(), field.End(), structSym)

				// Try to get type information
				_, typeInfo := extractTypeInfo(pkg, nil, field.Type)
				sym.TypeInfo = typeInfo

				// Add the symbol
				file.AddSymbol(sym)
				symbols = append(symbols, sym)
			}
			continue
		}

		// Process named fields
		for _, name := range field.Names {
			// Skip if invalid or should not be included
			if name.Name == "" || !shouldIncludeSymbol(name.Name, opts) {
				continue
			}

			// Create field symbol using helper
			sym := createSymbol(pkg, file, name.Name, KindField, name.Pos(), name.End(), structSym)

			// Extract type information
			obj, typeInfo := extractTypeInfo(pkg, name, field.Type)
			if obj != nil {
				sym.TypeObj = obj
				sym.TypeInfo = typeInfo
			} else if typeInfo != nil {
				// Fallback to just the type info
				sym.TypeInfo = typeInfo
			}

			// Add the symbol to the file
			file.AddSymbol(sym)
			symbols = append(symbols, sym)
		}
	}

	return symbols
}

// processInterfaceMethods processes methods in an interface type and returns extracted symbols.
func processInterfaceMethods(pkg *Package, file *File, interfaceSym *Symbol, interfaceType *ast.InterfaceType, opts *LoadOptions) []*Symbol {
	var symbols []*Symbol

	if interfaceType.Methods == nil {
		return nil
	}

	for _, method := range interfaceType.Methods.List {
		// Handle embedded interfaces
		if len(method.Names) == 0 {
			// Get the embedded interface name
			typeName := exprToString(method.Type)
			if typeName != "" {
				// Create a special symbol for the embedded interface using helper
				sym := createSymbol(pkg, file, typeName, KindEmbeddedInterface, method.Pos(), method.End(), interfaceSym)

				// Extract type information
				_, typeInfo := extractTypeInfo(pkg, nil, method.Type)
				sym.TypeInfo = typeInfo

				// Add the symbol
				file.AddSymbol(sym)
				symbols = append(symbols, sym)
			}
			continue
		}

		// Process named methods
		for _, name := range method.Names {
			// Skip if invalid or should not be included
			if name.Name == "" || !shouldIncludeSymbol(name.Name, opts) {
				continue
			}

			// Create method symbol using helper
			sym := createSymbol(pkg, file, name.Name, KindMethod, name.Pos(), name.End(), interfaceSym)

			// Extract type information
			obj, typeInfo := extractTypeInfo(pkg, name, nil)
			if obj != nil {
				sym.TypeObj = obj
				sym.TypeInfo = typeInfo
			} else if methodType, ok := method.Type.(*ast.FuncType); ok {
				// Fallback to AST-based type info
				sym.TypeInfo = pkg.TypesInfo.TypeOf(methodType)
			}

			// Add the symbol to the file
			file.AddSymbol(sym)
			symbols = append(symbols, sym)
		}
	}

	return symbols
}

// Helper function to extract module info from go.mod
func extractModuleInfo(module *Module) error {
	// Check if go.mod exists
	goModPath := filepath.Join(module.Dir, "go.mod")
	goModPath = normalizePath(goModPath)

	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found in %s", module.Dir)
	}

	// Read go.mod
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Parse module path and Go version more robustly
	lines := strings.Split(string(content), "\n")
	inMultilineBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle multiline blocks
		if strings.Contains(line, "(") {
			inMultilineBlock = true
			continue
		}

		if strings.Contains(line, ")") {
			inMultilineBlock = false
			continue
		}

		// Skip lines in multiline blocks
		if inMultilineBlock {
			continue
		}

		// Handle module declaration with proper word boundary checking
		if strings.HasPrefix(line, "module ") {
			// Extract the module path, handling quotes if present
			modulePath := strings.TrimPrefix(line, "module ")
			modulePath = strings.TrimSpace(modulePath)

			// Handle quoted module paths
			if strings.HasPrefix(modulePath, "\"") && strings.HasSuffix(modulePath, "\"") {
				modulePath = modulePath[1 : len(modulePath)-1]
			} else if strings.HasPrefix(modulePath, "'") && strings.HasSuffix(modulePath, "'") {
				modulePath = modulePath[1 : len(modulePath)-1]
			}

			module.Path = modulePath
		} else if strings.HasPrefix(line, "go ") {
			// Extract go version
			goVersion := strings.TrimPrefix(line, "go ")
			goVersion = strings.TrimSpace(goVersion)

			// Handle quoted go versions
			if strings.HasPrefix(goVersion, "\"") && strings.HasSuffix(goVersion, "\"") {
				goVersion = goVersion[1 : len(goVersion)-1]
			} else if strings.HasPrefix(goVersion, "'") && strings.HasSuffix(goVersion, "'") {
				goVersion = goVersion[1 : len(goVersion)-1]
			}

			module.GoVersion = goVersion
		}
	}

	// Validate that we found a module path
	if module.Path == "" {
		return fmt.Errorf("no module declaration found in go.mod")
	}

	return nil
}

// Helper function to convert an expression to a string representation
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	}
	return ""
}
