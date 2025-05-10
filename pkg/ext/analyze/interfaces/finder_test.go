package interfaces

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestGetInterfaceMethods tests the getInterfaceMethods helper function indirectly
func TestGetInterfaceMethods(t *testing.T) {
	// Create a simple test module
	module := typesys.NewModule("test")

	// Create a package
	pkg := typesys.NewPackage(module, "testpkg", "test/testpkg")
	module.Packages["test/testpkg"] = pkg

	// Create a test file
	file := &typesys.File{
		Path:    "testpkg/interfaces.go",
		Package: pkg,
	}
	pkg.Files[file.Path] = file

	// Create an interface with methods
	iface := typesys.NewSymbol("TestInterface", typesys.KindInterface)
	iface.Package = pkg
	iface.File = file
	pkg.Symbols[iface.ID] = iface
	pkg.Exported[iface.Name] = iface

	// Create methods for the interface
	method1 := typesys.NewSymbol("Method1", typesys.KindMethod)
	method1.Package = pkg
	method1.File = file
	method1.Parent = iface
	pkg.Symbols[method1.ID] = method1

	method2 := typesys.NewSymbol("Method2", typesys.KindMethod)
	method2.Package = pkg
	method2.File = file
	method2.Parent = iface
	pkg.Symbols[method2.ID] = method2

	// Add references from interface to methods
	iface.References = append(iface.References,
		&typesys.Reference{Symbol: method1, File: file, Context: iface},
		&typesys.Reference{Symbol: method2, File: file, Context: iface},
	)

	// Create a finder instance
	finder := NewInterfaceFinder(module)

	// Create another interface that embeds the first one
	embedIface := typesys.NewSymbol("EmbedInterface", typesys.KindInterface)
	embedIface.Package = pkg
	embedIface.File = file
	pkg.Symbols[embedIface.ID] = embedIface
	pkg.Exported[embedIface.Name] = embedIface

	// Add a reference to the embedded interface
	embedIface.References = append(embedIface.References,
		&typesys.Reference{Symbol: iface, File: file, Context: embedIface},
	)

	// Add a method to the embedding interface
	method3 := typesys.NewSymbol("Method3", typesys.KindMethod)
	method3.Package = pkg
	method3.File = file
	method3.Parent = embedIface
	pkg.Symbols[method3.ID] = method3

	embedIface.References = append(embedIface.References,
		&typesys.Reference{Symbol: method3, File: file, Context: embedIface},
	)

	// Test if IsImplementedBy correctly identifies methods
	// This indirectly tests getInterfaceMethods
	impl := typesys.NewSymbol("Implementer", typesys.KindStruct)
	impl.Package = pkg
	impl.File = file
	pkg.Symbols[impl.ID] = impl
	pkg.Exported[impl.Name] = impl

	// Add methods to implementer
	implMethod1 := typesys.NewSymbol("Method1", typesys.KindMethod)
	implMethod1.Package = pkg
	implMethod1.File = file
	implMethod1.Parent = impl
	pkg.Symbols[implMethod1.ID] = implMethod1

	implMethod2 := typesys.NewSymbol("Method2", typesys.KindMethod)
	implMethod2.Package = pkg
	implMethod2.File = file
	implMethod2.Parent = impl
	pkg.Symbols[implMethod2.ID] = implMethod2

	// Add references from struct to methods
	impl.References = append(impl.References,
		&typesys.Reference{Symbol: implMethod1, File: file, Context: impl},
		&typesys.Reference{Symbol: implMethod2, File: file, Context: impl},
	)

	// Test IsImplementedBy against the basic interface
	isImpl, err := finder.IsImplementedBy(iface, impl)
	if err != nil {
		t.Fatalf("IsImplementedBy failed: %v", err)
	}
	if !isImpl {
		t.Errorf("Implementer should implement TestInterface")
	}

	// It should fail for the embedding interface since we're missing Method3
	isImpl, err = finder.IsImplementedBy(embedIface, impl)
	if err != nil {
		t.Fatalf("IsImplementedBy failed: %v", err)
	}
	if isImpl {
		t.Errorf("Implementer should not implement EmbedInterface (missing Method3)")
	}
}

