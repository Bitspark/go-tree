package interfaceanalysis

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/module"
)

// TestAnalyzeReceivers tests the core receiver analysis functionality
func TestAnalyzeReceivers(t *testing.T) {
	// Create a test package with methods
	pkg := createTestPackage()

	// Create analyzer and analyze the package
	analyzer := NewAnalyzer()
	analysis := analyzer.AnalyzeReceivers(pkg)

	// Check package name
	if analysis.Package != "testpackage" {
		t.Errorf("Expected package name 'testpackage', got '%s'", analysis.Package)
	}

	// Check receiver groups
	if len(analysis.Groups) != 3 {
		t.Errorf("Expected 3 receiver groups, got %d", len(analysis.Groups))
	}

	// Check specific groups
	userGroup, ok := analysis.Groups["*User"]
	if !ok {
		t.Fatal("Expected to find *User receiver group")
	}

	if userGroup.BaseType != "User" {
		t.Errorf("Expected User base type, got '%s'", userGroup.BaseType)
	}

	if !userGroup.IsPointer {
		t.Error("Expected *User to be recognized as pointer receiver")
	}

	if len(userGroup.Methods) != 2 {
		t.Errorf("Expected 2 methods for *User, got %d", len(userGroup.Methods))
	}

	// Check auth group
	authGroup, ok := analysis.Groups["Auth"]
	if !ok {
		t.Fatal("Expected to find Auth receiver group")
	}

	if authGroup.IsPointer {
		t.Error("Expected Auth to be recognized as value receiver")
	}

	if len(authGroup.Methods) != 1 {
		t.Errorf("Expected 1 method for Auth, got %d", len(authGroup.Methods))
	}
}

// TestCreateSummary tests the summary creation functionality
func TestCreateSummary(t *testing.T) {
	pkg := createTestPackage()
	analyzer := NewAnalyzer()

	analysis := analyzer.AnalyzeReceivers(pkg)
	summary := analyzer.CreateSummary(analysis)

	// Check summary values
	if summary.TotalMethods != 4 {
		t.Errorf("Expected 4 total methods, got %d", summary.TotalMethods)
	}

	if summary.TotalReceiverTypes != 3 {
		t.Errorf("Expected 3 receiver types, got %d", summary.TotalReceiverTypes)
	}

	if summary.PointerReceivers != 3 {
		t.Errorf("Expected 3 pointer receivers, got %d", summary.PointerReceivers)
	}

	if summary.ValueReceivers != 1 {
		t.Errorf("Expected 1 value receiver, got %d", summary.ValueReceivers)
	}

	// Check method counts per type
	if count, ok := summary.MethodsPerType["*User"]; !ok || count != 2 {
		t.Errorf("Expected 2 methods for *User, got %d", count)
	}

	if count, ok := summary.MethodsPerType["Auth"]; !ok || count != 1 {
		t.Errorf("Expected 1 method for Auth, got %d", count)
	}
}

// TestGroupMethodsByBaseType tests grouping methods by their base type
func TestGroupMethodsByBaseType(t *testing.T) {
	pkg := createTestPackage()
	analyzer := NewAnalyzer()

	analysis := analyzer.AnalyzeReceivers(pkg)
	baseGroups := analyzer.GroupMethodsByBaseType(analysis)

	// Check User base type
	userMethods, ok := baseGroups["User"]
	if !ok {
		t.Fatal("Expected to find User base type group")
	}

	if len(userMethods) != 2 {
		t.Errorf("Expected 2 methods for User base type, got %d", len(userMethods))
	}

	// Check Auth base type
	authMethods, ok := baseGroups["Auth"]
	if !ok {
		t.Fatal("Expected to find Auth base type group")
	}

	if len(authMethods) != 1 {
		t.Errorf("Expected 1 method for Auth base type, got %d", len(authMethods))
	}

	// Check Request base type
	requestMethods, ok := baseGroups["Request"]
	if !ok {
		t.Fatal("Expected to find Request base type group")
	}

	if len(requestMethods) != 1 {
		t.Errorf("Expected 1 method for Request base type, got %d", len(requestMethods))
	}
}

