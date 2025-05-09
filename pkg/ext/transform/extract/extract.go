// Package extract provides transformers for extracting interfaces from implementations
// with type system awareness.
package extract

import (
	"bitspark.dev/go-tree/pkg/core/graph"
	"bitspark.dev/go-tree/pkg/ext/transform"
	"fmt"
	"sort"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// MethodPattern represents a pattern of methods that could form an interface
type MethodPattern struct {
	// The method signatures that form this pattern
	Methods []*typesys.Symbol

	// Types that implement this pattern
	ImplementingTypes []*typesys.Symbol

	// Generated interface name
	InterfaceName string

	// Package where the interface should be created
	TargetPackage *typesys.Package
}

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

// Transform implements the transform.Transformer interface
func (e *InterfaceExtractor) Transform(ctx *transform.Context) (*transform.TransformResult, error) {
	result := &transform.TransformResult{
		Summary:       "Extract common interfaces",
		Success:       false,
		IsDryRun:      ctx.DryRun,
		AffectedFiles: []string{},
		Changes:       []transform.Change{},
	}

	// Find common method patterns across types
	patterns, err := e.findMethodPatterns(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to find method patterns: %w", err)
		return result, result.Error
	}

	// Filter patterns based on options
	filteredPatterns := e.filterPatterns(patterns)

	// If no patterns found, return early
	if len(filteredPatterns) == 0 {
		result.Success = true
		result.Details = "No suitable interface patterns found"
		return result, nil
	}

	result.Details = fmt.Sprintf("Found %d interface patterns", len(filteredPatterns))

	// Generate and add interfaces for each pattern
	for _, pattern := range filteredPatterns {
		if err := e.createInterface(ctx, pattern, result); err != nil {
			result.Error = fmt.Errorf("failed to create interface: %w", err)
			return result, result.Error
		}
	}

	// If this is a dry run, we're done
	if ctx.DryRun {
		result.Success = true
		result.FilesAffected = len(result.AffectedFiles)
		return result, nil
	}

	// Update the index with changes
	if err := ctx.Index.Update(result.AffectedFiles); err != nil {
		result.Error = fmt.Errorf("failed to update index: %w", err)
		return result, result.Error
	}

	result.Success = true
	result.FilesAffected = len(result.AffectedFiles)
	return result, nil
}

// Validate implements the transform.Transformer interface
func (e *InterfaceExtractor) Validate(ctx *transform.Context) error {
	// Check that we have at least some types in the module
	typeCount := len(ctx.Index.FindSymbolsByKind(typesys.KindStruct))
	if typeCount == 0 {
		return fmt.Errorf("no struct types found in the module")
	}

	// Validate options
	if e.options.MinimumTypes < 1 {
		return fmt.Errorf("minimum types must be at least 1")
	}

	if e.options.MinimumMethods < 1 {
		return fmt.Errorf("minimum methods must be at least 1")
	}

	if e.options.MethodThreshold <= 0 || e.options.MethodThreshold > 1.0 {
		return fmt.Errorf("method threshold must be between 0 and 1")
	}

	// Check if target package exists if specified
	if e.options.TargetPackage != "" {
		targetPkg := findPackageByImportPath(ctx.Module, e.options.TargetPackage)
		if targetPkg == nil {
			return fmt.Errorf("target package '%s' not found", e.options.TargetPackage)
		}
	}

	return nil
}

// Name implements the transform.Transformer interface
func (e *InterfaceExtractor) Name() string {
	return "InterfaceExtractor"
}

// Description implements the transform.Transformer interface
func (e *InterfaceExtractor) Description() string {
	return "Extracts common interfaces from implementation types"
}

// findMethodPatterns identifies common method patterns across types
func (e *InterfaceExtractor) findMethodPatterns(ctx *transform.Context) ([]*MethodPattern, error) {
	// This will use the graph package to build a bipartite graph of types and methods
	g := graph.NewDirectedGraph()

	// Map of type ID to symbol
	typeSymbols := make(map[string]*typesys.Symbol)

	// Map of method signature to symbol
	methodSymbols := make(map[string]*typesys.Symbol)

	// Process all struct types
	structTypes := ctx.Index.FindSymbolsByKind(typesys.KindStruct)
	for _, typeSymbol := range structTypes {
		// Skip types in excluded packages
		if typeSymbol.Package != nil && e.options.IsExcludedPackage(typeSymbol.Package.ImportPath) {
			continue
		}

		// Skip excluded types
		if e.options.IsExcludedType(typeSymbol.Name) {
			continue
		}

		// Add type to graph
		typeID := typeSymbol.ID
		g.AddNode(typeID, typeSymbol)
		typeSymbols[typeID] = typeSymbol

		// Find methods for this type
		methods := ctx.Index.FindMethods(typeSymbol.Name)
		if len(methods) == 0 {
			continue
		}

		// Add methods to graph and connect to this type
		for _, method := range methods {
			// Skip excluded methods
			if e.options.IsExcludedMethod(method.Name) {
				continue
			}

			// Create a signature key
			signatureKey := fmt.Sprintf("%s-%s", method.Name, getMethodSignature(ctx, method))

			// Add method to graph
			g.AddNode(signatureKey, method)
			g.AddEdge(typeID, signatureKey, nil)

			// Store method symbol
			methodSymbols[signatureKey] = method
		}
	}

	// Use graph traversal to find types with common methods
	// For each method, find types that implement it
	methodToTypes := make(map[string][]*typesys.Symbol)
	for methodID := range methodSymbols {
		// Find all types that connect to this method
		// Use manual traversal using EdgeList
		var implementors []interface{}

		// Find all edges in the graph where this method is the target
		edges := g.EdgeList()
		for _, edge := range edges {
			// Check if the target is our method
			targetID, ok := edge.To.ID.(string)
			if ok && targetID == methodID {
				// Add the source of this edge (implementing type)
				implementors = append(implementors, edge.From.ID)
			}
		}

		// Add types to map
		var types []*typesys.Symbol
		for _, typeID := range implementors {
			if typeSymbol, ok := typeSymbols[typeID.(string)]; ok {
				types = append(types, typeSymbol)
			}
		}

		methodToTypes[methodID] = types
	}

	// Now create method patterns
	// Group methods by the types that implement them
	patternMap := make(map[string]*MethodPattern)

	// For types with multiple methods, create combination patterns
	for typeID, typeSymbol := range typeSymbols {
		// Get all methods for this type using EdgeList
		var methodIDs []string

		// Find all edges where this type is the source
		edges := g.EdgeList()
		for _, edge := range edges {
			// Check if the source is our type
			sourceID, ok := edge.From.ID.(string)
			if ok && sourceID == typeID {
				// Add the target (method) to the list
				if targetID, ok := edge.To.ID.(string); ok {
					methodIDs = append(methodIDs, targetID)
				}
			}
		}

		if len(methodIDs) < e.options.MinimumMethods {
			continue
		}

		// Sort methods for consistent key generation
		sort.Strings(methodIDs)

		// Generate a pattern key from the sorted method IDs
		patternKey := strings.Join(methodIDs, "|")

		// If pattern doesn't exist yet, create it
		if _, ok := patternMap[patternKey]; !ok {
			// Create list of method symbols
			var methods []*typesys.Symbol
			for _, methodID := range methodIDs {
				if methodSymbol, ok := methodSymbols[methodID]; ok {
					methods = append(methods, methodSymbol)
				}
			}

			// Create pattern with first implementing type
			patternMap[patternKey] = &MethodPattern{
				Methods:           methods,
				ImplementingTypes: []*typesys.Symbol{typeSymbol},
			}
		} else {
			// Add this type to existing pattern
			patternMap[patternKey].ImplementingTypes = append(
				patternMap[patternKey].ImplementingTypes, typeSymbol)
		}
	}

	// Convert map to slice of patterns
	var patterns []*MethodPattern
	for _, pattern := range patternMap {
		// Only include patterns with enough methods and types
		if len(pattern.Methods) >= e.options.MinimumMethods &&
			len(pattern.ImplementingTypes) >= e.options.MinimumTypes {
			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// filterPatterns filters and enhances method patterns
func (e *InterfaceExtractor) filterPatterns(patterns []*MethodPattern) []*MethodPattern {
	var filtered []*MethodPattern

	for _, pattern := range patterns {
		// Skip if doesn't meet minimums (should already be filtered, but check again)
		if len(pattern.Methods) < e.options.MinimumMethods ||
			len(pattern.ImplementingTypes) < e.options.MinimumTypes {
			continue
		}

		// Generate interface name
		pattern.InterfaceName = e.generateInterfaceName(pattern)

		// Select target package
		pattern.TargetPackage = e.selectTargetPackage(pattern)

		// Add to filtered list
		filtered = append(filtered, pattern)
	}

	return filtered
}

// createInterface creates an interface from a method pattern
func (e *InterfaceExtractor) createInterface(ctx *transform.Context, pattern *MethodPattern, result *transform.TransformResult) error {
	if pattern.TargetPackage == nil {
		return fmt.Errorf("no target package specified for interface %s", pattern.InterfaceName)
	}

	// Check if interface already exists
	existingSymbols := ctx.Index.FindSymbolsByName(pattern.InterfaceName)
	for _, sym := range existingSymbols {
		if sym.Kind == typesys.KindInterface && sym.Package == pattern.TargetPackage {
			// Interface already exists
			return nil
		}
	}

	// Determine which file to add the interface to
	var targetFile *typesys.File
	if e.options.CreateNewFiles {
		// Create a new file for the interface
		fileName := strings.ToLower(pattern.InterfaceName) + ".go"
		filePath := pattern.TargetPackage.Dir + "/" + fileName

		// Check if file already exists
		for _, file := range pattern.TargetPackage.Files {
			if file.Name == fileName {
				targetFile = file
				break
			}
		}

		if targetFile == nil {
			// Create new file - in a real implementation, we would create the actual file
			// For now, just simulate the file object
			targetFile = &typesys.File{
				Path:    filePath,
				Name:    fileName,
				Package: pattern.TargetPackage,
				// In a real implementation, would set up AST nodes
			}
			pattern.TargetPackage.Files[filePath] = targetFile
		}
	} else {
		// Use an existing file - preferably one that contains related types
		// First try to use a file from one of the implementing types
		if len(pattern.ImplementingTypes) > 0 && pattern.ImplementingTypes[0].File != nil {
			// Use the file of the first implementing type if it's in the target package
			if pattern.ImplementingTypes[0].Package == pattern.TargetPackage {
				targetFile = pattern.ImplementingTypes[0].File
			}
		}

		// If still no file, use the first non-test file in the package
		if targetFile == nil {
			for _, file := range pattern.TargetPackage.Files {
				if !strings.HasSuffix(file.Name, "_test.go") {
					targetFile = file
					break
				}
			}
		}
	}

	if targetFile == nil {
		return fmt.Errorf("could not find a suitable file to add interface %s", pattern.InterfaceName)
	}

	// Add the target file to affected files if not already there
	found := false
	for _, file := range result.AffectedFiles {
		if file == targetFile.Path {
			found = true
			break
		}
	}
	if !found {
		result.AffectedFiles = append(result.AffectedFiles, targetFile.Path)
	}

	// Build the interface source code
	var methodStrs []string
	for _, method := range pattern.Methods {
		signature := getMethodSignature(ctx, method)
		methodStrs = append(methodStrs, fmt.Sprintf("\t%s%s", method.Name, signature))
	}

	interfaceCode := fmt.Sprintf("type %s interface {\n%s\n}",
		pattern.InterfaceName, strings.Join(methodStrs, "\n"))

	// Add a comment
	interfaceCode = fmt.Sprintf("// %s is an interface extracted from %d implementing types.\n%s",
		pattern.InterfaceName, len(pattern.ImplementingTypes), interfaceCode)

	// Create a change record
	change := transform.Change{
		FilePath:  targetFile.Path,
		StartLine: 0, // Will be determined during actual insertion
		EndLine:   0,
		Original:  "",
		New:       interfaceCode,
	}
	result.Changes = append(result.Changes, change)

	// If this is a dry run, we're done
	if ctx.DryRun {
		return nil
	}

	// In a real implementation, we would update the AST and generate the new interface type
	// For this demonstration, we'll just create a new symbol and add it to the package

	// Create new interface symbol
	interfaceSymbol := &typesys.Symbol{
		ID:      "iface_" + pattern.InterfaceName, // Simplified ID generation
		Name:    pattern.InterfaceName,
		Kind:    typesys.KindInterface,
		File:    targetFile,
		Package: pattern.TargetPackage,
		// In a real implementation, would set correct positions
	}

	// Add symbol to file
	targetFile.Symbols = append(targetFile.Symbols, interfaceSymbol)

	// Add symbol to package
	pattern.TargetPackage.Symbols[interfaceSymbol.ID] = interfaceSymbol

	// In a real implementation, we would mark the file as modified
	// Since we don't have that field, just note that we would do it

	return nil
}

// generateInterfaceName generates a name for the interface
func (e *InterfaceExtractor) generateInterfaceName(pattern *MethodPattern) string {
	// If there's an explicit naming strategy, use it
	if e.options.NamingStrategy != nil {
		var methodNames []string
		for _, method := range pattern.Methods {
			methodNames = append(methodNames, method.Name)
		}
		return e.options.NamingStrategy(pattern.ImplementingTypes, methodNames)
	}

	// Default naming strategy
	// Try to find a common suffix in the implementing types (e.g., "Reader" in "FileReader", "BuffReader")
	commonSuffix := findCommonTypeSuffix(pattern.ImplementingTypes)
	if commonSuffix != "" {
		return commonSuffix
	}

	// Try to use a representative method name
	if len(pattern.Methods) > 0 {
		methodName := pattern.Methods[0].Name

		// Convert "Read" to "Reader"
		if methodName == "Read" {
			return "Reader"
		}
		// Convert "Write" to "Writer"
		if methodName == "Write" {
			return "Writer"
		}
		// Convert "Close" to "Closer"
		if methodName == "Close" {
			return "Closer"
		}
		// Convert other verbs to -er form
		if !strings.HasSuffix(methodName, "e") {
			return methodName + "er"
		}
		return methodName + "r"
	}

	// Fallback: use a generic name plus hash to ensure uniqueness
	return "Interface"
}

// selectTargetPackage selects the package where the interface should be created
func (e *InterfaceExtractor) selectTargetPackage(pattern *MethodPattern) *typesys.Package {
	if len(pattern.ImplementingTypes) == 0 {
		return nil
	}

	// If there's an explicit target package, use it
	if e.options.TargetPackage != "" {
		// Find the package by import path
		module := pattern.ImplementingTypes[0].Package.Module
		if module != nil {
			if pkg := findPackageByImportPath(module, e.options.TargetPackage); pkg != nil {
				return pkg
			}
		}
	}

	// Default strategy: use the package of the first implementing type
	return pattern.ImplementingTypes[0].Package
}

// Helper function to find common suffix among type names
func findCommonTypeSuffix(types []*typesys.Symbol) string {
	if len(types) == 0 {
		return ""
	}

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

// Helper function to get a method's signature
func getMethodSignature(ctx *transform.Context, method *typesys.Symbol) string {
	// For a full implementation, we would extract the method signature from the type system

	// Simplified signature generation
	// In a real implementation, we would use proper type resolution
	return fmt.Sprintf("(%s) %s", "args", "returnType")
}

// Helper function to find a package by import path
func findPackageByImportPath(mod *typesys.Module, importPath string) *typesys.Package {
	if mod == nil {
		return nil
	}

	for _, pkg := range mod.Packages {
		if pkg.ImportPath == importPath {
			return pkg
		}
	}

	return nil
}
