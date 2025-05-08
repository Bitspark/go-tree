package typesys

import (
	"fmt"
	"go/token"
	"go/types"
)

// SymbolKind represents the kind of a symbol in the code.
type SymbolKind int

const (
	KindUnknown   SymbolKind = iota
	KindPackage              // Package
	KindFunction             // Function
	KindMethod               // Method (function with receiver)
	KindType                 // Named type (struct, interface, etc.)
	KindVariable             // Variable
	KindConstant             // Constant
	KindField                // Struct field
	KindParameter            // Function parameter
	KindInterface            // Interface type
	KindStruct               // Struct type
	KindImport               // Import declaration
	KindLabel                // Label
)

// String returns a string representation of the symbol kind.
func (k SymbolKind) String() string {
	switch k {
	case KindPackage:
		return "package"
	case KindFunction:
		return "function"
	case KindMethod:
		return "method"
	case KindType:
		return "type"
	case KindVariable:
		return "variable"
	case KindConstant:
		return "constant"
	case KindField:
		return "field"
	case KindParameter:
		return "parameter"
	case KindInterface:
		return "interface"
	case KindStruct:
		return "struct"
	case KindImport:
		return "import"
	case KindLabel:
		return "label"
	default:
		return "unknown"
	}
}

// Symbol represents any named entity in Go code.
type Symbol struct {
	// Identity
	ID       string     // Unique identifier
	Name     string     // Name of the symbol
	Kind     SymbolKind // Type of symbol
	Exported bool       // Whether the symbol is exported

	// Type information
	TypeObj  types.Object // Go's type object
	TypeInfo types.Type   // Type information if applicable

	// Structure information
	Parent  *Symbol  // For methods, fields, etc.
	Package *Package // Package containing the symbol
	File    *File    // File containing the symbol

	// Position
	Pos token.Pos // Start position
	End token.Pos // End position

	// References
	Definitions []*Position  // Where this symbol is defined
	References  []*Reference // All references to this symbol
}

// Position represents a position in a file.
type Position struct {
	File   string    // File path
	Pos    token.Pos // Position
	Line   int       // Line number (1-based)
	Column int       // Column number (1-based)
}

// NewSymbol creates a new symbol with the given name and kind.
func NewSymbol(name string, kind SymbolKind) *Symbol {
	return &Symbol{
		ID:          GenerateSymbolID(name, kind),
		Name:        name,
		Kind:        kind,
		Exported:    isExported(name),
		Definitions: make([]*Position, 0),
		References:  make([]*Reference, 0),
	}
}

// AddReference adds a reference to this symbol.
func (s *Symbol) AddReference(ref *Reference) {
	s.References = append(s.References, ref)
}

// AddDefinition adds a definition position for this symbol.
func (s *Symbol) AddDefinition(file string, pos token.Pos, line, column int) {
	s.Definitions = append(s.Definitions, &Position{
		File:   file,
		Pos:    pos,
		Line:   line,
		Column: column,
	})
}

// GetPosition returns position information for this symbol.
func (s *Symbol) GetPosition() *PositionInfo {
	if s.File == nil {
		return nil
	}
	return s.File.GetPositionInfo(s.Pos, s.End)
}

// GenerateSymbolID creates a unique ID for a symbol.
func GenerateSymbolID(name string, kind SymbolKind) string {
	// Simple implementation for now
	return fmt.Sprintf("%s:%d", name, kind)
}

// isExported checks if a name is exported (starts with uppercase).
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	// In Go, exported names start with an uppercase letter
	return name[0] >= 'A' && name[0] <= 'Z'
}