// TestGetTypeMethods tests the getTypeMethods helper function indirectly
func TestGetTypeMethods(t *testing.T) {
	// Create a simple test module
	module := typesys.NewModule("test")

	// Create a package
	pkg := typesys.NewPackage(module, "testpkg", "test/testpkg")
	module.Packages["test/testpkg"] = pkg

	// Create a test file
	file := &typesys.File{
		Path:    "testpkg/types.go",
		Package: pkg,
	}
	pkg.Files[file.Path] = file

	// Create a base struct
	baseType := typesys.NewSymbol("Base", typesys.KindStruct)
	baseType.Package = pkg
	baseType.File = file
	pkg.Symbols[baseType.ID] = baseType
	pkg.Exported[baseType.Name] = baseType

	// Create methods for the base struct
	baseMethod := typesys.NewSymbol("BaseMethod", typesys.KindMethod)
	baseMethod.Package = pkg
	baseMethod.File = file
	baseMethod.Parent = baseType
	pkg.Symbols[baseMethod.ID] = baseMethod

	// Add references from base to methods
	baseType.References = append(baseType.References,
		&typesys.Reference{Symbol: baseMethod, File: file, Context: baseType},
	)

	// Create a derived struct that embeds the base
	derivedType := typesys.NewSymbol("Derived", typesys.KindStruct)
	derivedType.Package = pkg
	derivedType.File = file
	pkg.Symbols[derivedType.ID] = derivedType
	pkg.Exported[derivedType.Name] = derivedType

	// Add embedding reference
	derivedType.References = append(derivedType.References,
		&typesys.Reference{Symbol: baseType, File: file, Context: derivedType},
	)

	// Create derived methods
	derivedMethod := typesys.NewSymbol("DerivedMethod", typesys.KindMethod)
	derivedMethod.Package = pkg
	derivedMethod.File = file
	derivedMethod.Parent = derivedType
	pkg.Symbols[derivedMethod.ID] = derivedMethod

	// Add references from derived to methods
	derivedType.References = append(derivedType.References,
		&typesys.Reference{Symbol: derivedMethod, File: file, Context: derivedType},
	)

	// Create a test interface
	iface := typesys.NewSymbol("TestInterface", typesys.KindInterface)
	iface.Package = pkg
	iface.File = file
	pkg.Symbols[iface.ID] = iface
	pkg.Exported[iface.Name] = iface

	// Create methods for the interface
	baseMethodIface := typesys.NewSymbol("BaseMethod", typesys.KindMethod)
	baseMethodIface.Package = pkg
	baseMethodIface.File = file
	baseMethodIface.Parent = iface
	pkg.Symbols[baseMethodIface.ID] = baseMethodIface

	derivedMethodIface := typesys.NewSymbol("DerivedMethod", typesys.KindMethod)
	derivedMethodIface.Package = pkg
	derivedMethodIface.File = file
	derivedMethodIface.Parent = iface
	pkg.Symbols[derivedMethodIface.ID] = derivedMethodIface

	// Add references from interface to methods
	iface.References = append(iface.References,
		&typesys.Reference{Symbol: baseMethodIface, File: file, Context: iface},
		&typesys.Reference{Symbol: derivedMethodIface, File: file, Context: iface},
	)

	// Create a finder instance
	finder := NewInterfaceFinder(module)

	// The derived type should implement the interface due to embedding and its own methods
	isImpl, err := finder.IsImplementedBy(iface, derivedType)
	if err != nil {
		t.Fatalf("IsImplementedBy failed: %v", err)
	}
	if !isImpl {
		t.Errorf("Derived should implement TestInterface")
	}

	// Base type should not implement the interface (missing DerivedMethod)
	isImpl, err = finder.IsImplementedBy(iface, baseType)
	if err != nil {
		t.Fatalf("IsImplementedBy failed: %v", err)
	}
	if isImpl {
		t.Errorf("Base should not implement TestInterface (missing DerivedMethod)")
	}
}

