// Package html provides functionality for generating HTML documentation
// from Go modules.
package html

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/core/visitor"
)

// HTMLVisitor implements visitor.ModuleVisitor to generate HTML documentation
// for Go modules. It can be used with visitor.ModuleWalker to traverse a module
// structure and generate HTML content for each element.
type HTMLVisitor struct {
	// Buffer to store HTML content
	buffer bytes.Buffer

	// Current indentation level
	indentLevel int

	// Options for HTML generation
	IncludePrivate   bool
	IncludeTests     bool
	IncludeGenerated bool
	Title            string
}

// NewHTMLVisitor creates a new HTML visitor
func NewHTMLVisitor() *HTMLVisitor {
	return &HTMLVisitor{
		Title: "Go Module Documentation",
	}
}

// Helper methods for HTML generation
func (v *HTMLVisitor) writeString(s string) {
	for i := 0; i < v.indentLevel; i++ {
		v.buffer.WriteString("  ")
	}
	v.buffer.WriteString(s)
}

func (v *HTMLVisitor) indent() {
	v.indentLevel++
}

func (v *HTMLVisitor) dedent() {
	if v.indentLevel > 0 {
		v.indentLevel--
	}
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	return template.HTMLEscapeString(s)
}

// formatDocComment formats a documentation comment for HTML display
func formatDocComment(doc string) string {
	if doc == "" {
		return ""
	}

	// Escape HTML characters to prevent XSS
	doc = template.HTMLEscapeString(doc)

	// Replace newlines with <br> for HTML display
	doc = strings.ReplaceAll(doc, "\n", "<br>")

	return doc
}

// typeKindClass returns a CSS class based on the type kind
func typeKindClass(kind string) string {
	switch kind {
	case "struct":
		return "type-struct"
	case "interface":
		return "type-interface"
	case "alias":
		return "type-alias"
	default:
		return "type-other"
	}
}

// formatCode formats Go code with syntax highlighting
func formatCode(code string) string {
	// Simple formatting for now
	if code == "" {
		return ""
	}

	// Escape HTML characters
	code = template.HTMLEscapeString(code)

	// Add basic syntax highlighting classes
	code = strings.ReplaceAll(code, "func ", "<span class=\"keyword\">func</span> ")
	code = strings.ReplaceAll(code, "type ", "<span class=\"keyword\">type</span> ")
	code = strings.ReplaceAll(code, "struct ", "<span class=\"keyword\">struct</span> ")
	code = strings.ReplaceAll(code, "interface ", "<span class=\"keyword\">interface</span> ")
	code = strings.ReplaceAll(code, "package ", "<span class=\"keyword\">package</span> ")
	code = strings.ReplaceAll(code, "import ", "<span class=\"keyword\">import</span> ")
	code = strings.ReplaceAll(code, "const ", "<span class=\"keyword\">const</span> ")
	code = strings.ReplaceAll(code, "var ", "<span class=\"keyword\">var</span> ")
	code = strings.ReplaceAll(code, "return ", "<span class=\"keyword\">return</span> ")

	// Add string literals highlighting
	parts := strings.Split(code, "\"")
	for i := 1; i < len(parts); i += 2 {
		if i < len(parts) {
			parts[i] = "<span class=\"string\">\"" + parts[i] + "\"</span>"
		}
	}
	code = strings.Join(parts, "")

	return code
}

