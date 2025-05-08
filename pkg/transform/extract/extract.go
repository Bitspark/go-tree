// Package extract provides transformers for extracting interfaces from implementations.
package extract

import (
	"fmt"
	"strings"

	"bitspark.dev/go-tree/pkg/core/module"
)

// InterfaceExtractor extracts interfaces from implementations
type InterfaceExtractor struct {
	options Options
}

// NewInterfaceExtractor creates a new interface extractor with the given options
func NewInterfaceExtractor(options Options) *InterfaceExtractor {
	return &InterfaceExtractor{
		options: options,
	}
}

// Transform implements the ModuleTransformer interface
func (e *InterfaceExtractor) Transform(mod *module.Module) error {
	// Find common method patterns across types
	methodPatterns := e.findMethodPatterns(mod)

	// Filter patterns based on options
	filteredPatterns := e.filterPatterns(methodPatterns)

	// Generate and add interfaces for each pattern
	for _, pattern := range filteredPatterns {
		if err := e.createInterface(mod, pattern); err != nil {
			return fmt.Errorf("failed to create interface: %w", err)
		}
	}

	return nil
}

// Name returns the name of the transformer
func (e *InterfaceExtractor) Name() string {
	return "InterfaceExtractor"
}

// Description returns a description of what the transformer does
func (e *InterfaceExtractor) Description() string {
	return "Extracts common interfaces from implementation types"
}

// MethodPattern represents a pattern of methods that could form an interface
type MethodPattern struct {
	// The method signatures that form this pattern
	Signatures []string

	// Types that implement this pattern
	ImplementingTypes []*module.Type

	// Generated interface name
	InterfaceName string

	// Package where the interface should be created
	TargetPackage *module.Package
}

// findMethodPatterns identifies common method patterns across types
func (e *InterfaceExtractor) findMethodPatterns(mod *module.Module) []*MethodPattern {
	// Map of method signature sets to types implementing them
	patternMap := make(map[string][]*module.Type)

	// Process all packages
	for _, pkg := range mod.Packages {
		// Skip packages in the exclude list
		if e.isExcludedPackage(pkg.ImportPath) {
			continue
		}

		// Process each type in the package
		for _, typ := range pkg.Types {
			// Only consider struct types that have methods
			if typ.Kind != "struct" || len(typ.Methods) == 0 {
				continue
			}

			// Skip types in the exclude list
			if e.isExcludedType(typ.Name) {
				continue
			}

			// Create a signature set for this type's methods
			var signatures []string
			for _, method := range typ.Methods {
				// Skip excluded methods
				if e.isExcludedMethod(method.Name) {
					continue
				}

				// Add method signature to set
				signatures = append(signatures, method.Name+method.Signature)
			}

			// If we have methods to consider
			if len(signatures) > 0 {
				// Sort signatures for consistent key generation
				// (In a real implementation, we would sort here)

				// Generate a key from the signatures
				key := strings.Join(signatures, "|")

				// Add this type to the pattern map
				patternMap[key] = append(patternMap[key], typ)
			}
		}
	}

	// Convert the map to a list of patterns
	var patterns []*MethodPattern
	for sigKey, types := range patternMap {
		// Only consider patterns implemented by multiple types
		if len(types) < e.options.MinimumTypes {
			continue
		}

		signatures := strings.Split(sigKey, "|")

		// Create pattern
		pattern := &MethodPattern{
			Signatures:        signatures,
			ImplementingTypes: types,
			// Interface name will be generated later
			// Target package will be selected later
		}

		patterns = append(patterns, pattern)
	}

	return patterns
}

// filterPatterns filters method patterns based on options
func (e *InterfaceExtractor) filterPatterns(patterns []*MethodPattern) []*MethodPattern {
	var filtered []*MethodPattern

	for _, pattern := range patterns {
		// Skip patterns with too few methods
		if len(pattern.Signatures) < e.options.MinimumMethods {
			continue
		}

		// Skip patterns with too few implementing types
		if len(pattern.ImplementingTypes) < e.options.MinimumTypes {
			continue
		}

		// Generate interface name
		pattern.InterfaceName = e.generateInterfaceName(pattern)

		// Select target package
		pattern.TargetPackage = e.selectTargetPackage(pattern)

		filtered = append(filtered, pattern)
	}

	return filtered
}

