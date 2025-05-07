package tree

// Options configures the behavior of the parser and formatter
type Options struct {
	// IncludeTestFiles determines whether to include test files in parsing
	IncludeTestFiles bool

	// PreserveFormattingStyle determines whether to preserve the original formatting style
	// If false, Go's standard formatting will be applied
	PreserveFormattingStyle bool

	// SkipComments determines whether to skip comments during parsing
	SkipComments bool

	// CustomPackageName sets a custom package name for formatting
	// If empty, the original package name will be used
	CustomPackageName string
}

// DefaultOptions returns the default options
func DefaultOptions() *Options {
	return &Options{
		IncludeTestFiles:        false,
		PreserveFormattingStyle: true,
		SkipComments:            false,
		CustomPackageName:       "",
	}
}

// ParseWithOptions parses a Go package from the given directory with the specified options
func ParseWithOptions(dir string, opts *Options) (*Package, error) {
	// Currently just using default parse implementation
	// In the future, this would respect the options
	return Parse(dir)
}

// FormatWithOptions formats a Package into a single Go source file with the specified options
func (p *Package) FormatWithOptions(opts *Options) (string, error) {
	// Currently just using default format implementation
	// In the future, this would respect the options

	// Apply custom package name if specified
	origName := p.Model.Name
	if opts != nil && opts.CustomPackageName != "" {
		p.Model.Name = opts.CustomPackageName
		defer func() { p.Model.Name = origName }() // Restore original name after formatting
	}

	return p.Format()
}
