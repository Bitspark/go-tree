package formatter

import (
	"errors"
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
)

// MockVisitor implements Visitor for testing
type MockVisitor struct {
	VisitedPackage   bool
	VisitedTypes     int
	VisitedFunctions int
	VisitedConstants int
	VisitedVariables int
	VisitedImports   int
	ResultString     string
	ShouldFail       bool
}

func (m *MockVisitor) VisitPackage(pkg *model.GoPackage) error {
	if m.ShouldFail {
		return errors.New("mock package visit failure")
	}
	m.VisitedPackage = true
	return nil
}

func (m *MockVisitor) VisitType(typ model.GoType) error {
	if m.ShouldFail {
		return errors.New("mock type visit failure")
	}
	m.VisitedTypes++
	return nil
}

func (m *MockVisitor) VisitFunction(fn model.GoFunction) error {
	if m.ShouldFail {
		return errors.New("mock function visit failure")
	}
	m.VisitedFunctions++
	return nil
}

func (m *MockVisitor) VisitConstant(c model.GoConstant) error {
	if m.ShouldFail {
		return errors.New("mock constant visit failure")
	}
	m.VisitedConstants++
	return nil
}

func (m *MockVisitor) VisitVariable(v model.GoVariable) error {
	if m.ShouldFail {
		return errors.New("mock variable visit failure")
	}
	m.VisitedVariables++
	return nil
}

func (m *MockVisitor) VisitImport(imp model.GoImport) error {
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
// in a package and calls the appropriate visitor methods
func TestBaseFormatterVisitsAllElements(t *testing.T) {
	// Create a test package with various elements
	pkg := &model.GoPackage{
		Name: "testpackage",
		Imports: []model.GoImport{
			{Path: "fmt"},
			{Path: "os"},
		},
		Types: []model.GoType{
			{Name: "TestType1", Kind: "struct"},
			{Name: "TestType2", Kind: "interface"},
		},
		Functions: []model.GoFunction{
			{Name: "TestFunc1"},
			{Name: "TestFunc2"},
			{Name: "TestFunc3"},
		},
		Constants: []model.GoConstant{
			{Name: "Const1"},
			{Name: "Const2"},
		},
		Variables: []model.GoVariable{
			{Name: "Var1"},
		},
	}

	// Create a mock visitor
	mockVisitor := &MockVisitor{
		ResultString: "test result",
	}

	// Create a formatter with the mock visitor
	formatter := NewBaseFormatter(mockVisitor)

	// Format the package
	result, err := formatter.Format(pkg)

	// Check that there was no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check that the result is correct
	if result != "test result" {
		t.Errorf("Expected result 'test result', got: %s", result)
	}

	// Check that all elements were visited
	if !mockVisitor.VisitedPackage {
		t.Error("Package was not visited")
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
}

// TestBaseFormatterErrorHandling tests that the BaseFormatter correctly handles errors
// from the visitor methods
func TestBaseFormatterErrorHandling(t *testing.T) {
	// Create a test package with various elements
	pkg := &model.GoPackage{
		Name: "testpackage",
		Imports: []model.GoImport{
			{Path: "fmt"},
		},
		Types: []model.GoType{
			{Name: "TestType1"},
		},
		Functions: []model.GoFunction{
			{Name: "TestFunc1"},
		},
	}

	testCases := []struct {
		name        string
		visitor     *MockVisitor
		expectError bool
	}{
		{
			name: "VisitPackage fails",
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
			_, err := formatter.Format(pkg)

			if tc.expectError && err == nil {
				t.Error("Expected an error but got nil")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
