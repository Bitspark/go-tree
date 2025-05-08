package typesys

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
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

	if opts.Trace {
		fmt.Printf("Loading packages from directory: %s with pattern %s\n", module.Dir, pattern)
	}

	// Load packages
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	if opts.Trace {
		fmt.Printf("Loaded %d packages\n", len(pkgs))
	}

	// Debug any package errors
	var pkgsWithErrors int
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			pkgsWithErrors++
			if opts.Trace {
				fmt.Printf("Package %s has %d errors:\n", pkg.PkgPath, len(pkg.Errors))
				for _, err := range pkg.Errors {
					fmt.Printf("  - %v\n", err)
				}
			}
		}
	}

	if pkgsWithErrors > 0 && opts.Trace {
		fmt.Printf("%d packages had errors\n", pkgsWithErrors)
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
			if opts.Trace {
				fmt.Printf("Error processing package %s: %v\n", pkg.PkgPath, err)
			}
			continue // Don't fail completely, just skip this package
		}
		processedPkgs++
	}

	if opts.Trace {
		fmt.Printf("Successfully processed %d packages\n", processedPkgs)
	}

	// Extract module path and Go version from go.mod if available
	if err := extractModuleInfo(module); err != nil && opts.Trace {
		fmt.Printf("Warning: failed to extract module info: %v\n", err)
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
	p.Dir = pkg.PkgPath

	// Cache the package for later use
	module.pkgCache[pkg.PkgPath] = pkg

	// Add package to module
	module.Packages[pkg.PkgPath] = p

	// Build a map of all available file paths to use as fallbacks
	// This is needed because CompiledGoFiles might not match Syntax exactly
	filePathMap := make(map[string]string)
	for _, path := range pkg.GoFiles {
		base := filepath.Base(path)
		filePathMap[base] = path
	}
	for _, path := range pkg.CompiledGoFiles {
		base := filepath.Base(path)
		filePathMap[base] = path
	}

	// Track processed files for debugging
	processedFiles := 0

	// Process files - with improved file path handling
	for i, astFile := range pkg.Syntax {
		var filePath string

		// First try to use CompiledGoFiles
		if i < len(pkg.CompiledGoFiles) {
			filePath = pkg.CompiledGoFiles[i]
		} else if astFile.Name != nil {
			// Fall back to looking up by filename in our map
			fileName := astFile.Name.Name
			if fileName != "" {
				// Try to find a matching file using the filename
				for base, path := range filePathMap {
					if strings.HasPrefix(base, fileName) {
						filePath = path
						break
					}
				}

				// If still not found, construct a path
				if filePath == "" {
					possibleName := fileName + ".go"
					if path, ok := filePathMap[possibleName]; ok {
						filePath = path
					} else {
						// Last resort: use package path + filename
						filePath = filepath.Join(pkg.PkgPath, fileName+".go")
					}
				}
			}
		}

		// If we still don't have a path, skip this file
		if filePath == "" {
			if opts.Trace {
				fmt.Printf("Warning: Could not determine file path for AST file in package %s\n", pkg.PkgPath)
			}
			continue
		}

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

	if opts.Trace && processedFiles > 0 {
		fmt.Printf("Processed %d files for package %s\n", processedFiles, pkg.PkgPath)
	}

	// Process symbols (now that all files are loaded)
	processedSymbols := 0
	for _, file := range p.Files {
		beforeCount := len(p.Symbols)
		if err := processSymbols(p, file, opts); err != nil {
			if opts.Trace {
				fmt.Printf("Error processing symbols in file %s: %v\n", file.Path, err)
			}
			continue // Don't fail completely, just skip this file
		}
		processedSymbols += len(p.Symbols) - beforeCount
	}

	if opts.Trace && processedSymbols > 0 {
		fmt.Printf("Extracted %d symbols from package %s\n", processedSymbols, pkg.PkgPath)
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
		if opts.Trace {
			fmt.Printf("Warning: Missing AST for file %s\n", file.Path)
		}
		return nil
	}

	if opts.Trace {
		fmt.Printf("Processing symbols in file: %s\n", file.Path)
	}

	declCount := 0

	// Process declarations
	for _, decl := range astFile.Decls {
		declCount++
		switch d := decl.(type) {
		case *ast.FuncDecl:
			processFuncDecl(pkg, file, d, opts)
		case *ast.GenDecl:
			processGenDecl(pkg, file, d, opts)
		}
	}

	if opts.Trace {
		fmt.Printf("Processed %d declarations in file %s\n", declCount, file.Path)
	}

	return nil
}

// processFuncDecl processes a function declaration.
func processFuncDecl(pkg *Package, file *File, funcDecl *ast.FuncDecl, opts *LoadOptions) {
	// Skip unexported functions if not including private symbols
	if !opts.IncludePrivate && !ast.IsExported(funcDecl.Name.Name) {
		return
	}

	// Determine if this is a method
	isMethod := funcDecl.Recv != nil

	// Create a new symbol
	kind := KindFunction
	if isMethod {
		kind = KindMethod
	}

	sym := NewSymbol(funcDecl.Name.Name, kind)
	sym.Pos = funcDecl.Pos()
	sym.End = funcDecl.End()
	sym.File = file
	sym.Package = pkg

	// Get position info
	if posInfo := file.GetPositionInfo(funcDecl.Pos(), funcDecl.End()); posInfo != nil {
		sym.AddDefinition(file.Path, funcDecl.Pos(), posInfo.LineStart, posInfo.ColumnStart)
	}

	// If method, add receiver information
	if isMethod && len(funcDecl.Recv.List) > 0 {
		// Get receiver type
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
}

// processGenDecl processes a general declaration (type, var, const).
func processGenDecl(pkg *Package, file *File, genDecl *ast.GenDecl, opts *LoadOptions) {
	for _, spec := range genDecl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			// Skip unexported types if not including private symbols
			if !opts.IncludePrivate && !ast.IsExported(s.Name.Name) {
				continue
			}

			// Determine kind
			kind := KindType
			if _, ok := s.Type.(*ast.StructType); ok {
				kind = KindStruct
			} else if _, ok := s.Type.(*ast.InterfaceType); ok {
				kind = KindInterface
			}

			// Create symbol
			sym := NewSymbol(s.Name.Name, kind)
			sym.Pos = s.Pos()
			sym.End = s.End()
			sym.File = file
			sym.Package = pkg

			// Get position info
			if posInfo := file.GetPositionInfo(s.Pos(), s.End()); posInfo != nil {
				sym.AddDefinition(file.Path, s.Pos(), posInfo.LineStart, posInfo.ColumnStart)
			}

			// Add the symbol to the file
			file.AddSymbol(sym)

			// Process struct fields or interface methods
			switch t := s.Type.(type) {
			case *ast.StructType:
				processStructFields(pkg, file, sym, t, opts)
			case *ast.InterfaceType:
				processInterfaceMethods(pkg, file, sym, t, opts)
			}

		case *ast.ValueSpec:
			// Process each name in the value spec
			for i, name := range s.Names {
				// Skip unexported names if not including private symbols
				if !opts.IncludePrivate && !ast.IsExported(name.Name) {
					continue
				}

				// Determine kind
				kind := KindVariable
				if genDecl.Tok == token.CONST {
					kind = KindConstant
				}

				// Create symbol
				sym := NewSymbol(name.Name, kind)
				sym.Pos = name.Pos()
				sym.End = name.End()
				sym.File = file
				sym.Package = pkg

				// Get type info if available
				if s.Type != nil {
					// Get type name as string
					typeStr := exprToString(s.Type)
					if typeStr != "" {
						sym.TypeInfo = pkg.TypesInfo.TypeOf(s.Type)
					}
				} else if i < len(s.Values) {
					// Infer type from value
					sym.TypeInfo = pkg.TypesInfo.TypeOf(s.Values[i])
				}

				// Get position info
				if posInfo := file.GetPositionInfo(name.Pos(), name.End()); posInfo != nil {
					sym.AddDefinition(file.Path, name.Pos(), posInfo.LineStart, posInfo.ColumnStart)
				}

				// Add the symbol to the file
				file.AddSymbol(sym)
			}
		}
	}
}

