// Package parser provides functionality for parsing Go packages into model structures
package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"bitspark.dev/go-tree/internal/model"
)

// fileSet represents a set of parsed Go files
type fileSet struct {
	Files map[string]*ast.File
}

// ParsePackage parses a Go package from the specified directory and returns a model.GoPackage
func ParsePackage(dir string) (*model.GoPackage, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for directory: %w", err)
	}

	fset := token.NewFileSet()
	// Parse all files in the directory, excluding tests, with comments
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		name := fi.Name()
		return !fi.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no non-test Go files found in %s", dir)
	}

	// Assume a single package (use the first one in map)
	var files *fileSet
	var pkgName string
	for name, pkg := range pkgs {
		files = &fileSet{Files: pkg.Files}
		pkgName = name
		break
	}
	if len(pkgs) > 1 {
		fmt.Fprintf(os.Stderr, "Warning: multiple packages in directory; using package %s\n", pkgName)
	}
	pkg := &model.GoPackage{Name: pkgName}

	// Read all files' contents for extracting original source snippets
	fileContents := make(map[string][]byte)
	for filename := range files.Files {
		// Validate the filename is within the package directory to prevent path traversal
		absFilename, err := filepath.Abs(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for file %s: %w", filename, err)
		}

		// Ensure file is within the package directory
		if !strings.HasPrefix(absFilename, absDir) {
			return nil, fmt.Errorf("file %s is outside package directory", filename)
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filename, err)
		}
		fileContents[filename] = data
	}

	// Process imports
	if err := parseImports(pkg, files, fset, fileContents); err != nil {
		return nil, err
	}

	// Process declarations
	if err := parseDeclarations(pkg, files, fset, fileContents); err != nil {
		return nil, err
	}

	// Extract package-level comments
	extractPackageDocs(pkg, files, fset, fileContents)

	return pkg, nil
}

// parseImports extracts all imports from the package
func parseImports(pkg *model.GoPackage, files *fileSet, fset *token.FileSet, fileContents map[string][]byte) error {
	// Prepare to collect imports (unique by path) with insertion order
	importMap := make(map[string]model.GoImport)
	var importOrder []string

	for _, file := range files.Files {
		for _, importSpec := range file.Imports {
			pathLit := importSpec.Path.Value // e.g. "\"fmt\""
			path, err := strconv.Unquote(pathLit)
			if err != nil {
				path = strings.Trim(pathLit, "\"")
			}
			alias := ""
			if importSpec.Name != nil {
				alias = importSpec.Name.Name
			}
			if _, exists := importMap[path]; !exists {
				imp := model.GoImport{Path: path, Alias: alias}
				if importSpec.Doc != nil {
					imp.Doc = strings.TrimSpace(importSpec.Doc.Text())
				}
				if importSpec.Comment != nil {
					imp.Comment = strings.TrimSpace(importSpec.Comment.Text())
				}
				importMap[path] = imp
				importOrder = append(importOrder, path)
			} else {
				// Duplicate import path encountered
				existing := importMap[path]
				if alias != "" && existing.Alias != "" && existing.Alias != alias {
					fmt.Fprintf(os.Stderr, "Warning: import %q with conflicting aliases (%s vs %s)\n", path, existing.Alias, alias)
				}
				// (We keep the first alias and ignore duplicates for output.)
			}
		}
	}

	// Finalize Imports in model (preserve original encounter order)
	for _, path := range importOrder {
		pkg.Imports = append(pkg.Imports, importMap[path])
	}

	return nil
}

