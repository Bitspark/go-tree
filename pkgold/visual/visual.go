// Package visual defines interfaces and implementations for visualizing Go modules.
package visual

import (
	"bitspark.dev/go-tree/pkgold/core/module"
)

// ModuleVisualizer creates visual representations of a module
type ModuleVisualizer interface {
	// Visualize creates a visual representation of a module
	Visualize(module *module.Module) ([]byte, error)

	// Name returns the name of the visualizer
	Name() string

	// Description returns a description of what the visualizer produces
	Description() string
}

// BaseVisualizerOptions contains common options for visualizers
type BaseVisualizerOptions struct {
	// Include private (unexported) elements
	IncludePrivate bool

	// Include test files in the visualization
	IncludeTests bool

	// Include generated files in the visualization
	IncludeGenerated bool

	// Custom title for the visualization
	Title string
}
