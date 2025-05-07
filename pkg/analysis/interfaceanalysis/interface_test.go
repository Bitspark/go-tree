package interfaceanalysis

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
)

// TestExtractInterfaces tests finding and extracting potential interfaces
func TestExtractInterfaces(t *testing.T) {
	// Create a test package with common methods
	pkg := &model.GoPackage{
		Name: "testpackage",
		Functions: []model.GoFunction{
			{
				Name:      "Read",
				Signature: "(p []byte) (n int, err error)",
				Receiver:  &model.GoReceiver{Type: "*File"},
			},
			{
				Name:      "Write",
				Signature: "(p []byte) (n int, err error)",
				Receiver:  &model.GoReceiver{Type: "*File"},
			},
			{
				Name:      "Close",
				Signature: "() error",
				Receiver:  &model.GoReceiver{Type: "*File"},
			},
			{
				Name:      "Read",
				Signature: "(p []byte) (n int, err error)",
				Receiver:  &model.GoReceiver{Type: "*Socket"},
			},
			{
				Name:      "Write",
				Signature: "(p []byte) (n int, err error)",
				Receiver:  &model.GoReceiver{Type: "*Socket"},
			},
			{
				Name:      "Close",
				Signature: "() error",
				Receiver:  &model.GoReceiver{Type: "*Socket"},
			},
			{
				Name:      "Read",
				Signature: "(p []byte) (n int, err error)",
				Receiver:  &model.GoReceiver{Type: "*Buffer"},
			},
			{
				Name:      "Write",
				Signature: "(p []byte) (n int, err error)",
				Receiver:  &model.GoReceiver{Type: "*Buffer"},
			},
			{
				Name:      "Reset",
				Signature: "()",
				Receiver:  &model.GoReceiver{Type: "*Buffer"},
			},
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
