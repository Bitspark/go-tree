// Package model defines the core structures used for representing Go packages
package model

// GoPackage represents a complete Go package with all its components
type GoPackage struct {
	Name          string       `json:"name"`          // Package name
	Imports       []GoImport   `json:"imports"`       // Imported packages
	Functions     []GoFunction `json:"functions"`     // Top-level functions and methods
	Types         []GoType     `json:"types"`         // Type definitions (structs, interfaces, aliases, etc.)
	Constants     []GoConstant `json:"constants"`     // Top-level constants
	Variables     []GoVariable `json:"variables"`     // Top-level variables
	PackageDoc    string       `json:"packageDoc"`    // Package documentation comment, if any
	LicenseHeader string       `json:"licenseHeader"` // License/header comments before package, if any
}

// GoImport represents an imported package
type GoImport struct {
	Path    string `json:"path"`    // import path
	Alias   string `json:"alias"`   // local alias or "_" if used, empty if none
	Comment string `json:"comment"` // trailing comment on the import line, if any
	Doc     string `json:"doc"`     // documentation comment above the import spec, if any
}

// GoFunction represents a function or method declaration
type GoFunction struct {
	Name      string      `json:"name"`
	Receiver  *GoReceiver `json:"receiver,omitempty"` // method receiver, or nil if function
	Signature string      `json:"signature"`          // function signature (params and results)
	Body      string      `json:"body"`               // function body code (inside braces)
	Code      string      `json:"code"`               // full function source code (including signature and body)
	Doc       string      `json:"doc"`                // documentation comment above the function, if any
}

// GoReceiver represents a method receiver
type GoReceiver struct {
	Name string `json:"name"` // receiver name (may be empty if omitted)
	Type string `json:"type"` // receiver type (e.g. "T" or "*T")
}

// GoType represents a type declaration
type GoType struct {
	Name             string     `json:"name"`
	Kind             string     `json:"kind"`              // "struct", "interface", "alias", or "type" (for other definitions)
	AliasOf          string     `json:"aliasOf,omitempty"` // target type if Kind == "alias"
	UnderlyingType   string     `json:"underlyingType"`    // underlying type (for non-alias; "struct" or "interface" or literal type)
	Fields           []GoField  `json:"fields,omitempty"`  // fields if struct
	InterfaceMethods []GoMethod `json:"methods,omitempty"` // methods if interface (includes embedded interfaces as entries with empty Signature)
	Code             string     `json:"code"`              // full type declaration source code
	Doc              string     `json:"doc"`               // documentation comment above the type, if any
}

// GoField represents a field in a struct type
type GoField struct {
	Name    string `json:"name"`    // field name (empty for embedded field)
	Type    string `json:"type"`    // field type
	Tag     string `json:"tag"`     // struct tag string, if any (including quotes)
	Comment string `json:"comment"` // trailing comment for this field, if any
	Doc     string `json:"doc"`     // documentation comment above this field, if any
}

// GoMethod represents a method in an interface type
type GoMethod struct {
	Name      string `json:"name"`      // method name (or embedded interface type name)
	Signature string `json:"signature"` // method signature (empty if embedded interface)
	Comment   string `json:"comment"`   // trailing comment, if any
	Doc       string `json:"doc"`       // documentation comment above method, if any
}

// GoConstant represents a constant declaration
type GoConstant struct {
	Name    string `json:"name"`
	Type    string `json:"type"`    // type of the constant, if specified
	Value   string `json:"value"`   // constant value/expression (as written, empty if implicit)
	Doc     string `json:"doc"`     // documentation comment above, if any
	Comment string `json:"comment"` // trailing comment, if any
}

// GoVariable represents a variable declaration
type GoVariable struct {
	Name    string `json:"name"`
	Type    string `json:"type"`    // type of the variable, if specified
	Value   string `json:"value"`   // initial value, if any
	Doc     string `json:"doc"`     // documentation comment above, if any
	Comment string `json:"comment"` // trailing comment, if any
}

// Declaration represents a general declaration with position information for ordering
type Declaration struct {
	Type     string      `json:"type"`     // Type of declaration: "const", "var", "type", "func", or "comment"
	Position int         `json:"position"` // Source position for ordering
	Data     interface{} `json:"data"`     // The actual declaration data (one of the Go* types or string for comment)
}
