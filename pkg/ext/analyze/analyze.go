// Package analyze provides type-aware code analysis capabilities for Go programs.
// It builds on the type system core to provide accurate and comprehensive
// static analysis of Go code, including complex features like interfaces,
// type embedding, and generics.
package analyze

// Analyzer is the base interface for all analyzers in the analyze package.
type Analyzer interface {
	// Name returns the name of the analyzer.
	Name() string

	// Description returns a brief description of what the analyzer does.
	Description() string
}

// AnalysisResult represents the result of an analysis operation.
type AnalysisResult interface {
	// GetAnalyzer returns the analyzer that produced this result.
	GetAnalyzer() Analyzer

	// IsSuccess returns true if the analysis completed successfully.
	IsSuccess() bool

	// GetError returns any error that occurred during analysis.
	GetError() error
}

// BaseAnalyzer provides common functionality for all analyzers.
type BaseAnalyzer struct {
	name        string
	description string
}

// Name returns the name of the analyzer.
func (a *BaseAnalyzer) Name() string {
	return a.name
}

// Description returns a brief description of what the analyzer does.
func (a *BaseAnalyzer) Description() string {
	return a.description
}

// NewBaseAnalyzer creates a new base analyzer with the given name and description.
func NewBaseAnalyzer(name, description string) *BaseAnalyzer {
	return &BaseAnalyzer{
		name:        name,
		description: description,
	}
}

// BaseResult provides a basic implementation of AnalysisResult.
type BaseResult struct {
	analyzer Analyzer
	err      error
}

// GetAnalyzer returns the analyzer that produced this result.
func (r *BaseResult) GetAnalyzer() Analyzer {
	return r.analyzer
}

// IsSuccess returns true if the analysis completed successfully.
func (r *BaseResult) IsSuccess() bool {
	return r.err == nil
}

// GetError returns any error that occurred during analysis.
func (r *BaseResult) GetError() error {
	return r.err
}

// NewBaseResult creates a new base result for the given analyzer and error.
func NewBaseResult(analyzer Analyzer, err error) *BaseResult {
	return &BaseResult{
		analyzer: analyzer,
		err:      err,
	}
}
