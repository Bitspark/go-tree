package typesys

// TypeSystemVisitor provides a type-aware traversal system.
type TypeSystemVisitor interface {
	VisitModule(mod *Module) error
	VisitPackage(pkg *Package) error
	VisitFile(file *File) error
	VisitSymbol(sym *Symbol) error

	// Symbol-specific visitors
	VisitType(typ *Symbol) error
	VisitFunction(fn *Symbol) error
	VisitVariable(v *Symbol) error
	VisitConstant(c *Symbol) error
	VisitField(f *Symbol) error
	VisitMethod(m *Symbol) error
	VisitParameter(p *Symbol) error
	VisitImport(i *Import) error

	// Type-specific visitors
	VisitInterface(i *Symbol) error
	VisitStruct(s *Symbol) error

	// Generic type support
	VisitGenericType(g *Symbol) error
	VisitTypeParameter(p *Symbol) error

	// After visitor methods - called after all children have been visited
	AfterVisitModule(mod *Module) error
	AfterVisitPackage(pkg *Package) error
}

// BaseVisitor provides a default implementation of TypeSystemVisitor.
// All methods return nil, so derived visitors only need to implement
// the methods they care about.
type BaseVisitor struct{}

// VisitModule visits a module.
func (v *BaseVisitor) VisitModule(mod *Module) error {
	return nil
}

// VisitPackage visits a package.
func (v *BaseVisitor) VisitPackage(pkg *Package) error {
	return nil
}

// VisitFile visits a file.
func (v *BaseVisitor) VisitFile(file *File) error {
	return nil
}

// VisitSymbol visits a symbol.
func (v *BaseVisitor) VisitSymbol(sym *Symbol) error {
	return nil
}

// VisitType visits a type.
func (v *BaseVisitor) VisitType(typ *Symbol) error {
	return nil
}

// VisitFunction visits a function.
func (v *BaseVisitor) VisitFunction(fn *Symbol) error {
	return nil
}

// VisitVariable visits a variable.
func (v *BaseVisitor) VisitVariable(vr *Symbol) error {
	return nil
}

// VisitConstant visits a constant.
func (v *BaseVisitor) VisitConstant(c *Symbol) error {
	return nil
}

// VisitField visits a field.
func (v *BaseVisitor) VisitField(f *Symbol) error {
	return nil
}

// VisitMethod visits a method.
func (v *BaseVisitor) VisitMethod(m *Symbol) error {
	return nil
}

// VisitParameter visits a parameter.
func (v *BaseVisitor) VisitParameter(p *Symbol) error {
	return nil
}

// VisitImport visits an import.
func (v *BaseVisitor) VisitImport(i *Import) error {
	return nil
}

// VisitInterface visits an interface.
func (v *BaseVisitor) VisitInterface(i *Symbol) error {
	return nil
}

// VisitStruct visits a struct.
func (v *BaseVisitor) VisitStruct(s *Symbol) error {
	return nil
}

// VisitGenericType visits a generic type.
func (v *BaseVisitor) VisitGenericType(g *Symbol) error {
	return nil
}

// VisitTypeParameter visits a type parameter.
func (v *BaseVisitor) VisitTypeParameter(p *Symbol) error {
	return nil
}

// AfterVisitModule is called after visiting a module and all its packages.
func (v *BaseVisitor) AfterVisitModule(mod *Module) error {
	return nil
}

// AfterVisitPackage is called after visiting a package and all its symbols.
func (v *BaseVisitor) AfterVisitPackage(pkg *Package) error {
	return nil
}

// Walk traverses a module with the visitor.
func Walk(v TypeSystemVisitor, mod *Module) error {
	// Visit the module
	if err := v.VisitModule(mod); err != nil {
		return err
	}

	// Visit each package
	for _, pkg := range mod.Packages {
		if err := walkPackage(v, pkg); err != nil {
			return err
		}
	}

	// Call AfterVisitModule after all packages have been visited
	if err := v.AfterVisitModule(mod); err != nil {
		return err
	}

	return nil
}

// walkPackage traverses a package with the visitor.
func walkPackage(v TypeSystemVisitor, pkg *Package) error {
	// Visit the package
	if err := v.VisitPackage(pkg); err != nil {
		return err
	}

	// Visit each file
	for _, file := range pkg.Files {
		if err := walkFile(v, file); err != nil {
			return err
		}
	}

	// Call AfterVisitPackage after all files have been visited
	if err := v.AfterVisitPackage(pkg); err != nil {
		return err
	}

	return nil
}

// walkFile traverses a file with the visitor.
func walkFile(v TypeSystemVisitor, file *File) error {
	// Visit the file
	if err := v.VisitFile(file); err != nil {
		return err
	}

	// Visit each import
	for _, imp := range file.Imports {
		if err := v.VisitImport(imp); err != nil {
			return err
		}
	}

	// Visit each symbol
	for _, sym := range file.Symbols {
		if err := walkSymbol(v, sym); err != nil {
			return err
		}
	}

	return nil
}

