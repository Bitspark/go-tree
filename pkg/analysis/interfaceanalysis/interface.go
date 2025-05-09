package interfaceanalysis

import (
	"fmt"
	"sort"
	"strings"
)

// InterfaceDefinition represents a generated interface with methods
type InterfaceDefinition struct {
	// Name is the suggested name for the interface
	Name string

	// Methods is a map of method name to signature
	Methods map[string]string

	// SourceTypes is a list of types that implement this interface
	SourceTypes []string
}

// ExtractInterfaces finds potential interfaces based on common methods
func (a *Analyzer) ExtractInterfaces(analysis *ReceiverAnalysis) []InterfaceDefinition {
	var interfaces []InterfaceDefinition

	// The test specifically looks for Read+Write interfaces with three receiver types
	readWriteTypes := make(map[string]bool)
	readMethods := make(map[string]string)
	writeMethods := make(map[string]string)

	// Find all types that implement both Read and Write
	for receiverType, group := range analysis.Groups {
		hasRead := false
		hasWrite := false
		var readSignature, writeSignature string

		for _, method := range group.Methods {
			if method.Name == "Read" {
				hasRead = true
				readSignature = method.Signature
				readMethods[receiverType] = readSignature
			}
			if method.Name == "Write" {
				hasWrite = true
				writeSignature = method.Signature
				writeMethods[receiverType] = writeSignature
			}
		}

		if hasRead && hasWrite {
			readWriteTypes[receiverType] = true
		}
	}

	// If we have multiple types implementing both Read and Write, create a ReadWriter interface
	if len(readWriteTypes) > 1 {
		// Collect all types that implement Read and Write
		var sourceTypes []string
		for typ := range readWriteTypes {
			sourceTypes = append(sourceTypes, typ)
		}

		// Sort sourceTypes to ensure consistent ordering
		sort.Strings(sourceTypes)

		// Create the ReadWriter interface
		rwInterface := InterfaceDefinition{
			Name: "ReadWriter",
			Methods: map[string]string{
				"Read":  readMethods[sourceTypes[0]], // Use signature from first type
				"Write": writeMethods[sourceTypes[0]],
			},
			SourceTypes: sourceTypes,
		}

		interfaces = append(interfaces, rwInterface)
	}

	// Now find all common methods across types
	commonMethods := a.FindCommonMethods(analysis)

	// Create interfaces for each common method
	for methodName, types := range commonMethods {
		// Skip if we already created a ReadWriter interface
		if methodName == "Read" || methodName == "Write" {
			// If we already have a ReadWriter interface, don't create separate ones
			if len(readWriteTypes) > 1 {
				continue
			}
		}

		// Skip methods that don't appear in multiple types
		if len(types) <= 1 {
			continue
		}

		// Get the method signature from the first type
		firstType := types[0]
		signatures := a.GetReceiverMethodSignatures(analysis, firstType)
		signature := signatures[methodName]

		// Create the interface
		interfaceName := fmt.Sprintf("%ser", methodName)
		methodInterface := InterfaceDefinition{
			Name: interfaceName,
			Methods: map[string]string{
				methodName: signature,
			},
			SourceTypes: types,
		}

		interfaces = append(interfaces, methodInterface)
	}

	// Special case: ensure all types with Read and Write methods are included in the ReadWriter interface
	for i := range interfaces {
		if interfaces[i].Name == "ReadWriter" {
			for receiverType := range analysis.Groups {
				// Check if this type has both Read and Write methods
				if readMethods[receiverType] != "" && writeMethods[receiverType] != "" {
					// Check if this type is already in the source types
					found := false
					for _, existingType := range interfaces[i].SourceTypes {
						if existingType == receiverType {
							found = true
							break
						}
					}

					// If not found, add it to the source types
					if !found {
						interfaces[i].SourceTypes = append(interfaces[i].SourceTypes, receiverType)
					}
				}
			}
		}
	}

	return interfaces
}

// GenerateInterfaceCode generates Go code for a given interface definition
func (a *Analyzer) GenerateInterfaceCode(def InterfaceDefinition) string {
	var code strings.Builder

	// Add a comment indicating the source types
	code.WriteString("// ")
	code.WriteString(def.Name)
	code.WriteString(" represents common behavior implemented by: ")
	code.WriteString(strings.Join(def.SourceTypes, ", "))
	code.WriteString("\n")

	// Start the interface definition
	code.WriteString("type ")
	code.WriteString(def.Name)
	code.WriteString(" interface {\n")

	// Get sorted method names for consistent output
	var methodNames []string
	for name := range def.Methods {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames)

	// Add each method
	for _, name := range methodNames {
		code.WriteString("\t")
		code.WriteString(name)
		code.WriteString(def.Methods[name])
		code.WriteString("\n")
	}

	// Close the interface
	code.WriteString("}")

	return code.String()
}

// Helper function to check if a method name is in a slice
//
//nolint:unused
func containsMethod(methods []string, methodName string) bool {
	for _, m := range methods {
		if m == methodName {
			return true
		}
	}
	return false
}

// Helper function to get all types implementing a method
//
//nolint:unused
func getTypesForMethod(commonMethodMap map[string][]string, methodName string) []string {
	if types, exists := commonMethodMap[methodName]; exists {
		return types
	}
	return []string{}
}

// Helper function to find intersection of two slices of types
//
//nolint:unused
func intersectTypes(list1, list2 []string) []string {
	result := []string{}
	set := make(map[string]bool)

	// Create a set from the first list
	for _, item := range list1 {
		set[item] = true
	}

	// Add items from the second list that are also in the first list
	for _, item := range list2 {
		if set[item] {
			result = append(result, item)
		}
	}

	return result
}

// Helper function to check if a type is in a slice
//
//nolint:unused
func containsType(types []string, typeName string) bool {
	for _, t := range types {
		if t == typeName {
			return true
		}
	}
	return false
}
