package typesys

import (
	"testing"
)

// MockVisitor implements TypeSystemVisitor and tracks which methods were called
type MockVisitor struct {
	BaseVisitor
	Called map[string]int
}

func NewMockVisitor() *MockVisitor {
	return &MockVisitor{
		Called: make(map[string]int),
	}
}

// Override each visitor method to track calls
func (v *MockVisitor) VisitModule(mod *Module) error {
	v.Called["VisitModule"]++
	return nil
}

func (v *MockVisitor) VisitPackage(pkg *Package) error {
	v.Called["VisitPackage"]++
	return nil
}

func (v *MockVisitor) VisitFile(file *File) error {
	v.Called["VisitFile"]++
	return nil
}

func (v *MockVisitor) VisitSymbol(sym *Symbol) error {
	v.Called["VisitSymbol"]++
	return nil
}

func (v *MockVisitor) VisitType(typ *Symbol) error {
	v.Called["VisitType"]++
	return nil
}

func (v *MockVisitor) VisitFunction(fn *Symbol) error {
	v.Called["VisitFunction"]++
	return nil
}

func (v *MockVisitor) VisitVariable(vr *Symbol) error {
	v.Called["VisitVariable"]++
	return nil
}

func (v *MockVisitor) VisitConstant(c *Symbol) error {
	v.Called["VisitConstant"]++
	return nil
}

func (v *MockVisitor) VisitField(f *Symbol) error {
	v.Called["VisitField"]++
	return nil
}

func (v *MockVisitor) VisitMethod(m *Symbol) error {
	v.Called["VisitMethod"]++
	return nil
}

func (v *MockVisitor) VisitParameter(p *Symbol) error {
	v.Called["VisitParameter"]++
	return nil
}

func (v *MockVisitor) VisitImport(i *Import) error {
	v.Called["VisitImport"]++
	return nil
}

func (v *MockVisitor) VisitInterface(i *Symbol) error {
	v.Called["VisitInterface"]++
	return nil
}

func (v *MockVisitor) VisitStruct(s *Symbol) error {
	v.Called["VisitStruct"]++
	return nil
}

func (v *MockVisitor) VisitGenericType(g *Symbol) error {
	v.Called["VisitGenericType"]++
	return nil
}

func (v *MockVisitor) VisitTypeParameter(p *Symbol) error {
	v.Called["VisitTypeParameter"]++
	return nil
}

func (v *MockVisitor) AfterVisitModule(mod *Module) error {
	v.Called["AfterVisitModule"]++
	return nil
}

func (v *MockVisitor) AfterVisitPackage(pkg *Package) error {
	v.Called["AfterVisitPackage"]++
	return nil
}

func TestBaseVisitor(t *testing.T) {
	visitor := &BaseVisitor{}

	// Test that all methods return nil (no errors)
	if err := visitor.VisitModule(nil); err != nil {
		t.Errorf("BaseVisitor.VisitModule returned error: %v", err)
	}
	if err := visitor.VisitPackage(nil); err != nil {
		t.Errorf("BaseVisitor.VisitPackage returned error: %v", err)
	}
	if err := visitor.VisitFile(nil); err != nil {
		t.Errorf("BaseVisitor.VisitFile returned error: %v", err)
	}
	if err := visitor.VisitSymbol(nil); err != nil {
		t.Errorf("BaseVisitor.VisitSymbol returned error: %v", err)
	}
	if err := visitor.VisitType(nil); err != nil {
		t.Errorf("BaseVisitor.VisitType returned error: %v", err)
	}
	if err := visitor.VisitFunction(nil); err != nil {
		t.Errorf("BaseVisitor.VisitFunction returned error: %v", err)
	}
	// Not testing all methods for brevity
}

func TestWalk(t *testing.T) {
	// Create test module with packages, files, and symbols
	module := NewModule("/test/module")

	// Create a package
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	module.Packages[pkg.ImportPath] = pkg

	// Create a file
	file := NewFile("/test/module/file.go", pkg)
	pkg.AddFile(file)

	// Add an import
	imp := &Import{
		Path: "github.com/other/pkg",
		Name: "other",
		File: file,
	}
	file.Imports = append(file.Imports, imp)

	// Add symbols of different kinds
	funcSym := NewSymbol("TestFunction", KindFunction)
	varSym := NewSymbol("TestVariable", KindVariable)
	typeSym := NewSymbol("TestType", KindType)
	structSym := NewSymbol("TestStruct", KindStruct)
	interfaceSym := NewSymbol("TestInterface", KindInterface)
	methodSym := NewSymbol("TestMethod", KindMethod)

	// Set the file for each symbol
	funcSym.File = file
	varSym.File = file
	typeSym.File = file
	structSym.File = file
	interfaceSym.File = file
	methodSym.File = file

	// Add symbols to file
	file.Symbols = append(file.Symbols, funcSym, varSym, typeSym, structSym, interfaceSym, methodSym)

	// Create the mock visitor
	visitor := NewMockVisitor()

	// Walk the module
	err := Walk(visitor, module)
	if err != nil {
		t.Errorf("Walk returned error: %v", err)
	}

	// Verify that the expected methods were called
	if visitor.Called["VisitModule"] != 1 {
		t.Errorf("VisitModule called %d times, want 1", visitor.Called["VisitModule"])
	}
	if visitor.Called["VisitPackage"] != 1 {
		t.Errorf("VisitPackage called %d times, want 1", visitor.Called["VisitPackage"])
	}
	if visitor.Called["VisitFile"] != 1 {
		t.Errorf("VisitFile called %d times, want 1", visitor.Called["VisitFile"])
	}
	if visitor.Called["VisitImport"] != 1 {
		t.Errorf("VisitImport called %d times, want 1", visitor.Called["VisitImport"])
	}

	// All symbols should be visited
	expectedSymbolVisits := 6 // one for each symbol
	if visitor.Called["VisitSymbol"] != expectedSymbolVisits {
		t.Errorf("VisitSymbol called %d times, want %d", visitor.Called["VisitSymbol"], expectedSymbolVisits)
	}

	// Each specific kind should be visited
	if visitor.Called["VisitFunction"] != 1 {
		t.Errorf("VisitFunction called %d times, want 1", visitor.Called["VisitFunction"])
	}
	if visitor.Called["VisitVariable"] != 1 {
		t.Errorf("VisitVariable called %d times, want 1", visitor.Called["VisitVariable"])
	}
	if visitor.Called["VisitType"] != 1 {
		t.Errorf("VisitType called %d times, want 1", visitor.Called["VisitType"])
	}
	if visitor.Called["VisitStruct"] != 1 {
		t.Errorf("VisitStruct called %d times, want 1", visitor.Called["VisitStruct"])
	}
	if visitor.Called["VisitInterface"] != 1 {
		t.Errorf("VisitInterface called %d times, want 1", visitor.Called["VisitInterface"])
	}
	if visitor.Called["VisitMethod"] != 1 {
		t.Errorf("VisitMethod called %d times, want 1", visitor.Called["VisitMethod"])
	}
}