// walkSymbol traverses a symbol with the visitor.
func walkSymbol(v TypeSystemVisitor, sym *Symbol) error {
	// Visit the symbol
	if err := v.VisitSymbol(sym); err != nil {
		return err
	}

	// Dispatch based on symbol kind
	switch sym.Kind {
	case KindType:
		if err := v.VisitType(sym); err != nil {
			return err
		}
	case KindFunction:
		if err := v.VisitFunction(sym); err != nil {
			return err
		}
	case KindMethod:
		if err := v.VisitMethod(sym); err != nil {
			return err
		}
	case KindVariable:
		if err := v.VisitVariable(sym); err != nil {
			return err
		}
	case KindConstant:
		if err := v.VisitConstant(sym); err != nil {
			return err
		}
	case KindField:
		if err := v.VisitField(sym); err != nil {
			return err
		}
	case KindParameter:
		if err := v.VisitParameter(sym); err != nil {
			return err
		}
	case KindInterface:
		if err := v.VisitInterface(sym); err != nil {
			return err
		}
	case KindStruct:
		if err := v.VisitStruct(sym); err != nil {
			return err
		}
	}

	return nil
}

// FilteredVisitor wraps another visitor and filters the symbols that are visited.
type FilteredVisitor struct {
	Visitor TypeSystemVisitor
	Filter  SymbolFilter
}

// SymbolFilter is a function that returns true if a symbol should be visited.
type SymbolFilter func(sym *Symbol) bool

// VisitModule visits a module.
func (v *FilteredVisitor) VisitModule(mod *Module) error {
	return v.Visitor.VisitModule(mod)
}

// VisitPackage visits a package.
func (v *FilteredVisitor) VisitPackage(pkg *Package) error {
	return v.Visitor.VisitPackage(pkg)
}

// VisitFile visits a file.
func (v *FilteredVisitor) VisitFile(file *File) error {
	return v.Visitor.VisitFile(file)
}

// VisitSymbol visits a symbol.
func (v *FilteredVisitor) VisitSymbol(sym *Symbol) error {
	if v.Filter(sym) {
		return v.Visitor.VisitSymbol(sym)
	}
	return nil
}

// VisitType visits a type.
func (v *FilteredVisitor) VisitType(typ *Symbol) error {
	if v.Filter(typ) {
		return v.Visitor.VisitType(typ)
	}
	return nil
}

// VisitFunction visits a function.
func (v *FilteredVisitor) VisitFunction(fn *Symbol) error {
	if v.Filter(fn) {
		return v.Visitor.VisitFunction(fn)
	}
	return nil
}

// VisitVariable visits a variable.
func (v *FilteredVisitor) VisitVariable(vr *Symbol) error {
	if v.Filter(vr) {
		return v.Visitor.VisitVariable(vr)
	}
	return nil
}

// VisitConstant visits a constant.
func (v *FilteredVisitor) VisitConstant(c *Symbol) error {
	if v.Filter(c) {
		return v.Visitor.VisitConstant(c)
	}
	return nil
}

// VisitField visits a field.
func (v *FilteredVisitor) VisitField(f *Symbol) error {
	if v.Filter(f) {
		return v.Visitor.VisitField(f)
	}
	return nil
}

// VisitMethod visits a method.
func (v *FilteredVisitor) VisitMethod(m *Symbol) error {
	if v.Filter(m) {
		return v.Visitor.VisitMethod(m)
	}
	return nil
}

// VisitParameter visits a parameter.
func (v *FilteredVisitor) VisitParameter(p *Symbol) error {
	if v.Filter(p) {
		return v.Visitor.VisitParameter(p)
	}
	return nil
}

// VisitImport visits an import.
func (v *FilteredVisitor) VisitImport(i *Import) error {
	return v.Visitor.VisitImport(i)
}

// VisitInterface visits an interface.
func (v *FilteredVisitor) VisitInterface(i *Symbol) error {
	if v.Filter(i) {
		return v.Visitor.VisitInterface(i)
	}
	return nil
}

// VisitStruct visits a struct.
func (v *FilteredVisitor) VisitStruct(s *Symbol) error {
	if v.Filter(s) {
		return v.Visitor.VisitStruct(s)
	}
	return nil
}

// VisitGenericType visits a generic type.
func (v *FilteredVisitor) VisitGenericType(g *Symbol) error {
	if v.Filter(g) {
		return v.Visitor.VisitGenericType(g)
	}
	return nil
}

// VisitTypeParameter visits a type parameter.
func (v *FilteredVisitor) VisitTypeParameter(p *Symbol) error {
	if v.Filter(p) {
		return v.Visitor.VisitTypeParameter(p)
	}
	return nil
}

// AfterVisitModule visits a module after all its packages.
func (v *FilteredVisitor) AfterVisitModule(mod *Module) error {
	return v.Visitor.AfterVisitModule(mod)
}

// AfterVisitPackage visits a package after all its symbols.
func (v *FilteredVisitor) AfterVisitPackage(pkg *Package) error {
	return v.Visitor.AfterVisitPackage(pkg)
}

// ExportedFilter returns a filter that only visits exported symbols.
func ExportedFilter() SymbolFilter {
	return func(sym *Symbol) bool {
		return sym.Exported
	}
}

// KindFilter returns a filter that only visits symbols of the given kinds.
func KindFilter(kinds ...SymbolKind) SymbolFilter {
	return func(sym *Symbol) bool {
		for _, kind := range kinds {
			if sym.Kind == kind {
				return true
			}
		}
		return false
	}
}

// FileFilter returns a filter that only visits symbols in the given file.
func FileFilter(file *File) SymbolFilter {
	return func(sym *Symbol) bool {
		return sym.File == file
	}
}

// PackageFilter returns a filter that only visits symbols in the given package.
func PackageFilter(pkg *Package) SymbolFilter {
	return func(sym *Symbol) bool {
		return sym.Package == pkg
	}
}
