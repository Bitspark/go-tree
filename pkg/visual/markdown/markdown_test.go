package markdown

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
)

// TestMarkdownVisitor tests the Markdown visitor implementation
func TestMarkdownVisitor(t *testing.T) {
	// Create a simple package
	pkg := &model.GoPackage{
		Name:       "testpkg",
		PackageDoc: "This is a test package",
		Types: []model.GoType{
			{
				Name: "Person",
				Kind: "struct",
				Doc:  "Person represents a person",
				Code: "type Person struct {\n\tName string\n\tAge int\n}",
				Fields: []model.GoField{
					{Name: "Name", Type: "string", Comment: "The person's name"},
					{Name: "Age", Type: "int", Comment: "The person's age"},
				},
			},
			{
				Name: "Reader",
				Kind: "interface",
				Doc:  "Reader is an interface for reading data",
				Code: "type Reader interface {\n\tRead(p []byte) (n int, err error)\n}",
				InterfaceMethods: []model.GoMethod{
					{Name: "Read", Signature: "(p []byte) (n int, err error)", Comment: "Reads data into p"},
				},
			},
		},
		Functions: []model.GoFunction{
			{
				Name:      "NewPerson",
				Signature: "(name string, age int) *Person",
				Doc:       "NewPerson creates a new person",
				Code:      "func NewPerson(name string, age int) *Person {\n\treturn &Person{Name: name, Age: age}\n}",
			},
			{
				Name:      "Read",
				Signature: "(p []byte) (n int, err error)",
				Doc:       "Read implements the Reader interface",
				Code:      "func (p *Person) Read(p []byte) (n int, err error) {\n\treturn 0, nil\n}",
				Receiver:  &model.GoReceiver{Name: "p", Type: "*Person"},
			},
		},
	}

	// Create visitor with default options
	visitor := NewMarkdownVisitor(DefaultOptions())

	// Test visiting the package and all its elements
	err := visitor.VisitPackage(pkg)
	if err != nil {
		t.Fatalf("VisitPackage failed: %v", err)
	}

	for _, typ := range pkg.Types {
		err = visitor.VisitType(typ)
		if err != nil {
			t.Fatalf("VisitType failed for %s: %v", typ.Name, err)
		}
	}

	for _, fn := range pkg.Functions {
		err = visitor.VisitFunction(fn)
		if err != nil {
			t.Fatalf("VisitFunction failed for %s: %v", fn.Name, err)
		}
	}

	// Get the result
	result, err := visitor.Result()
	if err != nil {
		t.Fatalf("Result failed: %v", err)
	}

	// Check that the markdown contains expected elements
	expectedElements := []string{
		"# Package testpkg",
		"This is a test package",
		"## Type: Person (struct)",
		"Person represents a person",
		"```go",
		"type Person struct {",
		"### Fields",
		"| Name | Type | Tag | Comment |",
		"| Name | string |",
		"| Age | int |",
		"## Type: Reader (interface)",
		"Reader is an interface for reading data",
		"### Methods",
		"| Name | Signature | Comment |",
		"| Read | (p []byte) (n int, err error) |",
		"## Function: NewPerson",
		"NewPerson creates a new person",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Result doesn't contain expected element: %s", expected)
		}
	}
}

// TestMarkdownGenerator tests the Markdown generator
func TestMarkdownGenerator(t *testing.T) {
	// Create a simple package
	pkg := &model.GoPackage{
		Name:       "testpkg",
		PackageDoc: "This is a test package",
		Types: []model.GoType{
			{
				Name: "Person",
				Kind: "struct",
				Doc:  "Person represents a person",
			},
		},
	}

	// Create generator with custom options
	options := Options{
		IncludeCodeBlocks: true,
		IncludeLinks:      false,
		IncludeTOC:        true,
	}
	generator := NewGenerator(options)

	// Generate markdown
	markdown, err := generator.Generate(pkg)
	if err != nil {
		t.Fatalf("Failed to generate markdown: %v", err)
	}

	// Check basic content
	if !strings.Contains(markdown, "# Package testpkg") {
		t.Error("Generated markdown doesn't contain package name")
	}

	if !strings.Contains(markdown, "This is a test package") {
		t.Error("Generated markdown doesn't contain package documentation")
	}

	if !strings.Contains(markdown, "## Type: Person") {
		t.Error("Generated markdown doesn't contain type information")
	}

	// Test with JSON input
	jsonData := []byte(`{
		"name": "testpkg",
		"packageDoc": "This is a test package from JSON",
		"types": [
			{
				"name": "Person",
				"kind": "struct",
				"doc": "Person represents a person"
			}
		]
	}`)

	markdownFromJSON, err := generator.GenerateFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to generate markdown from JSON: %v", err)
	}

	if !strings.Contains(markdownFromJSON, "This is a test package from JSON") {
		t.Error("Generated markdown from JSON doesn't contain expected content")
	}
}
