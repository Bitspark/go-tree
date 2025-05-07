// Package interfaceanalysis provides functionality for analyzing method receivers
// and extracting interface information from Go code.
package interfaceanalysis

import (
	"bitspark.dev/go-tree/pkg/core/model"
)

// ReceiverGroup organizes methods by their receiver type
type ReceiverGroup struct {
	// ReceiverType is the name of the receiver type (e.g., "*User" or "User")
	ReceiverType string

	// BaseType is the name of the receiver type without pointers (e.g., "User")
	BaseType string

	// IsPointer indicates if the receiver is a pointer type
	IsPointer bool

	// Methods is a list of methods that have this receiver type
	Methods []model.GoFunction
}

// ReceiverAnalysis contains the full method receiver analysis for a package
type ReceiverAnalysis struct {
	// Package is the name of the analyzed package
	Package string

	// Groups maps receiver types to their group of methods
	Groups map[string]*ReceiverGroup
}

// ReceiverSummary provides summary information about receivers in the package
type ReceiverSummary struct {
	// TotalMethods is the total number of methods in the package
	TotalMethods int

	// TotalReceiverTypes is the number of unique receiver types
	TotalReceiverTypes int

	// MethodsPerType is a map of receiver type to method count
	MethodsPerType map[string]int

	// PointerReceivers is the count of methods with pointer receivers
	PointerReceivers int

	// ValueReceivers is the count of methods with value receivers
	ValueReceivers int
}
