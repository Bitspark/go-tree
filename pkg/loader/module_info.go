package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/typesys"
)

// extractModuleInfo extracts module path and Go version from go.mod file
func extractModuleInfo(module *typesys.Module) error {
	// Check if go.mod exists
	goModPath := filepath.Join(module.Dir, "go.mod")
	goModPath = normalizePath(goModPath)

	// Validate that goModPath is within the module directory to prevent path traversal
	moduleDir := normalizePath(module.Dir)
	if !strings.HasPrefix(goModPath, moduleDir) {
		return fmt.Errorf("invalid go.mod path detected")
	}

	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found in %s", module.Dir)
	}

	// Read go.mod
	content, err := os.ReadFile(goModPath) // #nosec G304 - Path is validated above to be within module directory
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Parse module path and Go version more robustly
	lines := strings.Split(string(content), "\n")
	inMultilineBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle multiline blocks
		if strings.Contains(line, "(") {
			inMultilineBlock = true
			continue
		}

		if strings.Contains(line, ")") {
			inMultilineBlock = false
			continue
		}

		// Skip lines in multiline blocks
		if inMultilineBlock {
			continue
		}

		// Handle module declaration with proper word boundary checking
		if strings.HasPrefix(line, "module ") {
			// Extract the module path, handling quotes if present
			modulePath := strings.TrimPrefix(line, "module ")
			modulePath = strings.TrimSpace(modulePath)

			// Handle quoted module paths
			if strings.HasPrefix(modulePath, "\"") && strings.HasSuffix(modulePath, "\"") {
				modulePath = modulePath[1 : len(modulePath)-1]
			} else if strings.HasPrefix(modulePath, "'") && strings.HasSuffix(modulePath, "'") {
				modulePath = modulePath[1 : len(modulePath)-1]
			}

			module.Path = modulePath
		} else if strings.HasPrefix(line, "go ") {
			// Extract go version
			goVersion := strings.TrimPrefix(line, "go ")
			goVersion = strings.TrimSpace(goVersion)

			// Handle quoted go versions
			if strings.HasPrefix(goVersion, "\"") && strings.HasSuffix(goVersion, "\"") {
				goVersion = goVersion[1 : len(goVersion)-1]
			} else if strings.HasPrefix(goVersion, "'") && strings.HasSuffix(goVersion, "'") {
				goVersion = goVersion[1 : len(goVersion)-1]
			}

			module.GoVersion = goVersion
		}
	}

	// Validate that we found a module path
	if module.Path == "" {
		return fmt.Errorf("no module declaration found in go.mod")
	}

	return nil
}
