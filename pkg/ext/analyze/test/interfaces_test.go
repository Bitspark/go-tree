package test

import (
	"bitspark.dev/go-tree/pkg/ext/analyze/interfaces"
	"testing"
)

// TestInterfaceFinder tests the interface implementation finder
func TestInterfaceFinder(t *testing.T) {
	// Create test module
	module := CreateTestModule(t)

	// Create an interface finder
	finder := interfaces.NewInterfaceFinder(module)

	// Get the Authenticator interface
	authenticatorIface := FindSymbolByName(module, "Authenticator")
	if authenticatorIface == nil {
		t.Fatal("Could not find Authenticator interface")
	}

	// Get the Validator interface
	validatorIface := FindSymbolByName(module, "Validator")
	if validatorIface == nil {
		t.Fatal("Could not find Validator interface")
	}

	// Get the User type
	userType := FindSymbolByName(module, "User")
	if userType == nil {
		t.Fatal("Could not find User type")
	}

	// Test IsImplementedBy
	t.Run("TestIsImplementedBy", func(t *testing.T) {
		// Check if User implements Authenticator
		isAuthImpl, err := finder.IsImplementedBy(authenticatorIface, userType)
		if err != nil {
			t.Fatalf("IsImplementedBy failed: %v", err)
		}
		if !isAuthImpl {
			t.Error("User should implement Authenticator")
		}

		// Check if User implements Validator
		isValidatorImpl, err := finder.IsImplementedBy(validatorIface, userType)
		if err != nil {
			t.Fatalf("IsImplementedBy failed: %v", err)
		}
		if !isValidatorImpl {
			t.Error("User should implement Validator")
		}
	})

	// Test FindImplementations
	t.Run("TestFindImplementations", func(t *testing.T) {
		// Find all Authenticator implementations
		authImpls, err := finder.FindImplementations(authenticatorIface)
		if err != nil {
			t.Fatalf("FindImplementations failed: %v", err)
		}

		// Check if User is in the list
		foundUser := false
		for _, impl := range authImpls {
			if impl.ID == userType.ID {
				foundUser = true
				break
			}
		}
		if !foundUser {
			t.Error("User should be in the list of Authenticator implementations")
		}
	})

	// Test GetImplementationInfo
	t.Run("TestGetImplementationInfo", func(t *testing.T) {
		// Get implementation details
		implInfo, err := finder.GetImplementationInfo(authenticatorIface, userType)
		if err != nil {
			t.Fatalf("GetImplementationInfo failed: %v", err)
		}

		// Check method mapping
		// There should be Login, Logout, and Validate methods
		if len(implInfo.MethodMap) != 3 {
			t.Errorf("Expected 3 methods, got %d", len(implInfo.MethodMap))
		}

		// Check if Login method is in the map
		_, hasLogin := implInfo.MethodMap["Login"]
		if !hasLogin {
			t.Error("Login method not found in implementation info")
		}

		// Check if Logout method is in the map
		_, hasLogout := implInfo.MethodMap["Logout"]
		if !hasLogout {
			t.Error("Logout method not found in implementation info")
		}

		// Check if Validate method is in the map (from embedded Validator)
		_, hasValidate := implInfo.MethodMap["Validate"]
		if !hasValidate {
			t.Error("Validate method not found in implementation info")
		}
	})

	// Test GetAllImplementedInterfaces
	t.Run("TestGetAllImplementedInterfaces", func(t *testing.T) {
		// Get all interfaces implemented by User
		impls, err := finder.GetAllImplementedInterfaces(userType)
		if err != nil {
			t.Fatalf("GetAllImplementedInterfaces failed: %v", err)
		}

		// User should implement both Authenticator and Validator
		if len(impls) < 2 {
			t.Errorf("Expected User to implement at least 2 interfaces, got %d", len(impls))
		}

		// Check if Authenticator is in the list
		foundAuth := false
		foundValidator := false
		for _, impl := range impls {
			if impl.ID == authenticatorIface.ID {
				foundAuth = true
			}
			if impl.ID == validatorIface.ID {
				foundValidator = true
			}
		}

		if !foundAuth {
			t.Error("User should implement Authenticator")
		}
		if !foundValidator {
			t.Error("User should implement Validator")
		}
	})
}
