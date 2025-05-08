// Package loader provides implementations for loading Go modules.
package loader

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"

	"bitspark.dev/go-tree/pkg/core/module"
)

// GoModuleLoader implements ModuleLoader for Go modules
type GoModuleLoader struct {
	fset *token.FileSet
}

// NewGoModuleLoader creates a new module loader for Go modules
func NewGoModuleLoader() *GoModuleLoader {
	return &GoModuleLoader{
		fset: token.NewFileSet(),
	}
}

// Load loads a Go module with default options
func (l *GoModuleLoader) Load(dir string) (*module.Module, error) {
	return l.LoadWithOptions(dir, DefaultLoadOptions())
}

// LoadWithOptions loads a Go module with the specified options
func (l *GoModuleLoader) LoadWithOptions(dir string, options LoadOptions) (*module.Module, error) {
	// Check if dir is a valid Go module
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no go.mod file found in %s", dir)
	}

	// Parse go.mod file
	modContent, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	modFile, err := modfile.Parse(goModPath, modContent, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}

	// Create module
	mod := module.NewModule(modFile.Module.Mod.Path, dir)
	mod.GoVersion = modFile.Go.Version

	// Add dependencies
	for _, req := range modFile.Require {
		mod.AddDependency(req.Mod.Path, req.Mod.Version, req.Indirect)
	}

	// Add replacements
	for _, rep := range modFile.Replace {
		mod.AddReplace(rep.Old.Path, rep.Old.Version, rep.New.Path, rep.New.Version)
	}

	// Load packages
	pkgs, err := l.loadPackages(dir, options)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Convert loaded packages to module packages
	for _, pkg := range pkgs {
		modPkg := module.NewPackage(pkg.Name, pkg.PkgPath, pkg.Dir)

		// Process files in the package
		for _, file := range pkg.Syntax {
			filePath := l.fset.Position(file.Pos()).Filename
			fileName := filepath.Base(filePath)

			// Skip test files if not including tests
			isTest := strings.HasSuffix(fileName, "_test.go")
			if isTest && !options.IncludeTests {
				continue
			}

			// Create file
			modFile := module.NewFile(filePath, fileName, isTest)

			// Add imports
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, "\"")
				name := ""
				isBlank := false

				if imp.Name != nil {
					name = imp.Name.Name
					isBlank = name == "_"
				}

				importObj := &module.Import{
					Path:    path,
					Name:    name,
					IsBlank: isBlank,
				}

				modFile.AddImport(importObj)
			}

			// Process declarations in the file
			for _, decl := range file.Decls {
				l.processDeclaration(decl, modFile, modPkg, options)
			}

			// Set source code if requested
			if options.IncludeAST {
				modFile.AST = file
			}

			// Add file to package
			modPkg.AddFile(modFile)
		}

		// Add package to module
		mod.AddPackage(modPkg)
	}

	return mod, nil
}

// loadPackages loads Go packages using the go/packages API
func (l *GoModuleLoader) loadPackages(dir string, options LoadOptions) ([]*packages.Package, error) {
	// Configure the packages.Load call
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
		Dir:        dir,
		Fset:       l.fset,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(options.BuildTags, ","))},
	}

	// Determine patterns to load
	patterns := []string{"./..."}
	if len(options.PackagePaths) > 0 {
		patterns = options.PackagePaths
	}

	// Load the packages
	pkgs, err := packages.Load(config, patterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Check for errors in packages
	var errs []error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			errs = append(errs, fmt.Errorf("error in package %q: %v", pkg.PkgPath, err))
		}
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return pkgs, nil
}

// processDeclaration processes a declaration in a file
func (l *GoModuleLoader) processDeclaration(decl ast.Decl, file *module.File, pkg *module.Package, options LoadOptions) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		// Process function declaration
		l.processFunction(d, file, pkg, options)
	case *ast.GenDecl:
		// Process general declaration (type, var, const)
		l.processGenDecl(d, file, pkg, options)
	}
}

