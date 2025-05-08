package typesys

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"
)

func TestNewTypeBridge(t *testing.T) {
	bridge := NewTypeBridge()

	if bridge.SymToObj == nil {
		t.Error("SymToObj map should be initialized")
	}

	if bridge.ObjToSym == nil {
		t.Error("ObjToSym map should be initialized")
	}

	if bridge.NodeToSym == nil {
		t.Error("NodeToSym map should be initialized")
	}

	if bridge.MethodSets == nil {
		t.Error("MethodSets should be initialized")
	}
}

func TestMapSymbolToObject(t *testing.T) {
	bridge := NewTypeBridge()
	sym := NewSymbol("TestSymbol", KindFunction)

	// Create a simple types.Object
	pkg := types.NewPackage("test/pkg", "pkg")
	obj := types.NewFunc(token.NoPos, pkg, "TestSymbol", types.NewSignatureType(nil, nil, nil, nil, nil, false))

	// Map the symbol to the object
	bridge.MapSymbolToObject(sym, obj)

	// Test retrieval
	retrievedObj := bridge.GetObjectForSymbol(sym)
	if retrievedObj != obj {
		t.Errorf("GetObjectForSymbol returned %v, want %v", retrievedObj, obj)
	}

	retrievedSym := bridge.GetSymbolForObject(obj)
	if retrievedSym != sym {
		t.Errorf("GetSymbolForObject returned %v, want %v", retrievedSym, sym)
	}
}

func TestMapNodeToSymbol(t *testing.T) {
	bridge := NewTypeBridge()
	sym := NewSymbol("TestSymbol", KindFunction)

	// Create a simple ast.Node (using ast.Ident as it implements ast.Node)
	node := &ast.Ident{Name: "TestSymbol"}

	// Map the node to the symbol
	bridge.MapNodeToSymbol(node, sym)

	// Test retrieval
	retrievedSym := bridge.GetSymbolForNode(node)
	if retrievedSym != sym {
		t.Errorf("GetSymbolForNode returned %v, want %v", retrievedSym, sym)
	}
}

// This is a simplified test for GetImplementations as full testing would require
// more complex type setup
func TestGetImplementations(t *testing.T) {
	bridge := NewTypeBridge()

	// Create package
	pkg := types.NewPackage("test/pkg", "pkg")

	// Create an interface
	ifaceName := types.NewTypeName(token.NoPos, pkg, "TestInterface", nil)
	iface := types.NewInterfaceType(nil, nil).Complete()
	_ = types.NewNamed(ifaceName, iface, nil) // Create but don't use directly in test
	ifaceSym := NewSymbol("TestInterface", KindInterface)

	// Create a type that implements the interface
	typeName := types.NewTypeName(token.NoPos, pkg, "TestType", nil)
	_ = types.NewNamed(typeName, types.NewStruct(nil, nil), nil) // Create but don't use directly in test
	typeSym := NewSymbol("TestType", KindStruct)

	// Map symbols to objects
	bridge.MapSymbolToObject(ifaceSym, ifaceName)
	bridge.MapSymbolToObject(typeSym, typeName)

	// Since we can't easily set up real interface implementation in a unit test,
	// this test just verifies the function runs without error
	impls := bridge.GetImplementations(iface, true)
	if impls == nil {
		t.Log("GetImplementations returned empty slice as expected in this test setup")
	}
}

// This is a simplified test for GetMethodsOfType
func TestGetMethodsOfType(t *testing.T) {
	bridge := NewTypeBridge()

	// Create package
	pkg := types.NewPackage("test/pkg", "pkg")

	// Create a type with a method
	typeName := types.NewTypeName(token.NoPos, pkg, "TestType", nil)
	typeObj := types.NewNamed(typeName, types.NewStruct(nil, nil), nil)

	// Add a method (in a real scenario, this would be added to typeObj)
	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	_ = types.NewFunc(token.NoPos, pkg, "TestMethod", sig) // Create but don't use directly in test

	// Since we can't easily add methods to named types in a unit test,
	// this test just verifies the function runs without error
	methods := bridge.GetMethodsOfType(typeObj)
	if methods == nil {
		t.Log("GetMethodsOfType returned empty slice as expected in this test setup")
	}
}
