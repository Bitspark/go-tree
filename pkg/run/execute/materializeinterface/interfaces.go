// Package materializeinterface provides interfaces for materializing modules
// This package exists to break import cycles between materialize and execute packages
package materializeinterface

// Environment represents a code execution environment
type Environment interface {
	GetPath() string
	Cleanup() error
	SetOwned(owned bool)
}

// ModuleMaterializer defines the interface for materializing modules
type ModuleMaterializer interface {
	// Materialize materializes a module with the given options
	// The actual module and options types are opaque, so we use interface{}
	Materialize(module interface{}, options interface{}) (Environment, error)
}