// parseDeclarations extracts all top-level declarations from the package
func parseDeclarations(pkg *model.GoPackage, files *fileSet, fset *token.FileSet, fileContents map[string][]byte) error {
	// Track declarations by position for ordering
	type declInfo struct {
		declaration interface{}
		pos         token.Pos
	}
	var declarations []declInfo

	// Process each file's declarations
	for filename, file := range files.Files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				// Function or method
				if function, err := parseFunction(d, fset, fileContents[filename]); err == nil {
					declarations = append(declarations, declInfo{declaration: function, pos: d.Pos()})
				} else {
					return err
				}
			case *ast.GenDecl:
				switch d.Tok {
				case token.CONST:
					// Constants
					consts, err := parseConstants(d, fset, fileContents[filename])
					if err != nil {
						return err
					}
					for _, c := range consts {
						declarations = append(declarations, declInfo{declaration: c, pos: d.Pos()})
					}
				case token.VAR:
					// Variables
					vars, err := parseVariables(d, fset, fileContents[filename])
					if err != nil {
						return err
					}
					for _, v := range vars {
						declarations = append(declarations, declInfo{declaration: v, pos: d.Pos()})
					}
				case token.TYPE:
					// Type definitions
					types, err := parseTypes(d, fset, fileContents[filename])
					if err != nil {
						return err
					}
					for _, t := range types {
						declarations = append(declarations, declInfo{declaration: t, pos: d.Pos()})
					}
				}
			}
		}

		// Handle standalone comment groups
		for _, cg := range file.Comments {
			// Skip comments that are already associated with declarations
			isStandalone := true
			for _, decl := range file.Decls {
				if isCommentAssociatedWithDecl(cg, decl, fset) {
					isStandalone = false
					break
				}
			}

			if isStandalone && cg.Pos() > file.Package {
				// This is a standalone comment group (not a license or package doc)
				start := fset.Position(cg.Pos()).Offset
				end := fset.Position(cg.End()).Offset
				if start >= 0 && end <= len(fileContents[filename]) {
					commentText := string(fileContents[filename][start:end])
					comment := model.Declaration{
						Type:     "comment",
						Position: start,
						Data:     commentText,
					}
					declarations = append(declarations, declInfo{declaration: comment, pos: cg.Pos()})
				}
			}
		}
	}

	// Sort declarations by position
	sort.Slice(declarations, func(i, j int) bool {
		return declarations[i].pos < declarations[j].pos
	})

	// Add declarations to package model
	for _, d := range declarations {
		switch decl := d.declaration.(type) {
		case *model.GoFunction:
			pkg.Functions = append(pkg.Functions, *decl)
		case *model.GoConstant:
			pkg.Constants = append(pkg.Constants, *decl)
		case *model.GoVariable:
			pkg.Variables = append(pkg.Variables, *decl)
		case *model.GoType:
			pkg.Types = append(pkg.Types, *decl)
		case model.Declaration:
			// Standalone comments are handled separately
			// They could be stored in a separate field if needed
		}
	}

	return nil
}

// isCommentAssociatedWithDecl checks if a comment group is associated with a declaration
func isCommentAssociatedWithDecl(cg *ast.CommentGroup, decl ast.Decl, fset *token.FileSet) bool {
	switch d := decl.(type) {
	case *ast.GenDecl:
		if d.Doc == cg {
			return true
		}
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.ValueSpec:
				if s.Doc == cg || s.Comment == cg {
					return true
				}
			case *ast.TypeSpec:
				if s.Doc == cg || s.Comment == cg {
					return true
				}
			case *ast.ImportSpec:
				if s.Doc == cg || s.Comment == cg {
					return true
				}
			}
		}
	case *ast.FuncDecl:
		if d.Doc == cg {
			return true
		}
	}

	// Check if comment is between the declaration's start and end
	declStart := fset.Position(decl.Pos()).Offset
	declEnd := fset.Position(decl.End()).Offset
	commentStart := fset.Position(cg.Pos()).Offset

	return commentStart >= declStart && commentStart < declEnd
}

