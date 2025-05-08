package saver

import (
	"sync"

	"bitspark.dev/go-tree/pkg/typesys"
)

// DefaultModificationTracker is a simple implementation of ModificationTracker
// that tracks modified elements using a map.
type DefaultModificationTracker struct {
	// Use a sync.Map for thread safety
	modifiedElements sync.Map
}

// NewDefaultModificationTracker creates a new default modification tracker
func NewDefaultModificationTracker() *DefaultModificationTracker {
	return &DefaultModificationTracker{}
}

// IsModified checks if an element has been modified
func (t *DefaultModificationTracker) IsModified(element interface{}) bool {
	if element == nil {
		return false
	}

	// Check if the element is in the map
	_, found := t.modifiedElements.Load(element)
	return found
}

// MarkModified marks an element as modified
func (t *DefaultModificationTracker) MarkModified(element interface{}) {
	if element == nil {
		return
	}

	// Add the element to the map
	t.modifiedElements.Store(element, true)

	// Also mark parent elements as modified if applicable
	switch v := element.(type) {
	case *typesys.Symbol:
		// Mark file as modified
		if v.File != nil {
			t.MarkModified(v.File)
		}

		// Mark package as modified
		if v.Package != nil {
			t.MarkModified(v.Package)
		}

	case *typesys.File:
		// Mark package as modified
		if v.Package != nil {
			t.MarkModified(v.Package)
		}

	case *typesys.Package:
		// Mark module as modified
		if v.Module != nil {
			t.MarkModified(v.Module)
		}
	}
}

// ClearModified clears the modified status of an element
func (t *DefaultModificationTracker) ClearModified(element interface{}) {
	if element == nil {
		return
	}

	// Remove the element from the map
	t.modifiedElements.Delete(element)
}

// ClearAll clears all modification tracking
func (t *DefaultModificationTracker) ClearAll() {
	// Create a new map
	t.modifiedElements = sync.Map{}
}

// ModificationsAnalyzer provides utilities for analyzing modifications
type ModificationsAnalyzer struct {
	tracker ModificationTracker
}

// NewModificationsAnalyzer creates a new modifications analyzer
func NewModificationsAnalyzer(tracker ModificationTracker) *ModificationsAnalyzer {
	return &ModificationsAnalyzer{
		tracker: tracker,
	}
}

// GetModifiedFiles returns all modified files in a module
func (a *ModificationsAnalyzer) GetModifiedFiles(module *typesys.Module) []*typesys.File {
	modified := make([]*typesys.File, 0)

	// Check each package in the module
	for _, pkg := range module.Packages {
		// Check each file in the package
		for _, file := range pkg.Files {
			// Check if the file is modified
			if a.tracker.IsModified(file) {
				modified = append(modified, file)
			} else {
				// Check if any symbol in the file is modified
				for _, sym := range file.Symbols {
					if a.tracker.IsModified(sym) {
						modified = append(modified, file)
						break
					}
				}
			}
		}
	}

	return modified
}

// GetModifiedSymbols returns all modified symbols in a file
func (a *ModificationsAnalyzer) GetModifiedSymbols(file *typesys.File) []*typesys.Symbol {
	modified := make([]*typesys.Symbol, 0)

	// Check each symbol in the file
	for _, sym := range file.Symbols {
		if a.tracker.IsModified(sym) {
			modified = append(modified, sym)
		}
	}

	return modified
}
