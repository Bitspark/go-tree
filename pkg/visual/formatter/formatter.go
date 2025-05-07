// Package formatter provides base interfaces and functionality for
// formatting and visualizing Go package data into different output formats.
package formatter

import (
	"bitspark.dev/go-tree/pkg/core/model"
)

// Formatter defines the interface for different visualization formats
type Formatter interface {
	// Format converts a package model to a formatted representation
	Format(pkg *model.GoPackage) (string, error)
}

// Visitor defines the interface for traversing package elements
type Visitor interface {
	VisitPackage(pkg *model.GoPackage) error
	VisitType(typ model.GoType) error
	VisitFunction(fn model.GoFunction) error
	VisitConstant(c model.GoConstant) error
	VisitVariable(v model.GoVariable) error
	VisitImport(imp model.GoImport) error

	// Result returns the final formatted output
	Result() (string, error)
}

// BaseFormatter provides common functionality for formatters
type BaseFormatter struct {
	visitor Visitor
}

// NewBaseFormatter creates a new formatter with the given visitor
func NewBaseFormatter(visitor Visitor) *BaseFormatter {
	return &BaseFormatter{visitor: visitor}
}

// Format applies the visitor to a package and returns the formatted result
func (f *BaseFormatter) Format(pkg *model.GoPackage) (string, error) {
	if err := f.visitor.VisitPackage(pkg); err != nil {
		return "", err
	}

	// Visit all package elements
	for _, imp := range pkg.Imports {
		if err := f.visitor.VisitImport(imp); err != nil {
			return "", err
		}
	}

	for _, typ := range pkg.Types {
		if err := f.visitor.VisitType(typ); err != nil {
			return "", err
		}
	}

	for _, fn := range pkg.Functions {
		if err := f.visitor.VisitFunction(fn); err != nil {
			return "", err
		}
	}

	for _, c := range pkg.Constants {
		if err := f.visitor.VisitConstant(c); err != nil {
			return "", err
		}
	}

	for _, v := range pkg.Variables {
		if err := f.visitor.VisitVariable(v); err != nil {
			return "", err
		}
	}

	return f.visitor.Result()
}