// isExported checks if a name is exported (starts with uppercase)
func isExported(name string) bool {
	if name == "" {
		return false
	}
	firstChar := name[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}

// sanitizeAnchor creates a valid HTML anchor from a name
func sanitizeAnchor(name string) string {
	// Replace spaces and special characters with dashes
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}

// CSS for HTML documentation
const htmlCSS = `
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
  line-height: 1.5;
  color: #333;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  margin-bottom: 30px;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}

.header h1 {
  margin-bottom: 5px;
}

.version {
  color: #666;
  font-size: 0.9em;
}

.module-info {
  margin-bottom: 20px;
}

.description {
  margin-top: 10px;
  margin-bottom: 20px;
  font-style: italic;
}

.package {
  margin-bottom: 40px;
  border: 1px solid #eee;
  border-radius: 5px;
  padding: 20px;
}

.package-name {
  margin-top: 0;
  color: #333;
}

.type, .function, .method, .variable, .constant {
  margin: 20px 0;
  padding: 15px;
  border-left: 4px solid #ddd;
  background-color: #f9f9f9;
}

.type-struct {
  border-left-color: #4caf50;
}

.type-interface {
  border-left-color: #2196f3;
}

.type-alias {
  border-left-color: #ff9800;
}

.doc-comment {
  margin: 10px 0;
  padding: 10px;
  background-color: #f5f5f5;
  border-radius: 3px;
}

.code {
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 3px;
  overflow-x: auto;
  font-size: 0.9em;
  line-height: 1.4;
}

.keyword {
  color: #07a;
}

.string {
  color: #690;
}

.fields-table {
  width: 100%;
  border-collapse: collapse;
  margin: 15px 0;
}

.fields-table th, .fields-table td {
  padding: 8px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.fields-table th {
  background-color: #f5f5f5;
}

h2, h3, h4, h5 {
  color: #333;
}

a {
  color: #0366d6;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}

.line-number {
  color: #999;
  margin-right: 10px;
  user-select: none;
}
`

// VisitModule implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitModule(mod *module.Module) error {
	// Start HTML document
	v.writeString("<!DOCTYPE html>\n")
	v.writeString("<html lang=\"en\">\n")
	v.writeString("<head>\n")
	v.indent()
	v.writeString("<meta charset=\"UTF-8\">\n")
	v.writeString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")

	// Use title or module path
	title := v.Title
	if title == "" {
		title = fmt.Sprintf("Documentation for %s", mod.Path)
	}
	v.writeString(fmt.Sprintf("<title>%s</title>\n", escapeHTML(title)))

	// Add CSS
	v.writeString("<style>\n")
	v.writeString(htmlCSS)
	v.writeString("</style>\n")
	v.dedent()
	v.writeString("</head>\n")
	v.writeString("<body>\n")
	v.indent()

	// Document header
	v.writeString("<div class=\"header\">\n")
	v.indent()
	v.writeString(fmt.Sprintf("<h1>%s</h1>\n", escapeHTML(title)))
	if mod.Version != "" {
		v.writeString(fmt.Sprintf("<div class=\"version\">%s</div>\n", escapeHTML(mod.Version)))
	}
	v.dedent()
	v.writeString("</div>\n")

	// Module info section
	v.writeString("<div class=\"module-info\">\n")
	v.indent()
	v.writeString(fmt.Sprintf("<p><strong>Path:</strong> %s</p>\n", escapeHTML(mod.Path)))
	if mod.GoVersion != "" {
		v.writeString(fmt.Sprintf("<p><strong>Go Version:</strong> %s</p>\n", escapeHTML(mod.GoVersion)))
	}
	v.dedent()
	v.writeString("</div>\n")

	// Create package index
	v.writeString("<div class=\"package-index\">\n")
	v.indent()
	v.writeString("<h2>Packages</h2>\n")
	v.writeString("<ul>\n")
	v.indent()
	for _, pkg := range mod.Packages {
		// Skip test packages if not included
		if pkg.IsTest && !v.IncludeTests {
			continue
		}

		pkgID := sanitizeAnchor(pkg.ImportPath)
		v.writeString(fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>\n",
			pkgID, escapeHTML(pkg.ImportPath)))
	}
	v.dedent()
	v.writeString("</ul>\n")
	v.dedent()
	v.writeString("</div>\n")

	return nil
}

