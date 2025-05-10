package resolve

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// DetectGoVersion returns the Go version used by a module or falls back to runtime
func DetectGoVersion(module *typesys.Module) string {
	// Check if module has a version set
	if module != nil && module.GoVersion != "" {
		return module.GoVersion
	}

	// Use runtime version
	version := runtime.Version()

	// Strip "go" prefix if present
	if strings.HasPrefix(version, "go") {
		version = version[2:]
	}

	return version
}

// GetLatestModuleVersion returns the latest available version of a module
func GetLatestModuleVersion(importPath string) (string, error) {
	// Use go list to get the latest version
	cmd := exec.Command("go", "list", "-m", "-versions", importPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get versions: %w", err)
	}

	// Parse the output
	versions := strings.Fields(string(output))
	if len(versions) <= 1 {
		return "", fmt.Errorf("no versions found for %s", importPath)
	}

	// The first field is the module path, the rest are versions (newest last)
	latestVersion := versions[len(versions)-1]
	return latestVersion, nil
}

// NormalizeVersion ensures a version string is properly formatted
func NormalizeVersion(version string) string {
	// If it's already a proper version, return it
	if version == "" || strings.HasPrefix(version, "v") {
		return version
	}

	// Otherwise, add v prefix
	return "v" + version
}

// CompareVersions compares two version strings and returns:
// -1 if v1 < v2
//
//	0 if v1 == v2
//	1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Normalize versions first
	v1 = NormalizeVersion(v1)
	v2 = NormalizeVersion(v2)

	// If they're the same, return 0
	if v1 == v2 {
		return 0
	}

	// Split version strings into parts (remove v prefix first)
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	// Compare each part
	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		// If parts aren't numeric, compare them as strings
		if v1Parts[i] > v2Parts[i] {
			return 1
		} else if v1Parts[i] < v2Parts[i] {
			return -1
		}
	}

	// If all compared parts are equal, the longer version is greater
	if len(v1Parts) > len(v2Parts) {
		return 1
	} else if len(v1Parts) < len(v2Parts) {
		return -1
	}

	// Should never reach here, but just in case
	return 0
}

// ParseModuleVersionFromGoMod extracts the Go version from a go.mod file
func ParseModuleVersionFromGoMod(goModContent string) string {
	// Look for a line matching "go 1.x"
	lines := strings.Split(goModContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			parts := strings.Fields(line)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}

	// Default to current Go version if not found
	return DetectGoVersion(nil)
}
