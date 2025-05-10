package resolve

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// DependencyAnalyzer analyzes module dependencies
type DependencyAnalyzer struct {
	registry ModuleRegistry
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(registry ModuleRegistry) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		registry: registry,
	}
}

// AnalyzeModule analyzes a module's dependencies and updates its Dependencies field
func (a *DependencyAnalyzer) AnalyzeModule(module *typesys.Module) error {
	if module == nil {
		return nil
	}

	// Skip if module doesn't have a directory
	if module.Dir == "" {
		return nil
	}

	// Parse go.mod file for dependencies
	goModPath := filepath.Join(module.Dir, "go.mod")
	deps, replacements, err := parseGoModFile(goModPath)
	if err != nil {
		return err
	}

	// Clear existing dependencies and replacements
	module.Dependencies = make([]*typesys.Dependency, 0, len(deps))
	module.Replacements = replacements

	// Add dependencies
	for importPath, version := range deps {
		// Check if this dependency is in the registry
		isLocal := false
		fsPath := ""

		if a.registry != nil {
			if resolvedModule, ok := a.registry.FindModule(importPath); ok {
				isLocal = resolvedModule.IsLocal
				fsPath = resolvedModule.FilesystemPath
			}
		}

		// Create dependency
		dep := &typesys.Dependency{
			ImportPath:     importPath,
			Version:        version,
			IsLocal:        isLocal,
			FilesystemPath: fsPath,
		}

		// Add to module
		module.Dependencies = append(module.Dependencies, dep)
	}

	return nil
}

// Regular expressions for parsing go.mod
var (
	requireRegex    = regexp.MustCompile(`(?m)^require\s+(\S+)\s+(\S+)$`)
	requireBlkRegex = regexp.MustCompile(`(?s)require\s*\((.*?)\)`)
	replaceRegex    = regexp.MustCompile(`(?m)^replace\s+(\S+)\s+=>\s+(\S+)(?:\s+(\S+))?$`)
	replaceBlkRegex = regexp.MustCompile(`(?s)replace\s*\((.*?)\)`)
	depEntryRegex   = regexp.MustCompile(`^\s*(\S+)\s+(\S+)$`)
)

// parseGoModFile parses a go.mod file and returns dependencies and replacements
func parseGoModFile(path string) (map[string]string, map[string]string, error) {
	// Open the file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	// Parse dependencies and replacements
	deps := make(map[string]string)
	replacements := make(map[string]string)

	contentStr := string(content)

	// Parse standalone requires
	for _, match := range requireRegex.FindAllStringSubmatch(contentStr, -1) {
		if len(match) >= 3 {
			deps[match[1]] = match[2]
		}
	}

	// Parse require blocks
	for _, block := range requireBlkRegex.FindAllStringSubmatch(contentStr, -1) {
		if len(block) >= 2 {
			scanner := bufio.NewScanner(strings.NewReader(block[1]))
			for scanner.Scan() {
				line := scanner.Text()
				parts := depEntryRegex.FindStringSubmatch(line)
				if len(parts) >= 3 {
					deps[parts[1]] = parts[2]
				}
			}
		}
	}

	// Parse standalone replaces
	for _, match := range replaceRegex.FindAllStringSubmatch(contentStr, -1) {
		if len(match) >= 3 {
			replacements[match[1]] = match[2]
		}
	}

	// Parse replace blocks
	for _, block := range replaceBlkRegex.FindAllStringSubmatch(contentStr, -1) {
		if len(block) >= 2 {
			scanner := bufio.NewScanner(strings.NewReader(block[1]))
			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.SplitN(line, "=>", 2)
				if len(parts) == 2 {
					from := strings.TrimSpace(parts[0])
					to := strings.TrimSpace(parts[1])
					if from != "" && to != "" {
						// Extract version if present
						toParts := strings.Fields(to)
						if len(toParts) > 0 {
							replacements[from] = toParts[0]
						}
					}
				}
			}
		}
	}

	return deps, replacements, nil
}
