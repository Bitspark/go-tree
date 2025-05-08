// Package extract provides transformers for extracting interfaces from implementations.
package extract

import (
	"bitspark.dev/go-tree/pkgold/core/module"
)

// NamingStrategy is a function that generates interface names
type NamingStrategy func(types []*module.Type, signatures []string) string

// Options configures the behavior of the interface extractor
type Options struct {
	// Minimum number of types that must implement a method pattern
	MinimumTypes int

	// Minimum number of methods required for an interface
	MinimumMethods int

	// Threshold for method overlap (percentage of methods that must match)
	MethodThreshold float64

	// Strategy for naming generated interfaces
	NamingStrategy NamingStrategy

	// Package where interfaces should be created
	TargetPackage string

	// Whether to create new files for interfaces
	CreateNewFiles bool

	// Packages to exclude from analysis
	ExcludePackages []string

	// Types to exclude from analysis
	ExcludeTypes []string

	// Methods to exclude from analysis
	ExcludeMethods []string
}

// DefaultOptions returns the default options for interface extraction
func DefaultOptions() Options {
	return Options{
		MinimumTypes:    2,
		MinimumMethods:  1,
		MethodThreshold: 0.8,
		NamingStrategy:  nil, // Use default naming
		CreateNewFiles:  false,
		ExcludePackages: []string{},
		ExcludeTypes:    []string{},
		ExcludeMethods:  []string{},
	}
}
