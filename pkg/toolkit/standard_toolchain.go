package toolkit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StandardGoToolchain uses the actual go binary
type StandardGoToolchain struct {
	// Path to the go binary, defaults to "go" (resolves using PATH)
	GoExecutable string

	// Environment variables for toolchain execution
	Env []string

	// Working directory for commands
	WorkDir string
}

// NewStandardGoToolchain creates a new standard toolchain
func NewStandardGoToolchain() *StandardGoToolchain {
	return &StandardGoToolchain{
		GoExecutable: "go",
		Env:          os.Environ(),
		WorkDir:      "",
	}
}

// RunCommand executes a Go command with arguments
func (t *StandardGoToolchain) RunCommand(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmdArgs := append([]string{command}, args...)
	cmd := exec.CommandContext(ctx, t.GoExecutable, cmdArgs...)

	if t.Env != nil {
		cmd.Env = t.Env
	}

	if t.WorkDir != "" {
		cmd.Dir = t.WorkDir
	}

	return cmd.CombinedOutput()
}

// GetModuleInfo retrieves information about a module
func (t *StandardGoToolchain) GetModuleInfo(ctx context.Context, importPath string) (path string, version string, err error) {
	output, err := t.RunCommand(ctx, "list", "-m", importPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to get module information for %s: %w", importPath, err)
	}

	// Parse output (format: "path version")
	parts := strings.Fields(string(output))
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected output format from go list -m: %s", output)
	}

	return parts[0], parts[1], nil
}

// DownloadModule downloads a module
func (t *StandardGoToolchain) DownloadModule(ctx context.Context, importPath string, version string) error {
	versionSpec := importPath
	if version != "" {
		versionSpec += "@" + version
	}

	_, err := t.RunCommand(ctx, "get", "-d", versionSpec)
	if err != nil {
		return fmt.Errorf("failed to download module %s@%s: %w", importPath, version, err)
	}

	return nil
}

// FindModule locates a module in the module cache
func (t *StandardGoToolchain) FindModule(ctx context.Context, importPath string, version string) (string, error) {
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

	return "", fmt.Errorf("module %s@%s not found in module cache or GOPATH", importPath, version)
}

// CheckModuleExists verifies a module exists and is accessible
func (t *StandardGoToolchain) CheckModuleExists(ctx context.Context, importPath string, version string) (bool, error) {
	path, err := t.FindModule(ctx, importPath, version)
	if err != nil {
		return false, nil // Module not found, but not an error
	}

	return path != "", nil
}