// parseFunction extracts a function or method declaration
func parseFunction(d *ast.FuncDecl, fset *token.FileSet, src []byte) (*model.GoFunction, error) {
	funcName := d.Name.Name

	// Handle receiver for methods
	var recvInfo *model.GoReceiver
	if d.Recv != nil && len(d.Recv.List) > 0 {
		recvField := d.Recv.List[0]
		var recvTypeBuf bytes.Buffer
		if err := format.Node(&recvTypeBuf, fset, recvField.Type); err == nil {
			recvType := recvTypeBuf.String()
			recvName := ""
			if len(recvField.Names) > 0 {
				recvName = recvField.Names[0].Name
			}
			recvInfo = &model.GoReceiver{Name: recvName, Type: recvType}
		}
	}

	// Function signature
	var sigBuf bytes.Buffer
	_ = format.Node(&sigBuf, fset, d.Type)
	signature := sigBuf.String()

	// Extract full function source code (including comments)
	start := d.Pos()
	if d.Doc != nil {
		start = d.Doc.Pos()
	}
	end := d.End()
	startOff := fset.Position(start).Offset
	endOff := fset.Position(end).Offset
	fullCode := ""
	if startOff >= 0 && endOff <= len(src) && startOff < endOff {
		fullCode = string(src[startOff:endOff])
	}

	// Extract function body code
	bodyCode := ""
	if d.Body != nil {
		bodyStart := fset.Position(d.Body.Lbrace + 1).Offset
		bodyEnd := fset.Position(d.Body.Rbrace).Offset
		if bodyStart < bodyEnd && bodyEnd <= len(src) {
			bodyCode = string(src[bodyStart:bodyEnd])
		}
	}

	// Documentation
	doc := ""
	if d.Doc != nil {
		doc = strings.TrimSpace(d.Doc.Text())
	}

	return &model.GoFunction{
		Name:      funcName,
		Receiver:  recvInfo,
		Signature: signature,
		Body:      bodyCode,
		Code:      fullCode,
		Doc:       doc,
	}, nil
}

// parseConstants extracts constant declarations
func parseConstants(d *ast.GenDecl, fset *token.FileSet, src []byte) ([]*model.GoConstant, error) {
	var constants []*model.GoConstant

	// We don't compute source positions since we're not extracting the full block code currently
	// start := d.Pos()
	// if d.Doc != nil {
	//	start = d.Doc.Pos()
	// }
	// We don't compute offsets since we're not using them currently
	// end := d.End()
	// startOff := fset.Position(start).Offset
	// endOff := fset.Position(end).Offset
	// We don't use blockCode here, but we keep the extraction logic for future use
	// blockCode := ""
	// if startOff >= 0 && endOff <= len(src) && startOff < endOff {
	// 	blockCode = string(src[startOff:endOff])
	// }

	// Process each const spec
	for si, spec := range d.Specs {
		valSpec := spec.(*ast.ValueSpec)
		for vi, nameIdent := range valSpec.Names {
			if nameIdent.Name == "_" {
				continue // ignore blank name
			}
			name := nameIdent.Name
			typStr, valStr := "", ""
			if valSpec.Type != nil {
				var buf bytes.Buffer
				_ = format.Node(&buf, fset, valSpec.Type)
				typStr = buf.String()
			}
			if len(valSpec.Values) > 0 {
				if vi < len(valSpec.Values) {
					var buf bytes.Buffer
					_ = format.Node(&buf, fset, valSpec.Values[vi])
					valStr = buf.String()
				} else {
					// If values are fewer (e.g., iota continuation), leave valStr empty
					valStr = ""
				}
			}
			// Doc comment for this spec
			doc := ""
			if vi == 0 { // attach spec's doc only once
				if valSpec.Doc != nil {
					doc = strings.TrimSpace(valSpec.Doc.Text())
				} else if d.Doc != nil && si == 0 {
					// Use GenDecl doc for the first spec if spec lacks its own doc
					doc = strings.TrimSpace(d.Doc.Text())
				}
			}
			// Trailing comment
			comment := ""
			if valSpec.Comment != nil {
				comment = strings.TrimSpace(valSpec.Comment.Text())
			}

			constants = append(constants, &model.GoConstant{
				Name:    name,
				Type:    typStr,
				Value:   valStr,
				Doc:     doc,
				Comment: comment,
			})
		}
	}

	return constants, nil
}

