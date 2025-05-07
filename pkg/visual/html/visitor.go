package html

import (
	"bytes"
	"html/template"

	"bitspark.dev/go-tree/pkg/core/model"
)

// HTMLVisitor implements formatter.Visitor for HTML output
type HTMLVisitor struct {
	options     Options
	packageData struct {
		Package            *model.GoPackage
		Title              string
		IncludeCSS         bool
		CustomCSS          template.CSS
		EnableHighlighting bool
		// We could collect more data here during traversal if needed
	}
}

// NewHTMLVisitor creates a new HTML visitor with the given options
func NewHTMLVisitor(options Options) *HTMLVisitor {
	v := &HTMLVisitor{options: options}
	v.packageData.Title = options.Title
	v.packageData.IncludeCSS = options.IncludeCSS
	v.packageData.CustomCSS = template.CSS(options.CustomCSS)
	v.packageData.EnableHighlighting = options.SyntaxHighlighting
	return v
}

// VisitPackage stores the package and initializes the visitor
func (v *HTMLVisitor) VisitPackage(pkg *model.GoPackage) error {
	v.packageData.Package = pkg
	return nil
}

// VisitType processes a type declaration
func (v *HTMLVisitor) VisitType(typ model.GoType) error {
	// We could do more processing here if needed
	return nil
}

// VisitFunction processes a function declaration
func (v *HTMLVisitor) VisitFunction(fn model.GoFunction) error {
	// Currently we just collect functions in the package model
	// but we could do additional processing here
	return nil
}

// VisitConstant processes a constant declaration
func (v *HTMLVisitor) VisitConstant(c model.GoConstant) error {
	// Process constants if needed
	return nil
}

// VisitVariable processes a variable declaration
func (v *HTMLVisitor) VisitVariable(vr model.GoVariable) error {
	// Process variables if needed
	return nil
}

// VisitImport processes an import declaration
func (v *HTMLVisitor) VisitImport(imp model.GoImport) error {
	// Process imports if needed
	return nil
}

// Result generates the final HTML output
func (v *HTMLVisitor) Result() (string, error) {
	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"formatCode":    formatCode,
		"typeKindClass": typeKindClass,
		"formatDoc":     formatDocComment,
	}).Parse(htmlTemplate)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, v.packageData); err != nil {
		return "", err
	}

	return buf.String(), nil
}
