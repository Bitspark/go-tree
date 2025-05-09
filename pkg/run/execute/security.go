package execute

import (
	"fmt"
	"os"

	"bitspark.dev/go-tree/pkg/io/materialize"
)

// StandardSecurityPolicy implements basic security constraints for execution
type StandardSecurityPolicy struct {
	// Whether to allow network access
	AllowNetwork bool

	// Whether to allow file I/O operations
	AllowFileIO bool

	// Memory limit in bytes (0 means no limit)
	MemoryLimit int64

	// Additional environment variables
	EnvVars map[string]string
}

// NewStandardSecurityPolicy creates a new security policy with default settings
func NewStandardSecurityPolicy() *StandardSecurityPolicy {
	return &StandardSecurityPolicy{
		AllowNetwork: false,
		AllowFileIO:  false,
		MemoryLimit:  100 * 1024 * 1024, // 100MB default
		EnvVars:      make(map[string]string),
	}
}

// WithAllowNetwork sets whether network access is allowed
func (p *StandardSecurityPolicy) WithAllowNetwork(allow bool) *StandardSecurityPolicy {
	p.AllowNetwork = allow
	return p
}

// WithAllowFileIO sets whether file I/O is allowed
func (p *StandardSecurityPolicy) WithAllowFileIO(allow bool) *StandardSecurityPolicy {
	p.AllowFileIO = allow
	return p
}

// WithMemoryLimit sets the memory limit in bytes
func (p *StandardSecurityPolicy) WithMemoryLimit(limit int64) *StandardSecurityPolicy {
	p.MemoryLimit = limit
	return p
}

// WithEnvVar adds an environment variable
func (p *StandardSecurityPolicy) WithEnvVar(key, value string) *StandardSecurityPolicy {
	if p.EnvVars == nil {
		p.EnvVars = make(map[string]string)
	}
	p.EnvVars[key] = value
	return p
}

// ApplyToEnvironment applies security constraints to an environment
func (p *StandardSecurityPolicy) ApplyToEnvironment(env *materialize.Environment) error {
	if env == nil {
		return fmt.Errorf("environment cannot be nil")
	}

	// Set environment variables for security constraints
	if !p.AllowNetwork {
		env.SetEnvVar("SANDBOX_NETWORK", "disabled")
	}

	if !p.AllowFileIO {
		env.SetEnvVar("SANDBOX_FILEIO", "disabled")
	}

	if p.MemoryLimit > 0 {
		env.SetEnvVar("GOMEMLIMIT", fmt.Sprintf("%d", p.MemoryLimit))
	}

	// Add any custom environment variables
	for k, v := range p.EnvVars {
		env.SetEnvVar(k, v)
	}

	return nil
}

// ApplyToExecution applies security constraints to command execution
func (p *StandardSecurityPolicy) ApplyToExecution(command []string) []string {
	// This is a simplified implementation
	// In a more comprehensive security model, this could modify the command or args
	// to apply additional security restrictions
	return command
}

// GetEnvironmentVariables returns environment variables for execution
func (p *StandardSecurityPolicy) GetEnvironmentVariables() map[string]string {
	vars := make(map[string]string)

	// Copy security-related environment variables
	if !p.AllowNetwork {
		vars["SANDBOX_NETWORK"] = "disabled"
	}

	if !p.AllowFileIO {
		vars["SANDBOX_FILEIO"] = "disabled"
	}

	if p.MemoryLimit > 0 {
		vars["GOMEMLIMIT"] = fmt.Sprintf("%d", p.MemoryLimit)
	}

	// Copy custom environment variables
	for k, v := range p.EnvVars {
		vars[k] = v
	}

	// Get the current working directory for GOMOD, useful for module-aware execution
	wd, err := os.Getwd()
	if err == nil {
		vars["GOMOD"] = wd
	}

	return vars
}
