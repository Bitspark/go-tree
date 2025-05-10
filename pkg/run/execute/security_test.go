package execute

import (
	"bitspark.dev/go-tree/pkg/env"
	"testing"
)

func TestStandardSecurityPolicy_ApplyToEnvironment(t *testing.T) {
	// Test cases for different security configurations
	testCases := []struct {
		name            string
		configurePolicy func(*StandardSecurityPolicy)
		checkEnv        func(*testing.T, *env.Environment)
	}{
		{
			name: "default policy",
			configurePolicy: func(p *StandardSecurityPolicy) {
				// Use default settings
			},
			checkEnv: func(t *testing.T, env *env.Environment) {
				if val := env.EnvVars["SANDBOX_NETWORK"]; val != "disabled" {
					t.Errorf("Expected SANDBOX_NETWORK=disabled, got %s", val)
				}
				if val := env.EnvVars["SANDBOX_FILEIO"]; val != "disabled" {
					t.Errorf("Expected SANDBOX_FILEIO=disabled, got %s", val)
				}
				if val := env.EnvVars["GOMEMLIMIT"]; val == "" {
					t.Errorf("Expected GOMEMLIMIT to be set")
				}
			},
		},
		{
			name: "allow network",
			configurePolicy: func(p *StandardSecurityPolicy) {
				p.WithAllowNetwork(true)
			},
			checkEnv: func(t *testing.T, env *env.Environment) {
				if val, exists := env.EnvVars["SANDBOX_NETWORK"]; exists {
					t.Errorf("Expected SANDBOX_NETWORK to not be set, got %s", val)
				}
				if val := env.EnvVars["SANDBOX_FILEIO"]; val != "disabled" {
					t.Errorf("Expected SANDBOX_FILEIO=disabled, got %s", val)
				}
			},
		},
		{
			name: "allow file I/O",
			configurePolicy: func(p *StandardSecurityPolicy) {
				p.WithAllowFileIO(true)
			},
			checkEnv: func(t *testing.T, env *env.Environment) {
				if val := env.EnvVars["SANDBOX_NETWORK"]; val != "disabled" {
					t.Errorf("Expected SANDBOX_NETWORK=disabled, got %s", val)
				}
				if val, exists := env.EnvVars["SANDBOX_FILEIO"]; exists {
					t.Errorf("Expected SANDBOX_FILEIO to not be set, got %s", val)
				}
			},
		},
		{
			name: "custom memory limit",
			configurePolicy: func(p *StandardSecurityPolicy) {
				p.WithMemoryLimit(50 * 1024 * 1024) // 50MB
			},
			checkEnv: func(t *testing.T, env *env.Environment) {
				if val := env.EnvVars["GOMEMLIMIT"]; val != "52428800" {
					t.Errorf("Expected GOMEMLIMIT=52428800, got %s", val)
				}
			},
		},
		{
			name: "custom environment variables",
			configurePolicy: func(p *StandardSecurityPolicy) {
				p.WithEnvVar("TEST_VAR", "test_value")
			},
			checkEnv: func(t *testing.T, env *env.Environment) {
				if val := env.EnvVars["TEST_VAR"]; val != "test_value" {
					t.Errorf("Expected TEST_VAR=test_value, got %s", val)
				}
			},
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh environment and policy for each test
			env := env.NewEnvironment("/tmp/test", false)
			policy := NewStandardSecurityPolicy()

			// Configure the policy
			tc.configurePolicy(policy)

			// Apply the policy to the environment
			err := policy.ApplyToEnvironment(env)
			if err != nil {
				t.Fatalf("Failed to apply security policy: %v", err)
			}

			// Check the environment
			tc.checkEnv(t, env)
		})
	}
}

func TestStandardSecurityPolicy_GetEnvironmentVariables(t *testing.T) {
	// Test cases for different security configurations
	testCases := []struct {
		name            string
		configurePolicy func(*StandardSecurityPolicy)
		expectedVars    map[string]string
	}{
		{
			name: "default policy",
			configurePolicy: func(p *StandardSecurityPolicy) {
				// Use default settings
			},
			expectedVars: map[string]string{
				"SANDBOX_NETWORK": "disabled",
				"SANDBOX_FILEIO":  "disabled",
				"GOMEMLIMIT":      "104857600", // 100MB in bytes
			},
		},
		{
			name: "allow network",
			configurePolicy: func(p *StandardSecurityPolicy) {
				p.WithAllowNetwork(true)
			},
			expectedVars: map[string]string{
				"SANDBOX_FILEIO": "disabled",
				"GOMEMLIMIT":     "104857600",
			},
		},
		{
			name: "custom environment variables",
			configurePolicy: func(p *StandardSecurityPolicy) {
				p.WithEnvVar("TEST_VAR1", "value1")
				p.WithEnvVar("TEST_VAR2", "value2")
			},
			expectedVars: map[string]string{
				"SANDBOX_NETWORK": "disabled",
				"SANDBOX_FILEIO":  "disabled",
				"GOMEMLIMIT":      "104857600",
				"TEST_VAR1":       "value1",
				"TEST_VAR2":       "value2",
			},
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh policy for each test
			policy := NewStandardSecurityPolicy()

			// Configure the policy
			tc.configurePolicy(policy)

			// Get environment variables
			vars := policy.GetEnvironmentVariables()

			// Check expected variables
			for k, expectedVal := range tc.expectedVars {
				if actualVal, exists := vars[k]; !exists {
					t.Errorf("Expected variable %s to exist, but it doesn't", k)
				} else if actualVal != expectedVal {
					t.Errorf("Expected %s=%s, got %s", k, expectedVal, actualVal)
				}
			}

			// Check for unexpected variables
			for k := range vars {
				if k != "GOMOD" && k != "GOMEMLIMIT" && tc.expectedVars[k] == "" {
					t.Errorf("Unexpected variable %s=%s", k, vars[k])
				}
			}
		})
	}
}

func TestStandardSecurityPolicy_ApplyToExecution(t *testing.T) {
	// Create a security policy
	policy := NewStandardSecurityPolicy()

	// Test that the command is passed through unchanged
	command := []string{"go", "run", "main.go"}
	modifiedCommand := policy.ApplyToExecution(command)

	// Commands should be the same (current implementation doesn't modify commands)
	if len(modifiedCommand) != len(command) {
		t.Fatalf("Expected command length %d, got %d", len(command), len(modifiedCommand))
	}

	for i, arg := range command {
		if modifiedCommand[i] != arg {
			t.Errorf("Expected command[%d]=%s, got %s", i, arg, modifiedCommand[i])
		}
	}
}
