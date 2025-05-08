package interfaceanalysis

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/module"
)

// TestExtractInterfaces tests finding and extracting potential interfaces
func TestExtractInterfaces(t *testing.T) {
	// Create a test package with common methods
	fileRead := &module.Function{
		Name:      "Read",
		Signature: "(p []byte) (n int, err error)",
		Receiver:  &module.Receiver{Type: "*File"},
	}
	fileWrite := &module.Function{
		Name:      "Write",
		Signature: "(p []byte) (n int, err error)",
		Receiver:  &module.Receiver{Type: "*File"},
	}
	fileClose := &module.Function{
		Name:      "Close",
		Signature: "() error",
		Receiver:  &module.Receiver{Type: "*File"},
	}
	socketRead := &module.Function{
		Name:      "Read",
		Signature: "(p []byte) (n int, err error)",
		Receiver:  &module.Receiver{Type: "*Socket"},
	}
	socketWrite := &module.Function{
		Name:      "Write",
		Signature: "(p []byte) (n int, err error)",
		Receiver:  &module.Receiver{Type: "*Socket"},
	}
	socketClose := &module.Function{
		Name:      "Close",
		Signature: "() error",
		Receiver:  &module.Receiver{Type: "*Socket"},
	}
	bufferRead := &module.Function{
		Name:      "Read",
		Signature: "(p []byte) (n int, err error)",
		Receiver:  &module.Receiver{Type: "*Buffer"},
	}
	bufferWrite := &module.Function{
		Name:      "Write",
		Signature: "(p []byte) (n int, err error)",
		Receiver:  &module.Receiver{Type: "*Buffer"},
	}
	bufferReset := &module.Function{
		Name:      "Reset",
		Signature: "()",
		Receiver:  &module.Receiver{Type: "*Buffer"},
	}

	pkg := &module.Package{
		Name: "testpackage",
		Functions: map[string]*module.Function{
			"File.Read":    fileRead,
			"File.Write":   fileWrite,
			"File.Close":   fileClose,
			"Socket.Read":  socketRead,
			"Socket.Write": socketWrite,
			"Socket.Close": socketClose,
			"Buffer.Read":  bufferRead,
			"Buffer.Write": bufferWrite,
			"Buffer.Reset": bufferReset,
		},
	}

	analyzer := NewAnalyzer()
	analysis := analyzer.AnalyzeReceivers(pkg)
	interfaces := analyzer.ExtractInterfaces(analysis)

	// Check that we found at least one interface
	if len(interfaces) == 0 {
		t.Fatal("Expected to extract at least one interface")
	}

	// Look for an interface with Read and Write methods
	var rwInterface *InterfaceDefinition
	for i, intf := range interfaces {
		if _, hasRead := intf.Methods["Read"]; hasRead {
			if _, hasWrite := intf.Methods["Write"]; hasWrite {
				rwInterface = &interfaces[i]
				break
			}
		}
	}

	if rwInterface == nil {
		t.Fatal("Expected to find an interface with Read and Write methods")
	}

	// Check that the interface has expected methods
	if len(rwInterface.Methods) < 2 {
		t.Errorf("Expected at least 2 methods, got %d", len(rwInterface.Methods))
	}

	if _, hasRead := rwInterface.Methods["Read"]; !hasRead {
		t.Error("Expected Read method in extracted interface")
	}

	if _, hasWrite := rwInterface.Methods["Write"]; !hasWrite {
		t.Error("Expected Write method in extracted interface")
	}

	// Check that all three receiver types are in the source types
	if len(rwInterface.SourceTypes) < 3 {
		t.Errorf("Expected at least 3 source types, got %d", len(rwInterface.SourceTypes))
	}

	hasFile := false
	hasSocket := false
	hasBuffer := false

	for _, sourceType := range rwInterface.SourceTypes {
		if sourceType == "*File" {
			hasFile = true
		}
		if sourceType == "*Socket" {
			hasSocket = true
		}
		if sourceType == "*Buffer" {
			hasBuffer = true
		}
	}

	if !hasFile {
		t.Error("Expected *File as a source type")
	}
	if !hasSocket {
		t.Error("Expected *Socket as a source type")
	}
	if !hasBuffer {
		t.Error("Expected *Buffer as a source type")
	}
}

// TestGenerateInterfaceCode tests generating Go code for an interface
func TestGenerateInterfaceCode(t *testing.T) {
	interfaceDef := InterfaceDefinition{
		Name: "Reader",
		Methods: map[string]string{
			"Read": "(p []byte) (n int, err error)",
		},
		SourceTypes: []string{"*File", "*Socket", "*Buffer"},
	}

	analyzer := NewAnalyzer()
	code := analyzer.GenerateInterfaceCode(interfaceDef)

	// Check basic structure
	if !strings.Contains(code, "type Reader interface {") {
		t.Error("Expected 'type Reader interface {' in generated code")
	}

	// Check method signature
	if !strings.Contains(code, "Read(p []byte) (n int, err error)") {
		t.Error("Expected Read method signature in generated code")
	}

	// Check comment
	if !strings.Contains(code, "// Reader represents common behavior implemented by: *File, *Socket, *Buffer") {
		t.Error("Expected documentation comment with source types")
	}

	// Test a more complex interface
	complexInterface := InterfaceDefinition{
		Name: "ReadWriter",
		Methods: map[string]string{
			"Read":  "(p []byte) (n int, err error)",
			"Write": "(p []byte) (n int, err error)",
			"Close": "() error",
		},
		SourceTypes: []string{"*File", "*Socket"},
	}

	complexCode := analyzer.GenerateInterfaceCode(complexInterface)

	// Verify all methods are included
	if !strings.Contains(complexCode, "Read(p []byte) (n int, err error)") {
		t.Error("Expected Read method signature in complex interface")
	}

	if !strings.Contains(complexCode, "Write(p []byte) (n int, err error)") {
		t.Error("Expected Write method signature in complex interface")
	}

	if !strings.Contains(complexCode, "Close() error") {
		t.Error("Expected Close method signature in complex interface")
	}
}