func TestFilteredVisitor(t *testing.T) {
	// Create test module with packages, files, and symbols
	module := NewModule("/test/module")

	// Create a package
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	module.Packages[pkg.ImportPath] = pkg

	// Create a file
	file := NewFile("/test/module/file.go", pkg)
	pkg.AddFile(file)

	// Add symbols of different kinds and export status
	exportedFunc := NewSymbol("ExportedFunc", KindFunction)
	unexportedFunc := NewSymbol("unexportedFunc", KindFunction)
	exportedVar := NewSymbol("ExportedVar", KindVariable)
	unexportedVar := NewSymbol("unexportedVar", KindVariable)

	// Set the file for each symbol
	exportedFunc.File = file
	unexportedFunc.File = file
	exportedVar.File = file
	unexportedVar.File = file

	// Add symbols to file
	file.Symbols = append(file.Symbols, exportedFunc, unexportedFunc, exportedVar, unexportedVar)

	// Create the mock visitor and filtered visitor
	mockVisitor := NewMockVisitor()
	filteredVisitor := &FilteredVisitor{
		Visitor: mockVisitor,
		Filter:  ExportedFilter(), // Only visit exported symbols
	}

	// Walk the module with the filtered visitor
	err := Walk(filteredVisitor, module)
	if err != nil {
		t.Errorf("Walk returned error: %v", err)
	}

	// Module, package, and file should be visited
	if mockVisitor.Called["VisitModule"] != 1 {
		t.Errorf("VisitModule called %d times, want 1", mockVisitor.Called["VisitModule"])
	}
	if mockVisitor.Called["VisitPackage"] != 1 {
		t.Errorf("VisitPackage called %d times, want 1", mockVisitor.Called["VisitPackage"])
	}
	if mockVisitor.Called["VisitFile"] != 1 {
		t.Errorf("VisitFile called %d times, want 1", mockVisitor.Called["VisitFile"])
	}

	// Only exported symbols should be visited
	expectedSymbolVisits := 2 // ExportedFunc and ExportedVar
	if mockVisitor.Called["VisitSymbol"] != expectedSymbolVisits {
		t.Errorf("VisitSymbol called %d times, want %d", mockVisitor.Called["VisitSymbol"], expectedSymbolVisits)
	}

	// Only one function should be visited (the exported one)
	if mockVisitor.Called["VisitFunction"] != 1 {
		t.Errorf("VisitFunction called %d times, want 1", mockVisitor.Called["VisitFunction"])
	}

	// Only one variable should be visited (the exported one)
	if mockVisitor.Called["VisitVariable"] != 1 {
		t.Errorf("VisitVariable called %d times, want 1", mockVisitor.Called["VisitVariable"])
	}

	// Test other filter types
	kindFilterVisitor := &FilteredVisitor{
		Visitor: NewMockVisitor(),
		Filter:  KindFilter(KindFunction), // Only visit functions
	}

	err = Walk(kindFilterVisitor, module)
	if err != nil {
		t.Errorf("Walk with KindFilter returned error: %v", err)
	}

	// Only functions should be visited (both exported and unexported)
	if kindFilterVisitor.Visitor.(*MockVisitor).Called["VisitFunction"] != 2 {
		t.Errorf("VisitFunction with KindFilter called %d times, want 2",
			kindFilterVisitor.Visitor.(*MockVisitor).Called["VisitFunction"])
	}

	// Variables should not be visited
	if kindFilterVisitor.Visitor.(*MockVisitor).Called["VisitVariable"] != 0 {
		t.Errorf("VisitVariable with KindFilter called %d times, want 0",
			kindFilterVisitor.Visitor.(*MockVisitor).Called["VisitVariable"])
	}

	// Test FileFilter
	fileFilterVisitor := &FilteredVisitor{
		Visitor: NewMockVisitor(),
		Filter:  FileFilter(file), // Only visit symbols in this file
	}

	err = Walk(fileFilterVisitor, module)
	if err != nil {
		t.Errorf("Walk with FileFilter returned error: %v", err)
	}

	// All symbols should be visited since they're all in the same file
	if fileFilterVisitor.Visitor.(*MockVisitor).Called["VisitSymbol"] != 4 {
		t.Errorf("VisitSymbol with FileFilter called %d times, want 4",
			fileFilterVisitor.Visitor.(*MockVisitor).Called["VisitSymbol"])
	}
}