// parseVariables extracts variable declarations
func parseVariables(d *ast.GenDecl, fset *token.FileSet, src []byte) ([]*model.GoVariable, error) {
	var variables []*model.GoVariable

	// We don't compute source positions since we're not extracting the full block code currently
	// start := d.Pos()
	// if d.Doc != nil {
	//	start = d.Doc.Pos()
	// }
	// We don't compute offsets since we're not using them currently
	// end := d.End()
	// startOff := fset.Position(start).Offset
	// endOff := fset.Position(end).Offset
	// We don't use blockCode here, but we keep the extraction logic for future use
	// blockCode := ""
	// if startOff >= 0 && endOff <= len(src) && startOff < endOff {
	// 	blockCode = string(src[startOff:endOff])
	// }

	// Process each var spec
	for si, spec := range d.Specs {
		valSpec := spec.(*ast.ValueSpec)
		for vi, nameIdent := range valSpec.Names {
			if nameIdent.Name == "_" {
				continue // ignore blank name
			}
			name := nameIdent.Name
			typStr, valStr := "", ""
			if valSpec.Type != nil {
				var buf bytes.Buffer
				_ = format.Node(&buf, fset, valSpec.Type)
				typStr = buf.String()
			}
			if len(valSpec.Values) > 0 {
				if vi < len(valSpec.Values) {
					var buf bytes.Buffer
					_ = format.Node(&buf, fset, valSpec.Values[vi])
					valStr = buf.String()
				} else {
					valStr = ""
				}
			}
			// Doc comment for this spec
			doc := ""
			if vi == 0 { // attach spec's doc only once
				if valSpec.Doc != nil {
					doc = strings.TrimSpace(valSpec.Doc.Text())
				} else if d.Doc != nil && si == 0 {
					// Use GenDecl doc for the first spec if spec lacks its own doc
					doc = strings.TrimSpace(d.Doc.Text())
				}
			}
			// Trailing comment
			comment := ""
			if valSpec.Comment != nil {
				comment = strings.TrimSpace(valSpec.Comment.Text())
			}

			variables = append(variables, &model.GoVariable{
				Name:    name,
				Type:    typStr,
				Value:   valStr,
				Doc:     doc,
				Comment: comment,
			})
		}
	}

	return variables, nil
}