// TestFindCommonMethods tests finding methods with the same name across different receiver types
func TestFindCommonMethods(t *testing.T) {
	// Create a test package with common method names
	userProcessFn := &module.Function{
		Name:     "Process",
		Receiver: &module.Receiver{Type: "*User"},
	}
	requestProcessFn := &module.Function{
		Name:     "Process", // Same name as User.Process
		Receiver: &module.Receiver{Type: "*Request"},
	}
	userValidateFn := &module.Function{
		Name:     "Validate",
		Receiver: &module.Receiver{Type: "*User"},
	}
	requestValidateFn := &module.Function{
		Name:     "Validate", // Same name as User.Validate
		Receiver: &module.Receiver{Type: "*Request"},
	}
	authValidateFn := &module.Function{
		Name:     "Validate", // Same name as Request.Validate and User.Validate
		Receiver: &module.Receiver{Type: "Auth"},
	}
	authUniqueFn := &module.Function{
		Name:     "Unique",
		Receiver: &module.Receiver{Type: "Auth"},
	}

	pkg := &module.Package{
		Name: "testpackage",
		Functions: map[string]*module.Function{
			"User.Process":     userProcessFn,
			"Request.Process":  requestProcessFn,
			"User.Validate":    userValidateFn,
			"Request.Validate": requestValidateFn,
			"Auth.Validate":    authValidateFn,
			"Auth.Unique":      authUniqueFn,
		},
	}

	analyzer := NewAnalyzer()
	analysis := analyzer.AnalyzeReceivers(pkg)
	commonMethods := analyzer.FindCommonMethods(analysis)

	// Check common methods
	if len(commonMethods) != 2 {
		t.Errorf("Expected 2 common method names, got %d", len(commonMethods))
	}

	// Check "Process" method
	process, ok := commonMethods["Process"]
	if !ok {
		t.Fatal("Expected to find Process in common methods")
	}

	if len(process) != 2 {
		t.Errorf("Expected Process to be implemented by 2 types, got %d", len(process))
	}

	// Check "Validate" method
	validate, ok := commonMethods["Validate"]
	if !ok {
		t.Fatal("Expected to find Validate in common methods")
	}

	if len(validate) != 3 {
		t.Errorf("Expected Validate to be implemented by 3 types, got %d", len(validate))
	}
}

// TestGetReceiverMethodSignatures tests getting method signatures for specific receiver types
func TestGetReceiverMethodSignatures(t *testing.T) {
	// Create a test package with method signatures
	loginFn := &module.Function{
		Name:      "Login",
		Signature: "(username, password string) (bool, error)",
		Receiver:  &module.Receiver{Type: "*User"},
	}
	logoutFn := &module.Function{
		Name:      "Logout",
		Signature: "() error",
		Receiver:  &module.Receiver{Type: "*User"},
	}

	pkg := &module.Package{
		Name: "testpackage",
		Functions: map[string]*module.Function{
			"User.Login":  loginFn,
			"User.Logout": logoutFn,
		},
	}

	analyzer := NewAnalyzer()
	analysis := analyzer.AnalyzeReceivers(pkg)
	signatures := analyzer.GetReceiverMethodSignatures(analysis, "*User")

	// Check signatures
	if len(signatures) != 2 {
		t.Errorf("Expected 2 signatures, got %d", len(signatures))
	}

	if sig, ok := signatures["Login"]; !ok || sig != "(username, password string) (bool, error)" {
		t.Errorf("Expected Login signature '(username, password string) (bool, error)', got '%s'", sig)
	}

	if sig, ok := signatures["Logout"]; !ok || sig != "() error" {
		t.Errorf("Expected Logout signature '() error', got '%s'", sig)
	}

	// Check non-existent receiver
	emptySignatures := analyzer.GetReceiverMethodSignatures(analysis, "NonExistent")
	if len(emptySignatures) != 0 {
		t.Errorf("Expected 0 signatures for non-existent receiver, got %d", len(emptySignatures))
	}
}

// createTestPackage creates a test package with methods and receivers for testing
func createTestPackage() *module.Package {
	loginFn := &module.Function{
		Name:     "Login",
		Receiver: &module.Receiver{Type: "*User"},
	}
	logoutFn := &module.Function{
		Name:     "Logout",
		Receiver: &module.Receiver{Type: "*User"},
	}
	validateFn := &module.Function{
		Name:     "Validate",
		Receiver: &module.Receiver{Type: "Auth"},
	}
	processFn := &module.Function{
		Name:     "Process",
		Receiver: &module.Receiver{Type: "*Request"},
	}
	noReceiverFn := &module.Function{
		Name:     "NoReceiver", // Function, not a method
		Receiver: nil,
	}

	return &module.Package{
		Name: "testpackage",
		Functions: map[string]*module.Function{
			"User.Login":      loginFn,
			"User.Logout":     logoutFn,
			"Auth.Validate":   validateFn,
			"Request.Process": processFn,
			"NoReceiver":      noReceiverFn,
		},
	}
}
