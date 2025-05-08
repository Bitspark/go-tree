package markdown

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkgold/core/module"
)

// TestMarkdownVisitor tests the Markdown visitor implementation
func TestMarkdownVisitor(t *testing.T) {
	// Create a simple module
	mod := module.NewModule("test-module", "")

	// Create a package
	pkg := &module.Package{
		Name:          "testpkg",
		ImportPath:    "test/testpkg",
		Documentation: "This is a test package",
		Types:         make(map[string]*module.Type),
		Functions:     make(map[string]*module.Function),
	}
	mod.AddPackage(pkg)

	// Create a struct type
	personType := module.NewType("Person", "struct", true)
	personType.Doc = "Person represents a person"
	pkg.Types["Person"] = personType

	// Add fields to the struct
	personType.AddField("Name", "string", "", false, "The person's name")
	personType.AddField("Age", "int", "", false, "The person's age")

	// Create an interface type
	readerType := module.NewType("Reader", "interface", true)
	readerType.Doc = "Reader is an interface for reading data"
	pkg.Types["Reader"] = readerType

	// Add a method to the interface
	readerType.AddInterfaceMethod("Read", "(p []byte) (n int, err error)", false, "Reads data into p")

	// Create a function
	newPersonFn := module.NewFunction("NewPerson", true, false)
	newPersonFn.Doc = "NewPerson creates a new person"
	newPersonFn.Signature = "(name string, age int) *Person"
	newPersonFn.AddParameter("name", "string", false)
	newPersonFn.AddParameter("age", "int", false)
	newPersonFn.AddResult("", "*Person")
	pkg.Functions["NewPerson"] = newPersonFn

	// Create a method
	readMethod := personType.AddMethod("Read", "(p []byte) (n int, err error)", false, "Read implements the Reader interface")

	// Create visitor with default options
	visitor := NewMarkdownVisitor(DefaultOptions())

	// Visit the module and package
	err := visitor.VisitModule(mod)
	if err != nil {
		t.Fatalf("VisitModule failed: %v", err)
	}

	err = visitor.VisitPackage(pkg)
	if err != nil {
		t.Fatalf("VisitPackage failed: %v", err)
	}

	// Visit types
	for _, typ := range pkg.Types {
		err = visitor.VisitType(typ)
		if err != nil {
			t.Fatalf("VisitType failed for %s: %v", typ.Name, err)
		}
	}

	// Visit functions
	for _, fn := range pkg.Functions {
		err = visitor.VisitFunction(fn)
		if err != nil {
			t.Fatalf("VisitFunction failed for %s: %v", fn.Name, err)
		}
	}

	// Visit method
	err = visitor.VisitMethod(readMethod)
	if err != nil {
		t.Fatalf("VisitMethod failed: %v", err)
	}

	// Get the result
	result, err := visitor.Result()
	if err != nil {
		t.Fatalf("Result failed: %v", err)
	}

	// Check that the markdown contains expected elements
	expectedElements := []string{
		"# Module test-module",
		"## Package testpkg",
		"This is a test package",
		"### Type: Person (struct)",
		"Person represents a person",
		"### Type: Reader (interface)",
		"Reader is an interface for reading data",
		"### Function: NewPerson",
		"NewPerson creates a new person",
		"**Signature:** `(name string, age int) *Person`",
		"### Method: (Person) Read",
		"Read implements the Reader interface",
		"**Signature:** `(p []byte) (n int, err error)`",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("Result doesn't contain expected element: %s", expected)
		}
	}
}

// TestMarkdownGenerator tests the Markdown generator
func TestMarkdownGenerator(t *testing.T) {
	// Create a simple module
	mod := module.NewModule("test-module", "")

	// Create a package
	pkg := &module.Package{
		Name:          "testpkg",
		ImportPath:    "test/testpkg",
		Documentation: "This is a test package",
		Types:         make(map[string]*module.Type),
	}
	mod.AddPackage(pkg)

	// Add a type to the package
	personType := module.NewType("Person", "struct", true)
	personType.Doc = "Person represents a person"
	pkg.Types["Person"] = personType

	// Create generator with custom options
	options := Options{
		IncludeCodeBlocks: true,
		IncludeLinks:      false,
		IncludeTOC:        true,
	}
	generator := NewGenerator(options)

	// Generate markdown
	markdown, err := generator.Generate(mod)
	if err != nil {
		t.Fatalf("Failed to generate markdown: %v", err)
	}

	// Check basic content
	if !strings.Contains(markdown, "# Module test-module") {
		t.Error("Generated markdown doesn't contain module name")
	}

	if !strings.Contains(markdown, "## Package testpkg") {
		t.Error("Generated markdown doesn't contain package name")
	}

	if !strings.Contains(markdown, "This is a test package") {
		t.Error("Generated markdown doesn't contain package documentation")
	}

	if !strings.Contains(markdown, "### Type: Person") {
		t.Error("Generated markdown doesn't contain type information")
	}

	// Test with JSON input
	jsonData := []byte(`{
		"Path": "test-module-json",
		"Packages": {
			"testpkg": {
				"Name": "testpkg",
				"ImportPath": "test/testpkg",
				"Documentation": "This is a test package from JSON",
				"Types": {
					"Person": {
						"Name": "Person",
						"Kind": "struct",
						"Doc": "Person represents a person"
					}
				}
			}
		}
	}`)

	markdownFromJSON, err := generator.GenerateFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to generate markdown from JSON: %v", err)
	}

	if !strings.Contains(markdownFromJSON, "This is a test package from JSON") {
		t.Error("Generated markdown from JSON doesn't contain expected content")
	}
}
