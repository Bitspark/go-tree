package visitor

import (
	"bitspark.dev/go-tree/pkgold/core/module"
)

// DefaultVisitor provides a no-op implementation of ModuleVisitor
// that can be embedded by other visitors to avoid implementing all methods.
type DefaultVisitor struct{}

// VisitModule provides a default implementation for visiting a module
func (v *DefaultVisitor) VisitModule(mod *module.Module) error {
	return nil
}

// VisitPackage provides a default implementation for visiting a package
func (v *DefaultVisitor) VisitPackage(pkg *module.Package) error {
	return nil
}

// VisitFile provides a default implementation for visiting a file
func (v *DefaultVisitor) VisitFile(file *module.File) error {
	return nil
}

// VisitType provides a default implementation for visiting a type
func (v *DefaultVisitor) VisitType(typ *module.Type) error {
	return nil
}

// VisitFunction provides a default implementation for visiting a function
func (v *DefaultVisitor) VisitFunction(fn *module.Function) error {
	return nil
}

// VisitMethod provides a default implementation for visiting a method
func (v *DefaultVisitor) VisitMethod(method *module.Method) error {
	return nil
}

// VisitField provides a default implementation for visiting a field
func (v *DefaultVisitor) VisitField(field *module.Field) error {
	return nil
}

// VisitVariable provides a default implementation for visiting a variable
func (v *DefaultVisitor) VisitVariable(variable *module.Variable) error {
	return nil
}

// VisitConstant provides a default implementation for visiting a constant
func (v *DefaultVisitor) VisitConstant(constant *module.Constant) error {
	return nil
}

// VisitImport provides a default implementation for visiting an import
func (v *DefaultVisitor) VisitImport(imp *module.Import) error {
	return nil
}
