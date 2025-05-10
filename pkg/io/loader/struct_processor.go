package loader

import (
	"go/ast"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// processStructFields processes fields in a struct type and returns extracted symbols.
func processStructFields(pkg *typesys.Package, file *typesys.File, structSym *typesys.Symbol, structType *ast.StructType, opts *typesys.LoadOptions) []*typesys.Symbol {
	var symbols []*typesys.Symbol

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
				sym := createSymbol(pkg, file, typeName, typesys.KindEmbeddedField, field.Pos(), field.End(), structSym)

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
			sym := createSymbol(pkg, file, name.Name, typesys.KindField, name.Pos(), name.End(), structSym)

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
func processInterfaceMethods(pkg *typesys.Package, file *typesys.File, interfaceSym *typesys.Symbol, interfaceType *ast.InterfaceType, opts *typesys.LoadOptions) []*typesys.Symbol {
	var symbols []*typesys.Symbol

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
				sym := createSymbol(pkg, file, typeName, typesys.KindEmbeddedInterface, method.Pos(), method.End(), interfaceSym)

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
			sym := createSymbol(pkg, file, name.Name, typesys.KindMethod, name.Pos(), name.End(), interfaceSym)

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
