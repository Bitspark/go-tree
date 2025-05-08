package formatter

import (
	"errors"
	"testing"

	"bitspark.dev/go-tree/pkg/core/module"
)

// MockVisitor implements FormatVisitor for testing
type MockVisitor struct {
	VisitedModule    bool
	VisitedPackage   bool
	VisitedFile      bool
	VisitedTypes     int
	VisitedFunctions int
	VisitedMethods   int
	VisitedFields    int
	VisitedConstants int
	VisitedVariables int
	VisitedImports   int
	ResultString     string
	ShouldFail       bool
}

func (m *MockVisitor) VisitModule(mod *module.Module) error {
	if m.ShouldFail {
		return errors.New("mock module visit failure")
	}
	m.VisitedModule = true
	return nil
}

func (m *MockVisitor) VisitPackage(pkg *module.Package) error {
	if m.ShouldFail {
		return errors.New("mock package visit failure")
	}
	m.VisitedPackage = true
	return nil
}

func (m *MockVisitor) VisitFile(file *module.File) error {
	if m.ShouldFail {
		return errors.New("mock file visit failure")
	}
	m.VisitedFile = true
	return nil
}

func (m *MockVisitor) VisitType(typ *module.Type) error {
	if m.ShouldFail {
		return errors.New("mock type visit failure")
	}
	m.VisitedTypes++
	return nil
}

func (m *MockVisitor) VisitFunction(fn *module.Function) error {
	if m.ShouldFail {
		return errors.New("mock function visit failure")
	}
	m.VisitedFunctions++
	return nil
}

func (m *MockVisitor) VisitMethod(method *module.Method) error {
	if m.ShouldFail {
		return errors.New("mock method visit failure")
	}
	m.VisitedMethods++
	return nil
}

func (m *MockVisitor) VisitField(field *module.Field) error {
	if m.ShouldFail {
		return errors.New("mock field visit failure")
	}
	m.VisitedFields++
	return nil
}

func (m *MockVisitor) VisitConstant(c *module.Constant) error {
	if m.ShouldFail {
		return errors.New("mock constant visit failure")
	}
	m.VisitedConstants++
	return nil
}

func (m *MockVisitor) VisitVariable(v *module.Variable) error {
	if m.ShouldFail {
		return errors.New("mock variable visit failure")
	}
	m.VisitedVariables++
	return nil
}

func (m *MockVisitor) VisitImport(imp *module.Import) error {
	if m.ShouldFail {
		return errors.New("mock import visit failure")
	}
	m.VisitedImports++
	return nil
}

func (m *MockVisitor) Result() (string, error) {
	if m.ShouldFail {
		return "", errors.New("mock result failure")
	}
	return m.ResultString, nil
}