// VisitPackage implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitPackage(pkg *module.Package) error {
	// Skip test packages if not included
	if pkg.IsTest && !v.IncludeTests {
		return nil
	}

	pkgID := sanitizeAnchor(pkg.ImportPath)
	v.writeString(fmt.Sprintf("<div id=\"%s\" class=\"package\">\n", pkgID))
	v.indent()
	v.writeString(fmt.Sprintf("<h2 class=\"package-name\">Package %s</h2>\n", escapeHTML(pkg.Name)))

	// Package documentation
	if pkg.Documentation != "" {
		v.writeString(fmt.Sprintf("<div class=\"doc-comment\">%s</div>\n", formatDocComment(pkg.Documentation)))
	}

	// Create section for types if any
	if len(pkg.Types) > 0 {
		v.writeString("<div class=\"types-section\">\n")
		v.indent()
		v.writeString("<h3>Types</h3>\n")
		v.writeString("<ul class=\"type-list\">\n")
		v.indent()

		for _, typ := range pkg.Types {
			if !v.IncludePrivate && !typ.IsExported {
				continue
			}

			typeID := sanitizeAnchor(pkg.ImportPath + "." + typ.Name)
			v.writeString(fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>\n",
				typeID, escapeHTML(typ.Name)))
		}

		v.dedent()
		v.writeString("</ul>\n")
		v.dedent()
		v.writeString("</div>\n")
	}

	// Create section for functions if any
	if len(pkg.Functions) > 0 {
		v.writeString("<div class=\"functions-section\">\n")
		v.indent()
		v.writeString("<h3>Functions</h3>\n")
		v.writeString("<ul class=\"function-list\">\n")
		v.indent()

		for _, fn := range pkg.Functions {
			if !v.IncludePrivate && !fn.IsExported {
				continue
			}

			fnID := sanitizeAnchor(pkg.ImportPath + "." + fn.Name)
			v.writeString(fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>\n",
				fnID, escapeHTML(fn.Name)))
		}

		v.dedent()
		v.writeString("</ul>\n")
		v.dedent()
		v.writeString("</div>\n")
	}

	v.dedent()
	v.writeString("</div>\n")

	return nil
}

// VisitFile implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitFile(file *module.File) error {
	// Skip test files if not included
	if file.IsTest && !v.IncludeTests {
		return nil
	}

	// Skip generated files if not included
	if file.IsGenerated && !v.IncludeGenerated {
		return nil
	}

	// We don't create separate sections for files in the HTML output
	// as we organize by package and types/functions instead
	return nil
}

// VisitType implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitType(typ *module.Type) error {
	if !v.IncludePrivate && !typ.IsExported {
		return nil
	}

	typeID := sanitizeAnchor(typ.Package.ImportPath + "." + typ.Name)
	v.writeString(fmt.Sprintf("<div id=\"%s\" class=\"type %s\">\n",
		typeID, typeKindClass(typ.Kind)))
	v.indent()

	v.writeString(fmt.Sprintf("<h4 class=\"type-name\">type %s</h4>\n", escapeHTML(typ.Name)))

	// Type documentation
	if typ.Doc != "" {
		v.writeString(fmt.Sprintf("<div class=\"doc-comment\">%s</div>\n", formatDocComment(typ.Doc)))
	}

	// Type definition
	v.writeString("<div class=\"type-def\">\n")
	v.indent()
	// Construct a simple definition from the type kind
	typeDef := fmt.Sprintf("type %s %s", typ.Name, typ.Kind)
	v.writeString(fmt.Sprintf("<pre class=\"code\">%s</pre>\n", formatCode(typeDef)))
	v.dedent()
	v.writeString("</div>\n")

	v.dedent()
	v.writeString("</div>\n")

	return nil
}

// VisitFunction implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitFunction(fn *module.Function) error {
	if !v.IncludePrivate && !fn.IsExported {
		return nil
	}

	fnID := sanitizeAnchor(fn.Package.ImportPath + "." + fn.Name)
	v.writeString(fmt.Sprintf("<div id=\"%s\" class=\"function\">\n", fnID))
	v.indent()

	v.writeString(fmt.Sprintf("<h4 class=\"function-name\">func %s</h4>\n", escapeHTML(fn.Name)))

	// Function documentation
	if fn.Doc != "" {
		v.writeString(fmt.Sprintf("<div class=\"doc-comment\">%s</div>\n", formatDocComment(fn.Doc)))
	}

	// Function signature
	v.writeString("<div class=\"function-signature\">\n")
	v.indent()
	v.writeString(fmt.Sprintf("<pre class=\"code\">%s</pre>\n", formatCode(fn.Signature)))
	v.dedent()
	v.writeString("</div>\n")

	v.dedent()
	v.writeString("</div>\n")

	return nil
}

// VisitMethod implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitMethod(method *module.Method) error {
	if !v.IncludePrivate && !isExported(method.Name) {
		return nil
	}

	methodID := sanitizeAnchor(method.Parent.Package.ImportPath + "." + method.Parent.Name + "." + method.Name)
	v.writeString(fmt.Sprintf("<div id=\"%s\" class=\"method\">\n", methodID))
	v.indent()

	// Construct receiver name from parent type
	receiverName := method.Parent.Name
	if receiverName == "" {
		receiverName = "receiver"
	}
	v.writeString(fmt.Sprintf("<h4 class=\"method-name\">func (%s) %s</h4>\n",
		escapeHTML(receiverName), escapeHTML(method.Name)))

	// Method documentation
	if method.Doc != "" {
		v.writeString(fmt.Sprintf("<div class=\"doc-comment\">%s</div>\n", formatDocComment(method.Doc)))
	}

	// Method signature
	v.writeString("<div class=\"method-signature\">\n")
	v.indent()
	v.writeString(fmt.Sprintf("<pre class=\"code\">%s</pre>\n", formatCode(method.Signature)))
	v.dedent()
	v.writeString("</div>\n")

	v.dedent()
	v.writeString("</div>\n")

	return nil
}

