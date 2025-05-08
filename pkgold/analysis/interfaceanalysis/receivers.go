package interfaceanalysis

import (
	"strings"

	"bitspark.dev/go-tree/pkgold/core/module"
)

// Analyzer for method receiver analysis
type Analyzer struct{}

// NewAnalyzer creates a new method receiver analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzeReceivers analyzes all method receivers in a package and groups them by receiver type
func (a *Analyzer) AnalyzeReceivers(pkg *module.Package) *ReceiverAnalysis {
	analysis := &ReceiverAnalysis{
		Package: pkg.Name,
		Groups:  make(map[string]*ReceiverGroup),
	}

	// Process all functions in the package
	for _, fn := range pkg.Functions {
		// Skip functions without receivers (not methods)
		if fn.Receiver == nil {
			continue
		}

		// Get the receiver type and normalize it
		receiverType := fn.Receiver.Type
		baseType := normalizeReceiverType(receiverType)
		isPointer := strings.HasPrefix(receiverType, "*")

		// Get or create a group for this receiver type
		group, exists := analysis.Groups[receiverType]
		if !exists {
			group = &ReceiverGroup{
				ReceiverType: receiverType,
				BaseType:     baseType,
				IsPointer:    isPointer,
				Methods:      []*module.Function{},
			}
			analysis.Groups[receiverType] = group
		}

		// Add the method to the group
		group.Methods = append(group.Methods, fn)
	}

	return analysis
}

// CreateSummary creates a summary of receiver usage in the package
func (a *Analyzer) CreateSummary(analysis *ReceiverAnalysis) *ReceiverSummary {
	summary := &ReceiverSummary{
		TotalMethods:       0,
		TotalReceiverTypes: len(analysis.Groups),
		MethodsPerType:     make(map[string]int),
		PointerReceivers:   0,
		ValueReceivers:     0,
	}

	// Count methods and categorize by receiver type
	for receiverType, group := range analysis.Groups {
		methodCount := len(group.Methods)
		summary.TotalMethods += methodCount
		summary.MethodsPerType[receiverType] = methodCount

		if group.IsPointer {
			summary.PointerReceivers += methodCount
		} else {
			summary.ValueReceivers += methodCount
		}
	}

	return summary
}

// GroupMethodsByBaseType groups methods by their base type, regardless of whether they are pointer receivers
func (a *Analyzer) GroupMethodsByBaseType(analysis *ReceiverAnalysis) map[string][]*module.Function {
	baseTypeGroups := make(map[string][]*module.Function)

	for _, group := range analysis.Groups {
		baseType := group.BaseType
		if _, exists := baseTypeGroups[baseType]; !exists {
			baseTypeGroups[baseType] = []*module.Function{}
		}

		// Add all methods from this group to the base type group
		baseTypeGroups[baseType] = append(baseTypeGroups[baseType], group.Methods...)
	}

	return baseTypeGroups
}

// FindCommonMethods finds methods with the same name and signature across different receiver types
func (a *Analyzer) FindCommonMethods(analysis *ReceiverAnalysis) map[string][]string {
	// Map of method name to slice of receiver types that implement it
	commonMethods := make(map[string][]string)

	// Map of method name and signature to ensure we only group methods with matching signatures
	methodSignatures := make(map[string]string)

	// First pass: collect method signatures
	for receiverType, group := range analysis.Groups {
		for _, method := range group.Methods {
			methodName := method.Name

			// If this is the first time we're seeing this method name, record its signature
			if existingSignature, exists := methodSignatures[methodName]; !exists {
				methodSignatures[methodName] = method.Signature
			} else if existingSignature != method.Signature {
				// If we have conflicting signatures for the same method name,
				// create a unique key that includes the signature hash
				// This is a simple approach - in a real implementation we might want more sophisticated
				// signature compatibility checking
				methodName = methodName + "_" + strings.ReplaceAll(method.Signature, " ", "")
				methodSignatures[methodName] = method.Signature
			}

			// Initialize the slice if it doesn't exist
			if _, exists := commonMethods[methodName]; !exists {
				commonMethods[methodName] = []string{}
			}

			// Add this receiver type to the list for this method
			found := false
			for _, existingType := range commonMethods[methodName] {
				if existingType == receiverType {
					found = true
					break
				}
			}

			if !found {
				commonMethods[methodName] = append(commonMethods[methodName], receiverType)
			}
		}
	}

	// Filter out method names with unique keys (created due to signature conflicts)
	finalMethods := make(map[string][]string)
	for methodName, receiverTypes := range commonMethods {
		// Only include methods implemented by multiple types
		if len(receiverTypes) > 1 {
			// Remove the signature hash if it was added
			baseName := strings.Split(methodName, "_")[0]
			finalMethods[baseName] = receiverTypes
		}
	}

	return finalMethods
}

// GetReceiverMethodSignatures returns a map of method signatures for each receiver type
func (a *Analyzer) GetReceiverMethodSignatures(analysis *ReceiverAnalysis, receiverType string) map[string]string {
	signatures := make(map[string]string)

	if group, exists := analysis.Groups[receiverType]; exists {
		for _, method := range group.Methods {
			signatures[method.Name] = method.Signature
		}
	}

	return signatures
}

// normalizeReceiverType removes pointer symbols and parentheses from receiver type
func normalizeReceiverType(receiverType string) string {
	// Remove pointer symbol if present
	baseType := strings.TrimPrefix(receiverType, "*")

	// Remove parentheses if present (e.g., "(T)" -> "T")
	baseType = strings.TrimPrefix(baseType, "(")
	baseType = strings.TrimSuffix(baseType, ")")

	return baseType
}
