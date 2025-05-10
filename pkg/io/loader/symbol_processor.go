package loader

import (
	"go/ast"
	"go/token"
	"go/types"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// processSymbols processes all symbols in a file.
func processSymbols(pkg *typesys.Package, file *typesys.File, opts *typesys.LoadOptions) error {
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
func processFuncDecl(pkg *typesys.Package, file *typesys.File, funcDecl *ast.FuncDecl, opts *typesys.LoadOptions) []*typesys.Symbol {
	// Skip if invalid or should not be included
	if funcDecl.Name == nil || funcDecl.Name.Name == "" ||
		!shouldIncludeSymbol(funcDecl.Name.Name, opts) {
		return nil
	}

	// Determine if this is a method
	isMethod := funcDecl.Recv != nil

	// Create a new symbol using helper
	kind := typesys.KindFunction
	if isMethod {
		kind = typesys.KindMethod
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
				parentSyms := pkg.SymbolByName(recvType, typesys.KindType, typesys.KindStruct, typesys.KindInterface)
				if len(parentSyms) > 0 {
					sym.Parent = parentSyms[0]
				}
			}
		}
	}

	// Add the symbol to the file
	file.AddSymbol(sym)

	return []*typesys.Symbol{sym}
}

// processGenDecl processes a general declaration (type, var, const) and returns extracted symbols.
func processGenDecl(pkg *typesys.Package, file *typesys.File, genDecl *ast.GenDecl, opts *typesys.LoadOptions) []*typesys.Symbol {
	var symbols []*typesys.Symbol

	for _, spec := range genDecl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			// Skip if invalid or should not be included
			if s.Name == nil || s.Name.Name == "" ||
				!shouldIncludeSymbol(s.Name.Name, opts) {
				continue
			}

			// Determine kind
			kind := typesys.KindType
			if _, ok := s.Type.(*ast.StructType); ok {
				kind = typesys.KindStruct
			} else if _, ok := s.Type.(*ast.InterfaceType); ok {
				kind = typesys.KindInterface
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
				kind := typesys.KindVariable
				if genDecl.Tok == token.CONST {
					kind = typesys.KindConstant
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

// Helper function to extract type information from an AST node
func extractTypeInfo(pkg *typesys.Package, nameNode *ast.Ident, typeNode ast.Expr) (types.Object, types.Type) {
	// Try to get type object from identifier
	if nameNode != nil && pkg.TypesInfo != nil {
		if obj := pkg.TypesInfo.ObjectOf(nameNode); obj != nil {
			return obj, obj.Type()
		}
	}

	// Fall back to type expression if available
	if typeNode != nil && pkg.TypesInfo != nil {
		if typeInfo := pkg.TypesInfo.TypeOf(typeNode); typeInfo != nil {
			return nil, typeInfo
		}
	}

	return nil, nil
}