// processFunction processes a function declaration
func (l *GoModuleLoader) processFunction(funcDecl *ast.FuncDecl, file *module.File, pkg *module.Package, options LoadOptions) {
	name := funcDecl.Name.Name
	isExported := ast.IsExported(name)

	// Check if it's a test function
	isTest := strings.HasPrefix(name, "Test") && file.IsTest

	// Create function
	fn := module.NewFunction(name, isExported, isTest)

	// Set signature
	// In a real implementation, we would extract the full signature
	// This is simplified for this example
	fn.Signature = fmt.Sprintf("func %s(...) {...}", name)

	// Process receiver if it's a method
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		// Extract receiver info (simplified)
		recvField := funcDecl.Recv.List[0]
		recvName := ""
		if len(recvField.Names) > 0 {
			recvName = recvField.Names[0].Name
		}

		// Determine receiver type and whether it's a pointer
		recvType := ""
		isPointer := false
		switch rt := recvField.Type.(type) {
		case *ast.StarExpr:
			isPointer = true
			if ident, ok := rt.X.(*ast.Ident); ok {
				recvType = ident.Name
			}
		case *ast.Ident:
			recvType = rt.Name
		}

		// Set receiver
		fn.SetReceiver(recvName, recvType, isPointer)
	}

	// Set documentation if requested
	if options.LoadDocs && funcDecl.Doc != nil {
		fn.Doc = funcDecl.Doc.Text()
	}

	// Set AST node if requested
	if options.IncludeAST {
		fn.AST = funcDecl
	}

	// Add function to file and package
	file.AddFunction(fn)
	pkg.AddFunction(fn)
}

// processGenDecl processes a general declaration (type, var, const)
func (l *GoModuleLoader) processGenDecl(genDecl *ast.GenDecl, file *module.File, pkg *module.Package, options LoadOptions) {
	switch genDecl.Tok {
	case token.TYPE:
		// Process type declarations
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			name := typeSpec.Name.Name
			isExported := ast.IsExported(name)

			// Determine kind of type
			kind := "type"
			switch typeSpec.Type.(type) {
			case *ast.StructType:
				kind = "struct"
			case *ast.InterfaceType:
				kind = "interface"
			}

			// Create type
			typ := module.NewType(name, kind, isExported)

			// Set documentation if requested
			if options.LoadDocs {
				if genDecl.Doc != nil {
					typ.Doc = genDecl.Doc.Text()
				} else if typeSpec.Doc != nil {
					typ.Doc = typeSpec.Doc.Text()
				}
			}

			// Process struct fields or interface methods (simplified)
			if structType, ok := typeSpec.Type.(*ast.StructType); ok && structType.Fields != nil {
				for _, field := range structType.Fields.List {
					fieldName := ""
					isEmbedded := len(field.Names) == 0

					if !isEmbedded && len(field.Names) > 0 {
						fieldName = field.Names[0].Name
					}

					fieldType := "any" // Simplified, would extract actual type in full implementation
					tag := ""

					if field.Tag != nil {
						tag = field.Tag.Value
					}

					doc := ""
					if options.LoadDocs && field.Doc != nil {
						doc = field.Doc.Text()
					}

					typ.AddField(fieldName, fieldType, tag, isEmbedded, doc)
				}
			} else if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok && interfaceType.Methods != nil {
				for _, method := range interfaceType.Methods.List {
					methodName := ""
					isEmbedded := len(method.Names) == 0

					if !isEmbedded && len(method.Names) > 0 {
						methodName = method.Names[0].Name
					}

					signature := ""
					if !isEmbedded {
						signature = "func(...) ..." // Simplified
					}

					doc := ""
					if options.LoadDocs && method.Doc != nil {
						doc = method.Doc.Text()
					}

					typ.AddInterfaceMethod(methodName, signature, isEmbedded, doc)
				}
			}

			// Add type to file and package
			file.AddType(typ)
			pkg.AddType(typ)
		}

	case token.VAR:
		// Process variable declarations
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, ident := range valueSpec.Names {
				name := ident.Name
				isExported := ast.IsExported(name)

				typeName := "any" // Simplified
				value := ""

				if i < len(valueSpec.Values) {
					// Simplified: In a real implementation, we would extract the actual value
					value = "..."
				}

				doc := ""
				if options.LoadDocs && genDecl.Doc != nil {
					doc = genDecl.Doc.Text()
				}

				variable := module.NewVariable(name, typeName, value, isExported)
				variable.Doc = doc

				file.AddVariable(variable)
				pkg.AddVariable(variable)
			}
		}

	case token.CONST:
		// Process constant declarations
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, ident := range valueSpec.Names {
				name := ident.Name
				isExported := ast.IsExported(name)

				typeName := "any" // Simplified
				value := ""

				if i < len(valueSpec.Values) {
					// Simplified: In a real implementation, we would extract the actual value
					value = "..."
				}

				doc := ""
				if options.LoadDocs && genDecl.Doc != nil {
					doc = genDecl.Doc.Text()
				}

				constant := module.NewConstant(name, typeName, value, isExported)
				constant.Doc = doc

				file.AddConstant(constant)
				pkg.AddConstant(constant)
			}
		}
	}
}
