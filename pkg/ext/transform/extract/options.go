// Package extract provides transformers for extracting interfaces from implementations
// with type system awareness.
package extract

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
)

// NamingStrategy is a function that generates interface names based on implementing types
type NamingStrategy func(types []*typesys.Symbol, methodNames []string) string

// Options configures the behavior of the interface extractor
type Options struct {
	// Minimum number of types that must implement a method pattern
	MinimumTypes int

	// Minimum number of methods required for an interface
	MinimumMethods int

	// Threshold for method similarity (percentage of methods that must match)
	MethodThreshold float64

	// Strategy for naming generated interfaces
	NamingStrategy NamingStrategy

	// Package where interfaces should be created (import path)
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

// IsExcludedPackage checks if a package is in the exclude list
func (o *Options) IsExcludedPackage(importPath string) bool {
	for _, excluded := range o.ExcludePackages {
		if excluded == importPath {
			return true
		}
	}
	return false
}

// IsExcludedType checks if a type is in the exclude list
func (o *Options) IsExcludedType(typeName string) bool {
	for _, excluded := range o.ExcludeTypes {
		if excluded == typeName {
			return true
		}
	}
	return false
}

// IsExcludedMethod checks if a method is in the exclude list
func (o *Options) IsExcludedMethod(methodName string) bool {
	for _, excluded := range o.ExcludeMethods {
		if excluded == methodName {
			return true
		}
	}
	return false
}
