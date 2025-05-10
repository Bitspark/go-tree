package materialize

import (
	"fmt"
	"path/filepath"
	"strings"
)

// NormalizePath standardizes a path for consistent handling
func NormalizePath(path string) string {
	// Clean the path first
	path = filepath.Clean(path)

	// Ensure forward slashes for go.mod
	return filepath.ToSlash(path)
}

// RelativizePath creates a relative path suitable for go.mod
func RelativizePath(basePath, targetPath string) string {
	// Try to create a relative path
	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		// Fall back to absolute path if we can't make it relative
		return NormalizePath(targetPath)
	}

	// If relative path starts with "..", it might be better to use absolute
	if strings.HasPrefix(relPath, "..") && strings.Count(relPath, "..") > 2 {
		// Too many levels up, use absolute path
		return NormalizePath(targetPath)
	}

	// Use relative path
	return NormalizePath(relPath)
}

// IsLocalPath determines if a path is a local filesystem path
func IsLocalPath(path string) bool {
	return filepath.IsAbs(path) || strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/")
}

// CreateUniqueModulePath generates a unique path for a module in a materialization environment
func CreateUniqueModulePath(env *Environment, layoutStrategy LayoutStrategy, modulePath string) string {
	var moduleDir string

	switch layoutStrategy {
	case FlatLayout:
		// Use safe module name
		safeName := strings.ReplaceAll(modulePath, "/", "_")
		moduleDir = filepath.Join(env.RootDir, safeName)

	case HierarchicalLayout:
		// Use hierarchy
		moduleDir = filepath.Join(env.RootDir, NormalizePath(modulePath))

	case GoPathLayout:
		// Use GOPATH style
		moduleDir = filepath.Join(env.RootDir, "src", NormalizePath(modulePath))

	default:
		// Default to flat
		safeName := strings.ReplaceAll(modulePath, "/", "_")
		moduleDir = filepath.Join(env.RootDir, safeName)
	}

	// Ensure unique: if path is already used for any module, add a suffix
	originalPath := moduleDir
	counter := 1

	// Check for path collisions with any module
	pathExists := false
	for _, path := range env.ModulePaths {
		if path == moduleDir {
			pathExists = true
			break
		}
	}

	// If path exists, add numbered suffix until we get a unique one
	for pathExists {
		moduleDir = fmt.Sprintf("%s_%d", originalPath, counter)

		// Check again
		pathExists = false
		for _, path := range env.ModulePaths {
			if path == moduleDir {
				pathExists = true
				break
			}
		}
		counter++
	}

	return moduleDir
}

// EnsureAbsolutePath makes a path absolute if it isn't already
func EnsureAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to convert to absolute path: %w", err)
	}

	return absPath, nil
}

// SanitizePathForFilename creates a filesystem-safe name from a path
func SanitizePathForFilename(path string) string {
	// Replace all path separators with underscores
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, "\\", "_")

	// Replace other problematic characters
	path = strings.ReplaceAll(path, ":", "_")
	path = strings.ReplaceAll(path, "*", "_")
	path = strings.ReplaceAll(path, "?", "_")
	path = strings.ReplaceAll(path, "\"", "_")
	path = strings.ReplaceAll(path, "<", "_")
	path = strings.ReplaceAll(path, ">", "_")
	path = strings.ReplaceAll(path, "|", "_")

	// Collapse multiple underscores
	for strings.Contains(path, "__") {
		path = strings.ReplaceAll(path, "__", "_")
	}

	return path
}