// parseTypes extracts type declarations
func parseTypes(d *ast.GenDecl, fset *token.FileSet, src []byte) ([]*model.GoType, error) {
	var types []*model.GoType

	// Full block source
	start := d.Pos()
	if d.Doc != nil {
		start = d.Doc.Pos()
	}
	end := d.End()
	startOff := fset.Position(start).Offset
	endOff := fset.Position(end).Offset
	blockCode := ""
	if startOff >= 0 && endOff <= len(src) && startOff < endOff {
		blockCode = string(src[startOff:endOff])
	}

	// Process each TypeSpec
	for si, spec := range d.Specs {
		typeSpec := spec.(*ast.TypeSpec)
		typeName := typeSpec.Name.Name

		// Documentation
		doc := ""
		if typeSpec.Doc != nil {
			doc = strings.TrimSpace(typeSpec.Doc.Text())
		} else if d.Doc != nil && si == 0 {
			doc = strings.TrimSpace(d.Doc.Text())
		}

		kind := "type"
		aliasOf := ""
		underlying := ""
		var fields []model.GoField
		var methods []model.GoMethod

		if typeSpec.Assign.IsValid() {
			// Type alias
			kind = "alias"
			var buf bytes.Buffer
			_ = format.Node(&buf, fset, typeSpec.Type)
			aliasOf = buf.String()
		} else {
			// New type definition
			switch t := typeSpec.Type.(type) {
			case *ast.StructType:
				kind = "struct"
				underlying = "struct"
				// Iterate struct fields
				for _, field := range t.Fields.List {
					if len(field.Names) == 0 {
						// Embedded field (anonymous)
						var buf bytes.Buffer
						_ = format.Node(&buf, fset, field.Type)
						ftype := buf.String()
						tag := ""
						if field.Tag != nil {
							tag = field.Tag.Value
						}
						docField := ""
						if field.Doc != nil {
							docField = strings.TrimSpace(field.Doc.Text())
						}
						comment := ""
						if field.Comment != nil {
							comment = strings.TrimSpace(field.Comment.Text())
						}
						fields = append(fields, model.GoField{
							Name: "", Type: ftype, Tag: tag, Doc: docField, Comment: comment,
						})
					} else {
						// Named field(s)
						var buf bytes.Buffer
						_ = format.Node(&buf, fset, field.Type)
						ftype := buf.String()
						tag := ""
						if field.Tag != nil {
							tag = field.Tag.Value
						}
						docField := ""
						if field.Doc != nil {
							docField = strings.TrimSpace(field.Doc.Text())
						}
						comment := ""
						if field.Comment != nil {
							comment = strings.TrimSpace(field.Comment.Text())
						}
						for i, nameIdent := range field.Names {
							name := nameIdent.Name
							// Only the first name in a multi-name field carries the Doc comment
							docText := docField
							if i > 0 {
								docText = ""
							}
							fields = append(fields, model.GoField{
								Name: name, Type: ftype, Tag: tag, Doc: docText, Comment: comment,
							})
						}
					}
				}
			case *ast.InterfaceType:
				kind = "interface"
				underlying = "interface"
				// Iterate interface methods
				for _, field := range t.Methods.List {
					if len(field.Names) == 0 {
						// Embedded interface
						var buf bytes.Buffer
						_ = format.Node(&buf, fset, field.Type)
						embedIface := buf.String()
						docField := ""
						if field.Doc != nil {
							docField = strings.TrimSpace(field.Doc.Text())
						}
						comment := ""
						if field.Comment != nil {
							comment = strings.TrimSpace(field.Comment.Text())
						}
						methods = append(methods, model.GoMethod{
							Name: embedIface, Signature: "", Doc: docField, Comment: comment,
						})
					} else {
						// Regular method
						methodName := field.Names[0].Name
						sig := ""
						if funcType, ok := field.Type.(*ast.FuncType); ok {
							var buf bytes.Buffer
							_ = format.Node(&buf, fset, funcType)
							sig = buf.String()
						}
						docField := ""
						if field.Doc != nil {
							docField = strings.TrimSpace(field.Doc.Text())
						}
						comment := ""
						if field.Comment != nil {
							comment = strings.TrimSpace(field.Comment.Text())
						}
						methods = append(methods, model.GoMethod{
							Name: methodName, Signature: sig, Doc: docField, Comment: comment,
						})
					}
				}
			default:
				// Some other type (e.g., alias without "=", or custom type like int, func, etc.)
				var buf bytes.Buffer
				_ = format.Node(&buf, fset, typeSpec.Type)
				underlying = buf.String()
			}
		}

		types = append(types, &model.GoType{
			Name:             typeName,
			Kind:             kind,
			AliasOf:          aliasOf,
			UnderlyingType:   underlying,
			Fields:           fields,
			InterfaceMethods: methods,
			Code:             blockCode,
			Doc:              doc,
		})
	}

	return types, nil
}

// extractPackageDocs extracts package-level documentation and license headers
func extractPackageDocs(pkg *model.GoPackage, files *fileSet, fset *token.FileSet, fileContents map[string][]byte) {
	// Sort filenames to have a deterministic order
	var filenames []string
	for fname := range files.Files {
		filenames = append(filenames, fname)
	}
	sort.Strings(filenames)

	// Capture package-level comments from the first file
	if len(filenames) > 0 {
		firstFile := files.Files[filenames[0]]

		// Package documentation comment
		if firstFile.Doc != nil {
			pkg.PackageDoc = strings.TrimSpace(firstFile.Doc.Text())
		}

		// License/header comments before the package clause
		pkgPos := firstFile.Package // position of "package" keyword
		for _, cg := range firstFile.Comments {
			if cg.Pos() < pkgPos {
				// Skip if this is the same group as package doc (already taken)
				if firstFile.Doc != nil && cg == firstFile.Doc {
					continue
				}
				pkg.LicenseHeader += cg.Text()
				if !strings.HasSuffix(pkg.LicenseHeader, "\n") {
					pkg.LicenseHeader += "\n"
				}
			} else {
				break // stop at the first comment not before the package
			}
		}
	}
}
