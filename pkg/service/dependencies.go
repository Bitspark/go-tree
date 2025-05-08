package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// parseGoMod parses a go.mod file and extracts dependencies
func parseGoMod(content string) (map[string]string, map[string]string, error) {
	deps := make(map[string]string)
	replacements := make(map[string]string)

	// Check if we have a require block
	hasRequireBlock := regexp.MustCompile(`require\s*\(`).MatchString(content)

	if hasRequireBlock {
		// Extract dependencies from require blocks
		reqBlockRe := regexp.MustCompile(`require\s*\(\s*([\s\S]*?)\s*\)`)
		blockMatches := reqBlockRe.FindAllStringSubmatch(content, -1)

		for _, blockMatch := range blockMatches {
			if len(blockMatch) >= 2 {
				blockContent := blockMatch[1]
				// Find all module/version pairs within the block
				moduleRe := regexp.MustCompile(`\s*([^\s]+)\s+v?([^(\s]+)`)
				moduleMatches := moduleRe.FindAllStringSubmatch(blockContent, -1)

				for _, modMatch := range moduleMatches {
					if len(modMatch) >= 3 {
						importPath := modMatch[1]
						version := modMatch[2]
						// Ensure version has v prefix if needed
						if !strings.HasPrefix(version, "v") && (strings.HasPrefix(version, "0.") || strings.HasPrefix(version, "1.") || strings.HasPrefix(version, "2.")) {
							version = "v" + version
						}
						deps[importPath] = version
					}
				}
			}
		}
	} else {
		// No require blocks, check for standalone require statements
		reqSingleRe := regexp.MustCompile(`require\s+([^\s]+)\s+v?([^(\s]+)`)
		singleMatches := reqSingleRe.FindAllStringSubmatch(content, -1)

		for _, match := range singleMatches {
			if len(match) >= 3 {
				importPath := match[1]
				version := match[2]
				// Ensure version has v prefix if needed
				if !strings.HasPrefix(version, "v") && (strings.HasPrefix(version, "0.") || strings.HasPrefix(version, "1.") || strings.HasPrefix(version, "2.")) {
					version = "v" + version
				}
				deps[importPath] = version
			}
		}
	}

	// Check if we have a replace block
	hasReplaceBlock := regexp.MustCompile(`replace\s*\(`).MatchString(content)

	if hasReplaceBlock {
		// Extract replacements from replace blocks
		replBlockRe := regexp.MustCompile(`replace\s*\(\s*([\s\S]*?)\s*\)`)
		blockReplMatches := replBlockRe.FindAllStringSubmatch(content, -1)

		for _, blockMatch := range blockReplMatches {
			if len(blockMatch) >= 2 {
				blockContent := blockMatch[1]
				// Find all replacement pairs within the block
				replRe := regexp.MustCompile(`\s*([^\s]+)(?:\s+v?[^=>\s]+)?\s+=>\s+(?:([^\s]+)\s+v?([^(\s]+)|([^\s]+))`)
				replMatches := replRe.FindAllStringSubmatch(blockContent, -1)

				for _, replMatch := range replMatches {
					if len(replMatch) >= 5 {
						originalPath := replMatch[1]
						if replMatch[4] != "" {
							// Local replacement (=> ./some/path)
							replacements[originalPath] = replMatch[4]
						} else if replMatch[2] != "" {
							// Remote replacement (=> github.com/... v1.2.3)
							replacements[originalPath] = replMatch[2]
						}
					}
				}
			}
		}
	} else {
		// No replace blocks, check for standalone replace statements
		replSingleRe := regexp.MustCompile(`replace\s+([^\s]+)(?:\s+v?[^=>\s]+)?\s+=>\s+(?:([^\s]+)\s+v?([^(\s]+)|([^\s]+))`)
		singleReplMatches := replSingleRe.FindAllStringSubmatch(content, -1)

		for _, match := range singleReplMatches {
			if len(match) >= 5 {
				originalPath := match[1]
				if match[4] != "" {
					// Local replacement (=> ./some/path)
					replacements[originalPath] = match[4]
				} else if match[2] != "" {
					// Remote replacement (=> github.com/... v1.2.3)
					replacements[originalPath] = match[2]
				}
			}
		}
	}

	return deps, replacements, nil
}

// findDependencyDir locates a dependency in the GOPATH or module cache
// This is a standalone utility function used by DependencyManager
func findDependencyDir(importPath, version string) (string, error) {
	// Check for local replacements in go.mod
	// This would be done in a more comprehensive implementation

	// Check GOPATH/pkg/mod
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// Fall back to default GOPATH if not set
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("GOPATH not set and could not determine home directory: %w", err)
		}
		gopath = filepath.Join(home, "go")
	}

	// Check GOMODCACHE if available (introduced in Go 1.15)
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		// Default location is $GOPATH/pkg/mod
		gomodcache = filepath.Join(gopath, "pkg", "mod")
	}

	// Format the expected path in the module cache
	// Module paths use @ as a separator between the module path and version
	modPath := filepath.Join(gomodcache, importPath+"@"+version)
	if _, err := os.Stat(modPath); err == nil {
		return modPath, nil
	}

	// Check if it's using a different version format (v prefix vs non-prefix)
	if len(version) > 0 && version[0] == 'v' {
		// Try without v prefix
		altVersion := version[1:]
		altModPath := filepath.Join(gomodcache, importPath+"@"+altVersion)
		if _, err := os.Stat(altModPath); err == nil {
			return altModPath, nil
		}
	} else {
		// Try with v prefix
		altVersion := "v" + version
		altModPath := filepath.Join(gomodcache, importPath+"@"+altVersion)
		if _, err := os.Stat(altModPath); err == nil {
			return altModPath, nil
		}
	}

	// Check in old-style GOPATH mode (pre-modules)
	oldStylePath := filepath.Join(gopath, "src", importPath)
	if _, err := os.Stat(oldStylePath); err == nil {
		return oldStylePath, nil
	}

	return "", fmt.Errorf("could not find dependency %s@%s in module cache or GOPATH", importPath, version)
}