// TestBaseFormatterVisitsAllElements tests that the BaseFormatter visits all elements
// in a module and calls the appropriate visitor methods
func TestBaseFormatterVisitsAllElements(t *testing.T) {
	// Create a test module with various elements
	mod := module.NewModule("testmodule", "")

	// Create a package in the module
	pkg := &module.Package{
		Name:       "testpackage",
		ImportPath: "testpackage",
		Files:      make(map[string]*module.File),
		Types:      make(map[string]*module.Type),
		Functions:  make(map[string]*module.Function),
		Constants:  make(map[string]*module.Constant),
		Variables:  make(map[string]*module.Variable),
	}
	mod.AddPackage(pkg)

	// Create a file
	file := &module.File{
		Name:    "testfile.go",
		Path:    "testpackage/testfile.go",
		Package: pkg,
		Imports: []*module.Import{
			{Path: "fmt"},
			{Path: "os"},
		},
	}
	pkg.Files["testfile.go"] = file

	// Create types
	type1 := module.NewType("TestType1", "struct", true)
	type2 := module.NewType("TestType2", "interface", true)
	pkg.Types["TestType1"] = type1
	pkg.Types["TestType2"] = type2

	// Add fields to struct
	type1.AddField("Field1", "string", "", false, "Field1 documentation")
	type1.AddField("Field2", "int", "", false, "Field2 documentation")

	// Add methods to types
	type1.AddMethod("Method1", "() error", false, "Method1 documentation")
	type2.AddInterfaceMethod("Method2", "() string", false, "Method2 documentation")

	// Create functions
	func1 := module.NewFunction("TestFunc1", true, false)
	func2 := module.NewFunction("TestFunc2", true, false)
	func3 := module.NewFunction("TestFunc3", true, false)
	pkg.Functions["TestFunc1"] = func1
	pkg.Functions["TestFunc2"] = func2
	pkg.Functions["TestFunc3"] = func3

	// Create constants
	const1 := &module.Constant{Name: "Const1", Value: "1", Type: "int", IsExported: true}
	const2 := &module.Constant{Name: "Const2", Value: "2", Type: "int", IsExported: true}
	pkg.Constants["Const1"] = const1
	pkg.Constants["Const2"] = const2

	// Create variables
	var1 := &module.Variable{Name: "Var1", Type: "string", IsExported: true}
	pkg.Variables["Var1"] = var1

	// Create a mock visitor
	mockVisitor := &MockVisitor{
		ResultString: "test result",
	}

	// Create a formatter with the mock visitor
	formatter := NewBaseFormatter(mockVisitor)

	// Format the module
	result, err := formatter.Format(mod)

	// Check that there was no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that the result is correct
	if result != "test result" {
		t.Errorf("Expected result 'test result', got: %s", result)
	}

	// Check that all elements were visited
	if !mockVisitor.VisitedModule {
		t.Error("Module was not visited")
	}

	if !mockVisitor.VisitedPackage {
		t.Error("Package was not visited")
	}

	if !mockVisitor.VisitedFile {
		t.Error("File was not visited")
	}

	if mockVisitor.VisitedTypes != 2 {
		t.Errorf("Expected 2 types to be visited, got: %d", mockVisitor.VisitedTypes)
	}

	if mockVisitor.VisitedFunctions != 3 {
		t.Errorf("Expected 3 functions to be visited, got: %d", mockVisitor.VisitedFunctions)
	}

	if mockVisitor.VisitedConstants != 2 {
		t.Errorf("Expected 2 constants to be visited, got: %d", mockVisitor.VisitedConstants)
	}

	if mockVisitor.VisitedVariables != 1 {
		t.Errorf("Expected 1 variable to be visited, got: %d", mockVisitor.VisitedVariables)
	}

	if mockVisitor.VisitedImports != 2 {
		t.Errorf("Expected 2 imports to be visited, got: %d", mockVisitor.VisitedImports)
	}

	// Fields and methods should also be visited
	if mockVisitor.VisitedFields != 2 {
		t.Errorf("Expected 2 fields to be visited, got: %d", mockVisitor.VisitedFields)
	}

	if mockVisitor.VisitedMethods != 1 {
		t.Errorf("Expected 1 method to be visited, got: %d", mockVisitor.VisitedMethods)
	}
}

// TestBaseFormatterErrorHandling tests that the BaseFormatter correctly handles errors
// from the visitor methods
func TestBaseFormatterErrorHandling(t *testing.T) {
	// Create a test module
	mod := module.NewModule("testmodule", "")

	// Create a package in the module
	pkg := &module.Package{
		Name:       "testpackage",
		ImportPath: "testpackage",
		Files:      make(map[string]*module.File),
		Types:      make(map[string]*module.Type),
		Functions:  make(map[string]*module.Function),
	}
	mod.AddPackage(pkg)

	// Add a file with an import
	file := &module.File{
		Name:    "testfile.go",
		Path:    "testpackage/testfile.go",
		Package: pkg,
		Imports: []*module.Import{
			{Path: "fmt"},
		},
	}
	pkg.Files["testfile.go"] = file

	// Add a type
	typ := module.NewType("TestType1", "struct", true)
	pkg.Types["TestType1"] = typ

	// Add a function
	fn := module.NewFunction("TestFunc1", true, false)
	pkg.Functions["TestFunc1"] = fn

	testCases := []struct {
		name        string
		visitor     *MockVisitor
		expectError bool
	}{
		{
			name: "VisitModule fails",
			visitor: &MockVisitor{
				ShouldFail: true,
			},
			expectError: true,
		},
		{
			name: "Everything succeeds",
			visitor: &MockVisitor{
				ShouldFail:   false,
				ResultString: "success",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewBaseFormatter(tc.visitor)
			_, err := formatter.Format(mod)

			if tc.expectError && err == nil {
				t.Error("Expected an error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