// createInterface creates an interface for a method pattern
func (e *InterfaceExtractor) createInterface(mod *module.Module, pattern *MethodPattern) error {
	// Check if interface already exists
	for _, existingType := range pattern.TargetPackage.Types {
		if existingType.Name == pattern.InterfaceName && existingType.Kind == "interface" {
			// Interface already exists, potentially update it
			return nil
		}
	}

	// Create new interface type
	interfaceType := module.NewType(pattern.InterfaceName, "interface", true)

	// Add methods to interface
	for _, sig := range pattern.Signatures {
		// In a real implementation, we would parse the signature to extract name and signature
		// This is simplified for the example
		methodName := strings.Split(sig, "(")[0]
		methodSignature := sig[len(methodName):]

		interfaceType.AddInterfaceMethod(methodName, methodSignature, false, "")
	}

	// Generate documentation
	interfaceType.Doc = fmt.Sprintf("%s is an interface extracted from %d implementing types.",
		pattern.InterfaceName, len(pattern.ImplementingTypes))

	// Create a file for the interface if needed
	var file *module.File
	if e.options.CreateNewFiles {
		// Create a new file for the interface
		fileName := strings.ToLower(pattern.InterfaceName) + ".go"
		file = module.NewFile(
			pattern.TargetPackage.Dir+"/"+fileName,
			fileName,
			false,
		)
		pattern.TargetPackage.AddFile(file)
	} else {
		// Add to an existing file
		// Simplified: use the first type's file
		if len(pattern.ImplementingTypes) > 0 {
			file = pattern.ImplementingTypes[0].File
		} else {
			// If we can't find a suitable file, use the first file in the package
			for _, f := range pattern.TargetPackage.Files {
				if !f.IsTest {
					file = f
					break
				}
			}
		}
	}

	// If we have a file, add the interface to it
	if file != nil {
		file.AddType(interfaceType)
	}

	// Add the interface to the package
	pattern.TargetPackage.AddType(interfaceType)

	return nil
}

// generateInterfaceName generates a name for the interface
func (e *InterfaceExtractor) generateInterfaceName(pattern *MethodPattern) string {
	// If there's an explicit naming strategy, use it
	if e.options.NamingStrategy != nil {
		return e.options.NamingStrategy(pattern.ImplementingTypes, pattern.Signatures)
	}

	// Default naming strategy
	// For this simple example, use a common prefix if it exists, otherwise use methods
	if len(pattern.ImplementingTypes) > 0 {
		// Try to find a common suffix (like "Reader" in "FileReader", "BuffReader")
		commonSuffix := findCommonTypeSuffix(pattern.ImplementingTypes)
		if commonSuffix != "" {
			return commonSuffix
		}

		// Try to use a representative method name
		if len(pattern.Signatures) > 0 {
			methodName := strings.Split(pattern.Signatures[0], "(")[0]
			// Convert "Read" to "Reader"
			if methodName == "Read" {
				return "Reader"
			}
			// Convert "Write" to "Writer"
			if methodName == "Write" {
				return "Writer"
			}
			// Convert other verbs to -er form
			if !strings.HasSuffix(methodName, "e") {
				return methodName + "er"
			}
			return methodName + "r"
		}
	}

	// Fallback to a generic name
	return "Common"
}

// selectTargetPackage selects the package where the interface should be created
func (e *InterfaceExtractor) selectTargetPackage(pattern *MethodPattern) *module.Package {
	// If there's an explicit target package, use it
	if e.options.TargetPackage != "" {
		for _, pkg := range pattern.ImplementingTypes[0].Package.Module.Packages {
			if pkg.ImportPath == e.options.TargetPackage {
				return pkg
			}
		}
	}

	// Default strategy: use the package of the first implementing type
	return pattern.ImplementingTypes[0].Package
}

// isExcludedPackage checks if a package is in the exclude list
func (e *InterfaceExtractor) isExcludedPackage(importPath string) bool {
	for _, excluded := range e.options.ExcludePackages {
		if excluded == importPath {
			return true
		}
	}
	return false
}

// isExcludedType checks if a type is in the exclude list
func (e *InterfaceExtractor) isExcludedType(typeName string) bool {
	for _, excluded := range e.options.ExcludeTypes {
		if excluded == typeName {
			return true
		}
	}
	return false
}

// isExcludedMethod checks if a method is in the exclude list
func (e *InterfaceExtractor) isExcludedMethod(methodName string) bool {
	for _, excluded := range e.options.ExcludeMethods {
		if excluded == methodName {
			return true
		}
	}
	return false
}

// findCommonTypeSuffix finds a common suffix among type names
func findCommonTypeSuffix(types []*module.Type) string {
	if len(types) == 0 {
		return ""
	}

	// This is a simplified implementation
	// In a real implementation, we would use a more sophisticated algorithm

	// Check for common suffixes like "Reader", "Writer", "Handler", etc.
	commonSuffixes := []string{"Reader", "Writer", "Handler", "Processor", "Service", "Controller"}

	for _, suffix := range commonSuffixes {
		matches := 0
		for _, t := range types {
			if strings.HasSuffix(t.Name, suffix) {
				matches++
			}
		}

		// If more than half of the types have this suffix, use it
		if float64(matches)/float64(len(types)) >= 0.5 {
			return suffix
		}
	}

	return ""
}
