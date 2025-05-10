package saver

// SaveOptions defines options for module saving
type SaveOptions struct {
	// Whether to format the code
	Format bool

	// Whether to organize imports
	OrganizeImports bool

	// Whether to generate gofmt-compatible output
	Gofmt bool

	// Whether to use tabs (true) or spaces (false) for indentation
	UseTabs bool

	// The number of spaces per indentation level (if UseTabs=false)
	TabWidth int

	// Force overwrite existing files
	Force bool

	// Whether to create a backup of modified files
	CreateBackups bool

	// Save only modified files (track modifications)
	OnlyModified bool

	// Mode for handling AST reconstruction
	ASTMode ASTReconstructionMode
}

// ASTReconstructionMode defines how to handle AST reconstruction
type ASTReconstructionMode int

const (
	// PreserveOriginal tries to preserve as much of the original formatting as possible
	PreserveOriginal ASTReconstructionMode = iota

	// ReformatAll completely reformats the code using go/printer
	ReformatAll

	// SmartMerge uses original formatting for unchanged code and standard formatting for new/modified code
	SmartMerge
)

// DefaultSaveOptions returns the default save options
func DefaultSaveOptions() SaveOptions {
	return SaveOptions{
		Format:          true,
		OrganizeImports: true,
		Gofmt:           true,
		UseTabs:         true,
		TabWidth:        8,
		Force:           false,
		CreateBackups:   false,
		OnlyModified:    false,
		ASTMode:         SmartMerge,
	}
}