// processStructFields processes fields in a struct type.
func processStructFields(pkg *Package, file *File, structSym *Symbol, structType *ast.StructType, opts *LoadOptions) {
	if structType.Fields == nil {
		return
	}

	for _, field := range structType.Fields.List {
		// Skip field without names (embedded types)
		if len(field.Names) == 0 {
			// TODO: Handle embedded types
			continue
		}

		for _, name := range field.Names {
			// Skip unexported fields if not including private symbols
			if !opts.IncludePrivate && !ast.IsExported(name.Name) {
				continue
			}

			// Create field symbol
			sym := NewSymbol(name.Name, KindField)
			sym.Pos = name.Pos()
			sym.End = name.End()
			sym.File = file
			sym.Package = pkg
			sym.Parent = structSym

			// Get type info if available
			if field.Type != nil {
				sym.TypeInfo = pkg.TypesInfo.TypeOf(field.Type)
			}

			// Get position info
			if posInfo := file.GetPositionInfo(name.Pos(), name.End()); posInfo != nil {
				sym.AddDefinition(file.Path, name.Pos(), posInfo.LineStart, posInfo.ColumnStart)
			}

			// Add the symbol to the file
			file.AddSymbol(sym)
		}
	}
}

// processInterfaceMethods processes methods in an interface type.
func processInterfaceMethods(pkg *Package, file *File, interfaceSym *Symbol, interfaceType *ast.InterfaceType, opts *LoadOptions) {
	if interfaceType.Methods == nil {
		return
	}

	for _, method := range interfaceType.Methods.List {
		// Skip embedded interfaces
		if len(method.Names) == 0 {
			// TODO: Handle embedded interfaces
			continue
		}

		for _, name := range method.Names {
			// Interface methods are always exported
			if !ast.IsExported(name.Name) && !opts.IncludePrivate {
				continue
			}

			// Create method symbol
			sym := NewSymbol(name.Name, KindMethod)
			sym.Pos = name.Pos()
			sym.End = name.End()
			sym.File = file
			sym.Package = pkg
			sym.Parent = interfaceSym

			// Get position info
			if posInfo := file.GetPositionInfo(name.Pos(), name.End()); posInfo != nil {
				sym.AddDefinition(file.Path, name.Pos(), posInfo.LineStart, posInfo.ColumnStart)
			}

			// Add the symbol to the file
			file.AddSymbol(sym)
		}
	}
}

// Helper function to extract module info from go.mod
func extractModuleInfo(module *Module) error {
	// Check if go.mod exists
	goModPath := filepath.Join(module.Dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found in %s", module.Dir)
	}

	// Read go.mod
	content, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Parse module path
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module.Path = strings.TrimSpace(strings.TrimPrefix(line, "module"))
		} else if strings.HasPrefix(line, "go ") {
			module.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go"))
		}
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
