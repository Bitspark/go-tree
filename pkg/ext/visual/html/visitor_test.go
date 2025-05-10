package html

import (
	"bitspark.dev/go-tree/pkg/ext/visual/formatter"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

func TestNewHTMLVisitor(t *testing.T) {
	// Test with nil options
	v1 := NewHTMLVisitor(nil)
	if v1 == nil {
		t.Fatal("NewHTMLVisitor returned nil with nil options")
	}

	if v1.options == nil {
		t.Fatal("HTMLVisitor has nil options when initialized with nil")
	}

	if v1.options.DetailLevel != 3 {
		t.Errorf("Expected default detail level 3, got %d", v1.options.DetailLevel)
	}

	// Test with custom options
	opts := &formatter.FormatOptions{
		DetailLevel:            5,
		IncludeTypeAnnotations: true,
	}

	v2 := NewHTMLVisitor(opts)
	if v2.options.DetailLevel != 5 {
		t.Errorf("Expected detail level 5, got %d", v2.options.DetailLevel)
	}

	if !v2.options.IncludeTypeAnnotations {
		t.Error("Expected IncludeTypeAnnotations to be true")
	}

	// Check initial state
	if v2.buffer == nil {
		t.Error("Buffer should be initialized")
	}

	if v2.visitedSymbols == nil {
		t.Error("VisitedSymbols map should be initialized")
	}
}

func TestHTMLVisitorResult(t *testing.T) {
	v := NewHTMLVisitor(nil)
	v.Write("<test>content</test>")

	result, err := v.Result()
	if err != nil {
		t.Fatalf("Result returned error: %v", err)
	}

	if result != "<test>content</test>" {
		t.Errorf("Expected '<test>content</test>', got '%s'", result)
	}
}

func TestVisitModule(t *testing.T) {
	v := NewHTMLVisitor(nil)
	mod := &typesys.Module{
		Path:      "example.com/test",
		GoVersion: "1.18",
	}

	err := v.VisitModule(mod)
	if err != nil {
		t.Fatalf("VisitModule returned error: %v", err)
	}

	result, _ := v.Result()
	if !strings.Contains(result, "<div class=\"packages\">") {
		t.Error("Expected output to contain packages div")
	}
}

func TestVisitPackage(t *testing.T) {
	v := NewHTMLVisitor(nil)
	mod := &typesys.Module{
		Path:      "example.com/test",
		GoVersion: "1.18",
	}

	pkg := &typesys.Package{
		Module:     mod,
		Name:       "test",
		ImportPath: "example.com/test",
	}

	err := v.VisitPackage(pkg)
	if err != nil {
		t.Fatalf("VisitPackage returned error: %v", err)
	}

	result, _ := v.Result()

	expectedFragments := []string{
		"<div class=\"package\" id=\"pkg-test\">",
		"<h2>Package test</h2>",
		"<div class=\"package-import\">example.com/test</div>",
		"<h3>Types</h3>",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(result, fragment) {
			t.Errorf("Expected output to contain '%s'", fragment)
		}
	}

	// Test if current package is set
	if v.currentPackage != pkg {
		t.Error("Expected currentPackage to be set to the visited package")
	}
}

func TestAfterVisitPackage(t *testing.T) {
	v := NewHTMLVisitor(nil)
	mod := &typesys.Module{
		Path:      "example.com/test",
		GoVersion: "1.18",
	}

	pkg := &typesys.Package{
		Module:     mod,
		Name:       "test",
		ImportPath: "example.com/test",
	}

	v.currentPackage = pkg

	// First need to visit the package to set up the proper indentation level
	err := v.VisitPackage(pkg)
	if err != nil {
		t.Fatalf("VisitPackage returned error: %v", err)
	}

	err = v.AfterVisitPackage(pkg)
	if err != nil {
		t.Fatalf("AfterVisitPackage returned error: %v", err)
	}

	result, _ := v.Result()

	expectedFragments := []string{
		"<h3>Functions</h3>",
		"<h3>Variables and Constants</h3>",
		"</div>", // Closing package div
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(result, fragment) {
			t.Errorf("Expected output to contain '%s'", fragment)
		}
	}

	// Test if current package is cleared
	if v.currentPackage != nil {
		t.Error("Expected currentPackage to be nil after AfterVisitPackage")
	}
}

func TestGetSymbolClass(t *testing.T) {
	v := NewHTMLVisitor(nil)

	// Test different kinds of symbols
	testCases := []struct {
		kind     typesys.SymbolKind
		exported bool
		expected string
	}{
		{typesys.KindFunction, true, "symbol symbol-fn exported"},
		{typesys.KindFunction, false, "symbol symbol-fn private"},
		{typesys.KindType, true, "symbol symbol-type exported"},
		{typesys.KindVariable, false, "symbol symbol-var private"},
		{typesys.KindConstant, true, "symbol symbol-const exported"},
	}

	for _, tc := range testCases {
		sym := &typesys.Symbol{
			Kind:     tc.kind,
			Exported: tc.exported,
		}

		class := v.getSymbolClass(sym)
		if class != tc.expected {
			t.Errorf("Expected class '%s' for kind %v, exported=%v, got '%s'",
				tc.expected, tc.kind, tc.exported, class)
		}
	}

	// Test nil symbol
	if class := v.getSymbolClass(nil); class != "" {
		t.Errorf("Expected empty class for nil symbol, got '%s'", class)
	}
}

func TestVisitType(t *testing.T) {
	opts := &formatter.FormatOptions{
		IncludePrivate: true,
	}
	v := NewHTMLVisitor(opts)

	mod := &typesys.Module{
		Path:      "example.com/test",
		GoVersion: "1.18",
	}

	pkg := &typesys.Package{
		Module:     mod,
		Name:       "test",
		ImportPath: "example.com/test",
	}

	file := &typesys.File{
		Package: pkg,
		Path:    "test.go",
		Symbols: []*typesys.Symbol{},
	}

	sym := &typesys.Symbol{
		ID:       "test.MyType",
		Package:  pkg,
		File:     file,
		Name:     "MyType",
		Kind:     typesys.KindType,
		Exported: true,
	}

	// Test the visit tracking functionality
	err := v.VisitType(sym)
	if err != nil {
		t.Fatalf("VisitType returned error: %v", err)
	}

	// Test symbol tracking
	if !v.visitedSymbols[sym.ID] {
		t.Error("Symbol should be marked as visited")
	}

	// Test repeated visits are ignored
	beforeLen := len(v.buffer.String())
	_ = v.VisitType(sym)
	afterResult := v.buffer.String()
	if len(afterResult) != beforeLen {
		t.Error("Visiting the same symbol twice should not add more content")
	}

	// Test that private symbols are filtered when IncludePrivate is false
	v = NewHTMLVisitor(&formatter.FormatOptions{
		IncludePrivate: false,
	})

	privateSym := &typesys.Symbol{
		ID:       "test.privateType",
		Package:  pkg,
		File:     file,
		Name:     "privateType",
		Kind:     typesys.KindType,
		Exported: false,
	}

	// Should not add the private symbol to the result
	_ = v.VisitType(privateSym)

	// Private symbol should not be tracked
	if v.visitedSymbols[privateSym.ID] {
		t.Error("Private symbol should not be marked as visited when IncludePrivate is false")
	}
}

func TestVisitSymbolFiltering(t *testing.T) {
	// Test that private symbols are filtered when IncludePrivate is false
	opts := &formatter.FormatOptions{
		IncludePrivate: false,
	}
	v := NewHTMLVisitor(opts)

	pkg := &typesys.Package{
		Name:       "test",
		ImportPath: "example.com/test",
	}

	privateSym := &typesys.Symbol{
		ID:       "test.privateFunc",
		Package:  pkg,
		Name:     "privateFunc",
		Kind:     typesys.KindFunction,
		Exported: false,
	}

	err := v.VisitFunction(privateSym)
	if err != nil {
		t.Fatalf("VisitFunction returned error: %v", err)
	}

	result, _ := v.Result()
	if strings.Contains(result, "privateFunc") {
		t.Error("Private function should be filtered out when IncludePrivate is false")
	}
}

func TestIndent(t *testing.T) {
	v := NewHTMLVisitor(nil)

	// Test initial indent level
	if v.Indent() != "" {
		t.Errorf("Expected empty indent at level 0, got '%s'", v.Indent())
	}

	// Test increasing indent
	v.indentLevel = 2
	if v.Indent() != "        " { // 8 spaces (4 * 2)
		t.Errorf("Expected 8 spaces at level 2, got '%s'", v.Indent())
	}
}