// VisitField implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitField(field *module.Field) error {
	// Field rendering is handled in VisitType
	return nil
}

// VisitVariable implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitVariable(variable *module.Variable) error {
	if !v.IncludePrivate && !variable.IsExported {
		return nil
	}

	varID := sanitizeAnchor(variable.Package.ImportPath + "." + variable.Name)
	v.writeString(fmt.Sprintf("<div id=\"%s\" class=\"variable\">\n", varID))
	v.indent()

	v.writeString(fmt.Sprintf("<h4 class=\"variable-name\">var %s</h4>\n", escapeHTML(variable.Name)))

	// Variable documentation
	if variable.Doc != "" {
		v.writeString(fmt.Sprintf("<div class=\"doc-comment\">%s</div>\n", formatDocComment(variable.Doc)))
	}

	// Variable definition
	v.writeString("<div class=\"variable-def\">\n")
	v.indent()
	// Construct a definition from Type and Value
	varDef := fmt.Sprintf("var %s %s", variable.Name, variable.Type)
	if variable.Value != "" {
		varDef += " = " + variable.Value
	}
	v.writeString(fmt.Sprintf("<pre class=\"code\">%s</pre>\n", formatCode(varDef)))
	v.dedent()
	v.writeString("</div>\n")

	v.dedent()
	v.writeString("</div>\n")

	return nil
}

// VisitConstant implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitConstant(constant *module.Constant) error {
	if !v.IncludePrivate && !constant.IsExported {
		return nil
	}

	constID := sanitizeAnchor(constant.Package.ImportPath + "." + constant.Name)
	v.writeString(fmt.Sprintf("<div id=\"%s\" class=\"constant\">\n", constID))
	v.indent()

	v.writeString(fmt.Sprintf("<h4 class=\"constant-name\">const %s</h4>\n", escapeHTML(constant.Name)))

	// Constant documentation
	if constant.Doc != "" {
		v.writeString(fmt.Sprintf("<div class=\"doc-comment\">%s</div>\n", formatDocComment(constant.Doc)))
	}

	// Constant definition
	v.writeString("<div class=\"constant-def\">\n")
	v.indent()
	// Construct a definition from Type and Value
	constDef := fmt.Sprintf("const %s", constant.Name)
	if constant.Type != "" {
		constDef += " " + constant.Type
	}
	if constant.Value != "" {
		constDef += " = " + constant.Value
	}
	v.writeString(fmt.Sprintf("<pre class=\"code\">%s</pre>\n", formatCode(constDef)))
	v.dedent()
	v.writeString("</div>\n")

	v.dedent()
	v.writeString("</div>\n")

	return nil
}

// VisitImport implements visitor.ModuleVisitor
func (v *HTMLVisitor) VisitImport(imp *module.Import) error {
	// We don't create separate sections for imports in the HTML output
	return nil
}

// CreateHTML generates HTML documentation for a module by walking its structure
func CreateHTML(mod *module.Module) (string, error) {
	htmlVisitor := NewHTMLVisitor()
	walker := visitor.NewModuleWalker(htmlVisitor)

	// Configure walker as needed
	walker.IncludePrivate = false
	walker.IncludeTests = false

	// Configure visitor with same settings
	htmlVisitor.IncludePrivate = walker.IncludePrivate
	htmlVisitor.IncludeTests = walker.IncludeTests
	htmlVisitor.IncludeGenerated = walker.IncludeGenerated

	// Walk the module structure
	if err := walker.Walk(mod); err != nil {
		return "", err
	}

	// Close HTML document
	htmlVisitor.dedent()
	htmlVisitor.writeString("</body>\n")
	htmlVisitor.writeString("</html>\n")

	// Return the generated HTML content
	return htmlVisitor.buffer.String(), nil
}
