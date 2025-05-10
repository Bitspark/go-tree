// Package materializehelper provides utilities for testing materialization
package materializehelper

import (
	"bitspark.dev/go-tree/pkg/run/execute/materializeinterface"
)

// GetMaterializer is a function type that provides a materializer
type GetMaterializer func() materializeinterface.ModuleMaterializer

// Global callback to get a materializer
var materializer GetMaterializer

// Initialize sets the function used to get materializers
func Initialize(getMaterializer GetMaterializer) {
	materializer = getMaterializer
}

// GetDefaultMaterializer returns a materializer for testing
func GetDefaultMaterializer() materializeinterface.ModuleMaterializer {
	if materializer == nil {
		panic("materializehelper not initialized - call Initialize with a provider function")
	}
	return materializer()
}
