// Package visitor defines interfaces and implementations for traversing Go modules.
package visitor

import (
	"bitspark.dev/go-tree/pkg/core/module"
)

// ModuleVisitor defines an interface for traversing a module structure
// using the visitor pattern. Each Visit* method is called when
// visiting the corresponding element in the module structure.
type ModuleVisitor interface {
	// VisitModule is called when visiting a module
	VisitModule(mod *module.Module) error

	// VisitPackage is called when visiting a package
	VisitPackage(pkg *module.Package) error

	// VisitFile is called when visiting a file
	VisitFile(file *module.File) error

	// VisitType is called when visiting a type
	VisitType(typ *module.Type) error

	// VisitFunction is called when visiting a function
	VisitFunction(fn *module.Function) error

	// VisitMethod is called when visiting a method
	VisitMethod(method *module.Method) error

	// VisitField is called when visiting a struct field
	VisitField(field *module.Field) error

	// VisitVariable is called when visiting a variable
	VisitVariable(variable *module.Variable) error

	// VisitConstant is called when visiting a constant
	VisitConstant(constant *module.Constant) error

	// VisitImport is called when visiting an import
	VisitImport(imp *module.Import) error
}

// ModuleWalker walks a module and its elements, calling the appropriate
// visitor methods for each element it encounters.
type ModuleWalker struct {
	Visitor ModuleVisitor

	// IncludePrivate determines whether to visit unexported elements
	IncludePrivate bool

	// IncludeTests determines whether to visit test files
	IncludeTests bool

	// IncludeGenerated determines whether to visit generated files
	IncludeGenerated bool
}

// NewModuleWalker creates a new module walker with the given visitor
func NewModuleWalker(visitor ModuleVisitor) *ModuleWalker {
	return &ModuleWalker{
		Visitor:        visitor,
		IncludePrivate: false,
		IncludeTests:   false,
	}
}

// Walk traverses a module structure and calls the appropriate visitor methods
func (w *ModuleWalker) Walk(mod *module.Module) error {
	if err := w.Visitor.VisitModule(mod); err != nil {
		return err
	}

	// Walk through packages
	for _, pkg := range mod.Packages {
		if err := w.walkPackage(pkg); err != nil {
			return err
		}
	}

	return nil
}

// walkPackage traverses a package and its elements
func (w *ModuleWalker) walkPackage(pkg *module.Package) error {
	// Skip test packages if not included
	if pkg.IsTest && !w.IncludeTests {
		return nil
	}

	if err := w.Visitor.VisitPackage(pkg); err != nil {
		return err
	}

	// Walk through files
	for _, file := range pkg.Files {
		if err := w.walkFile(file); err != nil {
			return err
		}
	}

	// Walk through types
	for _, typ := range pkg.Types {
		if !w.IncludePrivate && !typ.IsExported {
			continue
		}
		if err := w.walkType(typ); err != nil {
			return err
		}
	}

	// Walk through functions (not methods, which are processed with types)
	for _, fn := range pkg.Functions {
		if !w.IncludePrivate && !fn.IsExported {
			continue
		}
		if err := w.Visitor.VisitFunction(fn); err != nil {
			return err
		}
	}

	// Walk through variables
	for _, variable := range pkg.Variables {
		if !w.IncludePrivate && !variable.IsExported {
			continue
		}
		if err := w.Visitor.VisitVariable(variable); err != nil {
			return err
		}
	}

	// Walk through constants
	for _, constant := range pkg.Constants {
		if !w.IncludePrivate && !constant.IsExported {
			continue
		}
		if err := w.Visitor.VisitConstant(constant); err != nil {
			return err
		}
	}

	return nil
}

// walkFile traverses a file and its imports
func (w *ModuleWalker) walkFile(file *module.File) error {
	// Skip test files if not included
	if file.IsTest && !w.IncludeTests {
		return nil
	}

	// Skip generated files if not included
	if file.IsGenerated && !w.IncludeGenerated {
		return nil
	}

	if err := w.Visitor.VisitFile(file); err != nil {
		return err
	}

	// Walk through imports
	for _, imp := range file.Imports {
		if err := w.Visitor.VisitImport(imp); err != nil {
			return err
		}
	}

	return nil
}

// walkType traverses a type and its fields/methods
func (w *ModuleWalker) walkType(typ *module.Type) error {
	if err := w.Visitor.VisitType(typ); err != nil {
		return err
	}

	// If it's a struct, walk through fields
	if typ.Kind == "struct" {
		for _, field := range typ.Fields {
			if !w.IncludePrivate && field.Name != "" && !isExported(field.Name) {
				continue
			}
			if err := w.Visitor.VisitField(field); err != nil {
				return err
			}
		}
	}

	// Walk through methods
	for _, method := range typ.Methods {
		// For methods, check if the method name is exported
		if !w.IncludePrivate && !isExported(method.Name) {
			continue
		}
		if err := w.Visitor.VisitMethod(method); err != nil {
			return err
		}
	}

	return nil
}

// isExported checks if a name is exported (starts with uppercase)
func isExported(name string) bool {
	if name == "" {
		return false
	}
	firstChar := name[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}
