package env

import (
	"context"
	"fmt"
	"os"
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

	// Toolchain for Go operations (may be nil if not set)
	toolchain GoToolchain

	// Filesystem for operations (may be nil if not set)
	fs ModuleFS
}

// NewEnvironment creates a new environment
func NewEnvironment(rootDir string, isTemporary bool) *Environment {
	return &Environment{
		RootDir:     rootDir,
		ModulePaths: make(map[string]string),
		IsTemporary: isTemporary,
		EnvVars:     make(map[string]string),
		toolchain:   NewStandardGoToolchain(),
		fs:          NewStandardModuleFS(),
	}
}

// WithToolchain sets a custom toolchain
func (e *Environment) WithToolchain(toolchain GoToolchain) *Environment {
	e.toolchain = toolchain
	return e
}

// WithFS sets a custom filesystem
func (e *Environment) WithFS(fs ModuleFS) *Environment {
	e.fs = fs
	return e
}

// Execute runs a command in the context of the specified module
func (e *Environment) Execute(command []string, moduleDir string) ([]byte, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	// Create context for toolchain operations
	ctx := context.Background()

	// Get working directory
	var workDir string
	if moduleDir != "" {
		// Check if it's a module path
		if dir, ok := e.ModulePaths[moduleDir]; ok {
			workDir = dir
		} else {
			// Assume it's a direct path
			workDir = moduleDir
		}
	} else {
		// Default to root directory
		workDir = e.RootDir
	}

	// Check if we have a toolchain
	if e.toolchain == nil {
		e.toolchain = NewStandardGoToolchain()
	}

	// Set up the toolchain
	customToolchain := *e.toolchain.(*StandardGoToolchain)
	customToolchain.WorkDir = workDir

	// Add environment variables
	if len(e.EnvVars) > 0 {
		env := os.Environ()
		for k, v := range e.EnvVars {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		customToolchain.Env = env
	}

	// Execute the command
	return customToolchain.RunCommand(ctx, command[0], command[1:]...)
}

// ExecuteInModule runs a command in the context of the specified module and returns its output
func (e *Environment) ExecuteInModule(command []string, modulePath string) ([]byte, error) {
	return e.Execute(command, modulePath)
}

// ExecuteInRoot runs a command in the root directory
func (e *Environment) ExecuteInRoot(command []string) ([]byte, error) {
	return e.Execute(command, "")
}

// Cleanup removes the environment if it's temporary
func (e *Environment) Cleanup() error {
	if !e.IsTemporary {
		return nil
	}

	// Use filesystem abstraction if available
	if e.fs != nil {
		return e.fs.RemoveAll(e.RootDir)
	}

	// Fallback to standard library
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

	// Use filesystem abstraction if available
	if e.fs != nil {
		_, err := e.fs.Stat(fullPath)
		return err == nil
	}

	// Fallback to standard library
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetPath returns the root directory path
func (e *Environment) GetPath() string {
	return e.RootDir
}

// SetOwned sets whether this environment is temporary (owned)
func (e *Environment) SetOwned(owned bool) {
	e.IsTemporary = owned
}
