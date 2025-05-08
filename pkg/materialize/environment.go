package materialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Environment represents materialized modules and provides operations on them
type Environment struct {
	// Root directory where modules are materialized
	RootDir string

	// Mapping from module paths to filesystem paths
	ModulePaths map[string]string

	// Whether this is a temporary environment (will be cleaned up automatically)
	IsTemporary bool

	// Environment variables for command execution
	EnvVars map[string]string
}

// NewEnvironment creates a new environment
func NewEnvironment(rootDir string, isTemporary bool) *Environment {
	return &Environment{
		RootDir:     rootDir,
		ModulePaths: make(map[string]string),
		IsTemporary: isTemporary,
		EnvVars:     make(map[string]string),
	}
}

// Execute runs a command in the context of the specified module
func (e *Environment) Execute(command []string, moduleDir string) (*exec.Cmd, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	// Create command
	cmd := exec.Command(command[0], command[1:]...)

	// Set working directory if specified
	if moduleDir != "" {
		// Check if it's a module path
		if dir, ok := e.ModulePaths[moduleDir]; ok {
			cmd.Dir = dir
		} else {
			// Assume it's a direct path
			cmd.Dir = moduleDir
		}
	} else {
		// Default to root directory
		cmd.Dir = e.RootDir
	}

	// Set environment variables
	if len(e.EnvVars) > 0 {
		cmd.Env = os.Environ()
		for k, v := range e.EnvVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return cmd, nil
}

// ExecuteInModule runs a command in the context of the specified module and returns its output
func (e *Environment) ExecuteInModule(command []string, modulePath string) ([]byte, error) {
	cmd, err := e.Execute(command, modulePath)
	if err != nil {
		return nil, err
	}

	return cmd.CombinedOutput()
}

// ExecuteInRoot runs a command in the root directory
func (e *Environment) ExecuteInRoot(command []string) ([]byte, error) {
	cmd, err := e.Execute(command, "")
	if err != nil {
		return nil, err
	}

	return cmd.CombinedOutput()
}

// Cleanup removes the environment if it's temporary
func (e *Environment) Cleanup() error {
	if !e.IsTemporary {
		return nil
	}

	// Remove the root directory and all contents
	return os.RemoveAll(e.RootDir)
}

// GetModulePath returns the filesystem path for a given module
func (e *Environment) GetModulePath(modulePath string) (string, bool) {
	path, ok := e.ModulePaths[modulePath]
	return path, ok
}

// AllModulePaths returns all module paths in the environment
func (e *Environment) AllModulePaths() []string {
	paths := make([]string, 0, len(e.ModulePaths))
	for path := range e.ModulePaths {
		paths = append(paths, path)
	}
	return paths
}

// SetEnvVar sets an environment variable for command execution
func (e *Environment) SetEnvVar(key, value string) {
	if e.EnvVars == nil {
		e.EnvVars = make(map[string]string)
	}
	e.EnvVars[key] = value
}

// GetEnvVar gets an environment variable
func (e *Environment) GetEnvVar(key string) (string, bool) {
	if e.EnvVars == nil {
		return "", false
	}
	val, ok := e.EnvVars[key]
	return val, ok
}

// ClearEnvVars clears all environment variables
func (e *Environment) ClearEnvVars() {
	e.EnvVars = make(map[string]string)
}

// FileExists checks if a file exists in the environment
func (e *Environment) FileExists(modulePath, relPath string) bool {
	moduleDir, ok := e.ModulePaths[modulePath]
	if !ok {
		return false
	}

	fullPath := filepath.Join(moduleDir, relPath)
	_, err := os.Stat(fullPath)
	return err == nil
}