// TestGetAllImplementedInterfaces tests the GetAllImplementedInterfaces method
func TestGetAllImplementedInterfaces(t *testing.T) {
	// Create a simple test module
	module := typesys.NewModule("test")

	// Create a package
	pkg := typesys.NewPackage(module, "testpkg", "test/testpkg")
	module.Packages["test/testpkg"] = pkg

	// Create a test file
	file := &typesys.File{
		Path:    "testpkg/interfaces.go",
		Package: pkg,
	}
	pkg.Files[file.Path] = file

	// Create two interfaces
	iface1 := typesys.NewSymbol("Interface1", typesys.KindInterface)
	iface1.Package = pkg
	iface1.File = file
	pkg.Symbols[iface1.ID] = iface1
	pkg.Exported[iface1.Name] = iface1

	iface2 := typesys.NewSymbol("Interface2", typesys.KindInterface)
	iface2.Package = pkg
	iface2.File = file
	pkg.Symbols[iface2.ID] = iface2
	pkg.Exported[iface2.Name] = iface2

	// Create methods for interface1
	method1 := typesys.NewSymbol("Method1", typesys.KindMethod)
	method1.Package = pkg
	method1.File = file
	method1.Parent = iface1
	pkg.Symbols[method1.ID] = method1

	// Create methods for interface2
	method2 := typesys.NewSymbol("Method2", typesys.KindMethod)
	method2.Package = pkg
	method2.File = file
	method2.Parent = iface2
	pkg.Symbols[method2.ID] = method2

	// Add references from interfaces to methods
	iface1.References = append(iface1.References,
		&typesys.Reference{Symbol: method1, File: file, Context: iface1},
	)

	iface2.References = append(iface2.References,
		&typesys.Reference{Symbol: method2, File: file, Context: iface2},
	)

	// Create a struct that implements both interfaces
	impl := typesys.NewSymbol("Implementer", typesys.KindStruct)
	impl.Package = pkg
	impl.File = file
	pkg.Symbols[impl.ID] = impl
	pkg.Exported[impl.Name] = impl

	// Add methods to implementer
	implMethod1 := typesys.NewSymbol("Method1", typesys.KindMethod)
	implMethod1.Package = pkg
	implMethod1.File = file
	implMethod1.Parent = impl
	pkg.Symbols[implMethod1.ID] = implMethod1

	implMethod2 := typesys.NewSymbol("Method2", typesys.KindMethod)
	implMethod2.Package = pkg
	implMethod2.File = file
	implMethod2.Parent = impl
	pkg.Symbols[implMethod2.ID] = implMethod2

	// Add references from struct to methods
	impl.References = append(impl.References,
		&typesys.Reference{Symbol: implMethod1, File: file, Context: impl},
		&typesys.Reference{Symbol: implMethod2, File: file, Context: impl},
	)

	// Create a finder instance
	finder := NewInterfaceFinder(module)

	// Get all interfaces implemented by the struct
	impls, err := finder.GetAllImplementedInterfaces(impl)
	if err != nil {
		t.Fatalf("GetAllImplementedInterfaces failed: %v", err)
	}

	// It should implement both interfaces
	if len(impls) != 2 {
		t.Errorf("Expected 2 implemented interfaces, got %d", len(impls))
	}

	// Check if both interfaces are found
	foundIface1 := false
	foundIface2 := false
	for _, iface := range impls {
		if iface.ID == iface1.ID {
			foundIface1 = true
		}
		if iface.ID == iface2.ID {
			foundIface2 = true
		}
	}

	if !foundIface1 {
		t.Errorf("Implementer should implement Interface1")
	}
	if !foundIface2 {
		t.Errorf("Implementer should implement Interface2")
	}
}

// TestGetImplementationInfo tests the GetImplementationInfo method
func TestGetImplementationInfo(t *testing.T) {
	// Create a simple test module
	module := typesys.NewModule("test")

	// Create a package
	pkg := typesys.NewPackage(module, "testpkg", "test/testpkg")
	module.Packages["test/testpkg"] = pkg

	// Create a test file
	file := &typesys.File{
		Path:    "testpkg/types.go",
		Package: pkg,
	}
	pkg.Files[file.Path] = file

	// Create an interface with methods
	iface := typesys.NewSymbol("TestInterface", typesys.KindInterface)
	iface.Package = pkg
	iface.File = file
	pkg.Symbols[iface.ID] = iface
	pkg.Exported[iface.Name] = iface

	// Create methods for the interface
	method1 := typesys.NewSymbol("Method1", typesys.KindMethod)
	method1.Package = pkg
	method1.File = file
	method1.Parent = iface
	pkg.Symbols[method1.ID] = method1

	method2 := typesys.NewSymbol("Method2", typesys.KindMethod)
	method2.Package = pkg
	method2.File = file
	method2.Parent = iface
	pkg.Symbols[method2.ID] = method2

	// Add references from interface to methods
	iface.References = append(iface.References,
		&typesys.Reference{Symbol: method1, File: file, Context: iface},
		&typesys.Reference{Symbol: method2, File: file, Context: iface},
	)

	// Create a struct that implements the interface
	impl := typesys.NewSymbol("Implementer", typesys.KindStruct)
	impl.Package = pkg
	impl.File = file
	pkg.Symbols[impl.ID] = impl
	pkg.Exported[impl.Name] = impl

	// Add methods to implementer
	implMethod1 := typesys.NewSymbol("Method1", typesys.KindMethod)
	implMethod1.Package = pkg
	implMethod1.File = file
	implMethod1.Parent = impl
	pkg.Symbols[implMethod1.ID] = implMethod1

	implMethod2 := typesys.NewSymbol("Method2", typesys.KindMethod)
	implMethod2.Package = pkg
	implMethod2.File = file
	implMethod2.Parent = impl
	pkg.Symbols[implMethod2.ID] = implMethod2

	// Add references from struct to methods
	impl.References = append(impl.References,
		&typesys.Reference{Symbol: implMethod1, File: file, Context: impl},
		&typesys.Reference{Symbol: implMethod2, File: file, Context: impl},
	)

	// Create a finder instance
	finder := NewInterfaceFinder(module)

	// Get implementation info
	info, err := finder.GetImplementationInfo(iface, impl)
	if err != nil {
		t.Fatalf("GetImplementationInfo failed: %v", err)
	}

	// Check if both methods are in the method map
	if len(info.MethodMap) != 2 {
		t.Errorf("Expected 2 methods in method map, got %d", len(info.MethodMap))
	}

	// Check Method1
	if methodImpl, ok := info.MethodMap["Method1"]; !ok {
		t.Errorf("Method1 not found in method map")
	} else {
		// Check if method references are correct
		if methodImpl.InterfaceMethod.ID != method1.ID {
			t.Errorf("Incorrect interface method for Method1")
		}
		if methodImpl.ImplementingMethod.ID != implMethod1.ID {
			t.Errorf("Incorrect implementing method for Method1")
		}
		if !methodImpl.IsDirectMatch {
			t.Errorf("Method1 should be a direct match")
		}
	}

	// Check Method2
	if methodImpl, ok := info.MethodMap["Method2"]; !ok {
		t.Errorf("Method2 not found in method map")
	} else {
		// Check if method references are correct
		if methodImpl.InterfaceMethod.ID != method2.ID {
			t.Errorf("Incorrect interface method for Method2")
		}
		if methodImpl.ImplementingMethod.ID != implMethod2.ID {
			t.Errorf("Incorrect implementing method for Method2")
		}
		if !methodImpl.IsDirectMatch {
			t.Errorf("Method2 should be a direct match")
		}
	}
}

