// Package service provides a unified interface to Go-Tree functionality
package service

import (
	"bitspark.dev/go-tree/pkg/index"
	"bitspark.dev/go-tree/pkg/loader"
	"bitspark.dev/go-tree/pkg/typesys"
)

// Config holds service configuration
type Config struct {
	ModuleDir    string
	IncludeTests bool
	WithDeps     bool
	Verbose      bool
}

// Service provides a unified interface to Go-Tree functionality
type Service struct {
	Module *typesys.Module
	Index  *index.Index
	Config *Config
}

// NewService creates a new service instance
func NewService(config *Config) (*Service, error) {
	// Load module using the loader package
	module, err := loader.LoadModule(config.ModuleDir, &typesys.LoadOptions{
		IncludeTests: config.IncludeTests,
	})
	if err != nil {
		return nil, err
	}

	// Create index - adjusted to match actual signature
	idx := index.NewIndex(module)

	return &Service{
		Module: module,
		Index:  idx,
		Config: config,
	}, nil
}
