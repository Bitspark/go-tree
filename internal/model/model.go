// Package model defines the core structures used for representing Go packages
package model

// GoPackage represents a complete Go package with all its components
type GoPackage struct {
	Name          string       // Package name
	Imports       []GoImport   // Imported packages
	Functions     []GoFunction // Top-level functions and methods
	Types         []GoType     // Type definitions (structs, interfaces, aliases, etc.)
	Constants     []GoConstant // Top-level constants
	Variables     []GoVariable // Top-level variables
	PackageDoc    string       // Package documentation comment, if any
	LicenseHeader string       // License/header comments before package, if any
}

// GoImport represents an imported package
type GoImport struct {
	Path    string // import path
	Alias   string // local alias or "_" if used, empty if none
	Comment string // trailing comment on the import line, if any
	Doc     string // documentation comment above the import spec, if any
}

// GoFunction represents a function or method declaration
type GoFunction struct {
	Name      string
	Receiver  *GoReceiver // method receiver, or nil if function
	Signature string      // function signature (params and results)
	Body      string      // function body code (inside braces)
	Code      string      // full function source code (including signature and body)
	Doc       string      // documentation comment above the function, if any
}

// GoReceiver represents a method receiver
type GoReceiver struct {
	Name string // receiver name (may be empty if omitted)
	Type string // receiver type (e.g. "T" or "*T")
}

// GoType represents a type declaration
type GoType struct {
	Name             string
	Kind             string     // "struct", "interface", "alias", or "type" (for other definitions)
	AliasOf          string     // target type if Kind == "alias"
	UnderlyingType   string     // underlying type (for non-alias; "struct" or "interface" or literal type)
	Fields           []GoField  // fields if struct
	InterfaceMethods []GoMethod // methods if interface (includes embedded interfaces as entries with empty Signature)
	Code             string     // full type declaration source code
	Doc              string     // documentation comment above the type, if any
}

// GoField represents a field in a struct type
type GoField struct {
	Name    string // field name (empty for embedded field)
	Type    string // field type
	Tag     string // struct tag string, if any (including quotes)
	Comment string // trailing comment for this field, if any
	Doc     string // documentation comment above this field, if any
}

// GoMethod represents a method in an interface type
type GoMethod struct {
	Name      string // method name (or embedded interface type name)
	Signature string // method signature (empty if embedded interface)
	Comment   string // trailing comment, if any
	Doc       string // documentation comment above method, if any
}

// GoConstant represents a constant declaration
type GoConstant struct {
	Name    string
	Type    string // type of the constant, if specified
	Value   string // constant value/expression (as written, empty if implicit)
	Doc     string // documentation comment above, if any
	Comment string // trailing comment, if any
}

// GoVariable represents a variable declaration
type GoVariable struct {
	Name    string
	Type    string // type of the variable, if specified
	Value   string // initial value, if any
	Doc     string // documentation comment above, if any
	Comment string // trailing comment, if any
}

// Declaration represents a general declaration with position information for ordering
type Declaration struct {
	Type     string      // Type of declaration: "const", "var", "type", "func", or "comment"
	Position int         // Source position for ordering
	Data     interface{} // The actual declaration data (one of the Go* types or string for comment)
}