// TestFindImplementations tests the FindImplementations method
func TestFindImplementations(t *testing.T) {
	// Create a simple test module
	module := typesys.NewModule("test")

	// Create a package
	pkg := typesys.NewPackage(module, "testpkg", "test/testpkg")
	module.Packages["test/testpkg"] = pkg

	// Create a test file
	file := &typesys.File{
		Path:    "testpkg/interfaces.go",
		Package: pkg,
	}
	pkg.Files[file.Path] = file

	// Create an interface
	iface := typesys.NewSymbol("TestInterface", typesys.KindInterface)
	iface.Package = pkg
	iface.File = file
	pkg.Symbols[iface.ID] = iface
	pkg.Exported[iface.Name] = iface

	// Create a method for the interface
	method := typesys.NewSymbol("Method", typesys.KindMethod)
	method.Package = pkg
	method.File = file
	method.Parent = iface
	pkg.Symbols[method.ID] = method

	// Add reference from interface to method
	iface.References = append(iface.References,
		&typesys.Reference{Symbol: method, File: file, Context: iface},
	)

	// Create two structs, one implements the interface, one doesn't
	impl := typesys.NewSymbol("Implementer", typesys.KindStruct)
	impl.Package = pkg
	impl.File = file
	pkg.Symbols[impl.ID] = impl
	pkg.Exported[impl.Name] = impl

	nonImpl := typesys.NewSymbol("NonImplementer", typesys.KindStruct)
	nonImpl.Package = pkg
	nonImpl.File = file
	pkg.Symbols[nonImpl.ID] = nonImpl
	pkg.Exported[nonImpl.Name] = nonImpl

	// Add method to implementer
	implMethod := typesys.NewSymbol("Method", typesys.KindMethod)
	implMethod.Package = pkg
	implMethod.File = file
	implMethod.Parent = impl
	pkg.Symbols[implMethod.ID] = implMethod

	// Add a different method to non-implementer
	nonImplMethod := typesys.NewSymbol("DifferentMethod", typesys.KindMethod)
	nonImplMethod.Package = pkg
	nonImplMethod.File = file
	nonImplMethod.Parent = nonImpl
	pkg.Symbols[nonImplMethod.ID] = nonImplMethod

	// Add references from structs to methods
	impl.References = append(impl.References,
		&typesys.Reference{Symbol: implMethod, File: file, Context: impl},
	)

	nonImpl.References = append(nonImpl.References,
		&typesys.Reference{Symbol: nonImplMethod, File: file, Context: nonImpl},
	)

	// Create a finder instance
	finder := NewInterfaceFinder(module)

	// Find implementations
	impls, err := finder.FindImplementations(iface)
	if err != nil {
		t.Fatalf("FindImplementations failed: %v", err)
	}

	// Should find one implementation
	if len(impls) != 1 {
		t.Errorf("Expected 1 implementation, got %d", len(impls))
	}

	// Check if the right implementation is found
	if len(impls) > 0 && impls[0].ID != impl.ID {
		t.Errorf("Expected Implementer, got %s", impls[0].Name)
	}
}
