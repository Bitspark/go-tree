package resolve

// VersionPolicy determines how version conflicts are handled
type VersionPolicy int

const (
	// StrictVersionPolicy requires exact version matches
	StrictVersionPolicy VersionPolicy = iota

	// LenientVersionPolicy allows compatible semver versions
	LenientVersionPolicy

	// LatestVersionPolicy uses the latest available version
	LatestVersionPolicy
)

// DependencyPolicy determines which dependencies get resolved
type DependencyPolicy int

const (
	// AllDependencies resolves all dependencies recursively
	AllDependencies DependencyPolicy = iota

	// DirectDependenciesOnly resolves only direct dependencies
	DirectDependenciesOnly

	// NoDependencies doesn't resolve any dependencies
	NoDependencies
)

// ResolveOptions configures resolution behavior
type ResolveOptions struct {
	// Whether to include test files
	IncludeTests bool

	// Whether to include private (non-exported) symbols
	IncludePrivate bool

	// Maximum depth for dependency resolution (0 means direct dependencies only)
	DependencyDepth int

	// Whether to download missing dependencies
	DownloadMissing bool

	// Policy for handling version conflicts
	VersionPolicy VersionPolicy

	// Policy for dependency resolution
	DependencyPolicy DependencyPolicy

	// Enable verbose logging
	Verbose bool
}

// DefaultResolveOptions returns a ResolveOptions with default values
func DefaultResolveOptions() ResolveOptions {
	return ResolveOptions{
		IncludeTests:     false,
		IncludePrivate:   true,
		DependencyDepth:  1,
		DownloadMissing:  true,
		VersionPolicy:    LenientVersionPolicy,
		DependencyPolicy: AllDependencies,
		Verbose:          false,
	}
}
