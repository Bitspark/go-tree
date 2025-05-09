package usage

import (
	"bitspark.dev/go-tree/pkg/ext/analyze"
	"fmt"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// DeadCodeOptions provides options for dead code detection.
type DeadCodeOptions struct {
	// IgnoreExported indicates whether to ignore exported symbols.
	IgnoreExported bool

	// IgnoreGenerated indicates whether to ignore generated files.
	IgnoreGenerated bool

	// IgnoreMain indicates whether to ignore main functions.
	IgnoreMain bool

	// IgnoreTests indicates whether to ignore test files.
	IgnoreTests bool

	// ExcludedPackages is a list of packages to exclude from analysis.
	ExcludedPackages []string

	// ConsiderReflection indicates whether to consider potential reflection usage.
	ConsiderReflection bool
}

// DefaultDeadCodeOptions returns the default options for dead code detection.
func DefaultDeadCodeOptions() *DeadCodeOptions {
	return &DeadCodeOptions{
		IgnoreExported:     true,
		IgnoreGenerated:    true,
		IgnoreMain:         true,
		IgnoreTests:        true,
		ExcludedPackages:   nil,
		ConsiderReflection: true,
	}
}

// DeadSymbol represents an unused symbol with context.
type DeadSymbol struct {
	// The unused symbol
	Symbol *typesys.Symbol

	// Reason explains why the symbol is considered unused
	Reason string

	// Confidence level (0-100) of the dead code detection
	Confidence int
}

// DeadCodeAnalyzer analyzes code for unused symbols.
type DeadCodeAnalyzer struct {
	*analyze.BaseAnalyzer
	Module    *typesys.Module
	Collector *UsageCollector
}

// NewDeadCodeAnalyzer creates a new dead code analyzer.
func NewDeadCodeAnalyzer(module *typesys.Module) *DeadCodeAnalyzer {
	return &DeadCodeAnalyzer{
		BaseAnalyzer: analyze.NewBaseAnalyzer(
			"DeadCodeAnalyzer",
			"Analyzes code for unused symbols",
		),
		Module:    module,
		Collector: NewUsageCollector(module),
	}
}

// FindDeadCode identifies unused symbols in the module.
func (a *DeadCodeAnalyzer) FindDeadCode(opts *DeadCodeOptions) ([]*DeadSymbol, error) {
	if a.Module == nil {
		return nil, fmt.Errorf("module is nil")
	}

	if opts == nil {
		opts = DefaultDeadCodeOptions()
	}

	// Collect usage information for all symbols
	usages, err := a.Collector.CollectUsageForAllSymbols()
	if err != nil {
		return nil, err
	}

	// Find unused symbols
	var deadSymbols []*DeadSymbol

	// Check each package for unused symbols
	for _, pkg := range a.Module.Packages {
		// Skip excluded packages
		if isExcludedPackage(pkg.ImportPath, opts.ExcludedPackages) {
			continue
		}

		// Skip test files if configured
		if opts.IgnoreTests && isTestPackage(pkg) {
			continue
		}

		// Check each symbol in the package
		for _, sym := range pkg.Symbols {
			if isUnused(sym, usages, opts) {
				reason, confidence := determineUnusedReason(sym, opts)
				deadSymbols = append(deadSymbols, &DeadSymbol{
					Symbol:     sym,
					Reason:     reason,
					Confidence: confidence,
				})
			}
		}
	}

	return deadSymbols, nil
}

// isUnused determines if a symbol is unused.
func isUnused(sym *typesys.Symbol, usages map[string]*SymbolUsage, opts *DeadCodeOptions) bool {
	// Skip if it doesn't need analysis
	if !needsAnalysis(sym, opts) {
		return false
	}

	// Get symbol ID
	symID := getSymbolID(sym)

	// Check if we have usage information
	usage, found := usages[symID]
	if !found {
		// No usage information, but we should have some if it's used
		return true
	}

	// A symbol is used if it has references or if it's defined but not referenced
	// The latter case handles entry points and other special cases
	return usage.GetReferenceCount() == 0 && !isEntryPoint(sym, opts)
}

// determineUnusedReason provides a reason why a symbol is considered unused.
func determineUnusedReason(sym *typesys.Symbol, opts *DeadCodeOptions) (string, int) {
	// Base confidence level
	confidence := 90

	// Check for potential reflection usage
	if opts.ConsiderReflection && mightBeUsedViaReflection(sym) {
		confidence = 60
		return "No direct references, but might be used via reflection", confidence
	}

	// Default reason based on symbol kind
	switch sym.Kind {
	case typesys.KindFunction, typesys.KindMethod:
		return "Function is never called", confidence
	case typesys.KindType, typesys.KindStruct, typesys.KindInterface:
		return "Type is never used", confidence
	case typesys.KindVariable:
		return "Variable is never read or written to", confidence
	case typesys.KindConstant:
		return "Constant is never used", confidence
	default:
		return "Symbol is never referenced", confidence
	}
}

// needsAnalysis determines if a symbol needs to be analyzed for dead code.
func needsAnalysis(sym *typesys.Symbol, opts *DeadCodeOptions) bool {
	// Skip if the symbol is nil
	if sym == nil {
		return false
	}

	// Skip based on various options
	if opts.IgnoreExported && sym.Exported {
		return false
	}

	// Skip main function if configured
	if opts.IgnoreMain && isMainFunction(sym) {
		return false
	}

	// Skip generated file symbols if configured
	if opts.IgnoreGenerated && isGenerated(sym) {
		return false
	}

	// Skip symbols that are typically not considered dead code
	if isSpecialSymbol(sym) {
		return false
	}

	return true
}

// isEntryPoint determines if a symbol is an entry point.
func isEntryPoint(sym *typesys.Symbol, opts *DeadCodeOptions) bool {
	// Main function is an entry point
	if isMainFunction(sym) {
		return true
	}

	// Init functions are entry points
	if isInitFunction(sym) {
		return true
	}

	// Test functions are entry points if we're not ignoring tests
	if !opts.IgnoreTests && isTestFunction(sym) {
		return true
	}

	return false
}

// mightBeUsedViaReflection checks if a symbol might be used via reflection.
func mightBeUsedViaReflection(sym *typesys.Symbol) bool {
	// Exported struct fields are common targets for reflection
	if sym.Kind == typesys.KindField && sym.Exported {
		return true
	}

	// Exported methods on structs might be called via reflection
	if sym.Kind == typesys.KindMethod && sym.Exported {
		return true
	}

	// Types with JSON, XML, or YAML tags are likely used via reflection
	if hasSerializationTags(sym) {
		return true
	}

	return false
}

// Helper functions

// isExcludedPackage checks if a package is in the excluded list.
func isExcludedPackage(pkgPath string, excludedPackages []string) bool {
	for _, excluded := range excludedPackages {
		if pkgPath == excluded {
			return true
		}
	}
	return false
}

// isTestPackage checks if a package is a test package.
func isTestPackage(pkg *typesys.Package) bool {
	// In Go, test packages end with _test
	return pkg != nil && len(pkg.Name) > 5 && pkg.Name[len(pkg.Name)-5:] == "_test"
}

// isMainFunction checks if a symbol is the main function.
func isMainFunction(sym *typesys.Symbol) bool {
	return sym != nil && sym.Kind == typesys.KindFunction &&
		sym.Name == "main" && sym.Package != nil &&
		sym.Package.Name == "main"
}

// isInitFunction checks if a symbol is an init function.
func isInitFunction(sym *typesys.Symbol) bool {
	return sym != nil && sym.Kind == typesys.KindFunction && sym.Name == "init"
}

// isTestFunction checks if a symbol is a test function.
func isTestFunction(sym *typesys.Symbol) bool {
	return sym != nil && sym.Kind == typesys.KindFunction &&
		len(sym.Name) > 4 && sym.Name[:4] == "Test" &&
		sym.Name[4:5] == strings.ToUpper(sym.Name[4:5])
}

// isGenerated checks if a symbol is in a generated file.
func isGenerated(sym *typesys.Symbol) bool {
	// Check if the symbol is in a generated file
	if sym == nil || sym.File == nil {
		return false
	}

	// In Go, generated files often have a comment with "DO NOT EDIT"
	// A proper implementation would check file comments
	return false
}

// isSpecialSymbol checks if a symbol has special meaning and shouldn't be considered dead.
func isSpecialSymbol(sym *typesys.Symbol) bool {
	// Standard library symbols, type aliases, or embedded fields might have special handling
	return false
}

// hasSerializationTags checks if a struct type has JSON, XML, or YAML tags.
func hasSerializationTags(sym *typesys.Symbol) bool {
	// This would check struct field tags in a real implementation
	return false
}
