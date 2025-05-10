package integration

import (
	"bitspark.dev/go-tree/pkg/run/integration/testutil"
	"testing"

	"bitspark.dev/go-tree/pkg/run/execute"
)

// TestSecurityPolicies tests security policies with real functions
func TestSecurityPolicies(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a function runner with real dependencies
	runner := testutil.CreateRunner()

	// Create a restrictive security policy
	policy := execute.NewStandardSecurityPolicy().
		WithAllowNetwork(false).
		WithAllowFileIO(false).
		WithMemoryLimit(10 * 1024 * 1024) // 10MB

	runner.WithSecurity(policy)

	// Get the path to the test module
	modulePath, err := testutil.GetTestModulePath("complexreturn")
	if err != nil {
		t.Fatalf("Failed to get test module path: %v", err)
	}

	// Test an operation that doesn't require network/file access
	// This should succeed despite restrictive policy
	result, err := runner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/complexreturn",
		"GetPerson", // A function that just creates and returns an object
		"Alice")

	if err != nil {
		t.Fatalf("Failed to execute simple function with security policy: %v", err)
	}

	// Verify result
	personMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T: %v", result, result)
	}

	if name, ok := personMap["Name"].(string); !ok || name != "Alice" {
		t.Errorf("Expected Name: Alice, got %v", personMap["Name"])
	}

	// Now try to execute a function that attempts network access
	_, err = runner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/complexreturn",
		"AttemptNetworkAccess", // A function that tries to access the network
		"https://example.com")

	// This should fail due to security policy
	if err == nil {
		t.Error("Expected network access to be blocked, but it succeeded")
	} else {
		t.Logf("Network access correctly blocked: %v", err)
	}

	// Try with a more permissive policy
	permissivePolicy := execute.NewStandardSecurityPolicy().
		WithAllowNetwork(true).
		WithAllowFileIO(true)

	runner.WithSecurity(permissivePolicy)

	// Now the network operation might succeed
	result, err = runner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/complexreturn",
		"AttemptNetworkAccess",
		"https://example.com")

	// Log result for debugging - it might still fail if there's no actual network
	t.Logf("Network access with permissive policy: result=%v, err=%v", result, err)
}
