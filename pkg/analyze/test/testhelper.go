package test

import (
	"go/token"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// CreateTestModule creates a module with mock data for testing
func CreateTestModule(t *testing.T) *typesys.Module {
	t.Helper()

	// Create a new module
	module := typesys.NewModule("bitspark.dev/go-tree/testdata")

	// Create a package
	pkg := typesys.NewPackage(module, "samplepackage", "bitspark.dev/go-tree/testdata/samplepackage")
	module.Packages["bitspark.dev/go-tree/testdata/samplepackage"] = pkg

	// Add test file
	file := createTestFile(pkg)
	pkg.Files[file.Path] = file

	// Create symbols for testing
	createTestSymbols(pkg, file)

	return module
}

// createTestFile creates a mock file for testing
func createTestFile(pkg *typesys.Package) *typesys.File {
	file := &typesys.File{
		Path:    "samplepackage/types.go",
		Package: pkg,
	}
	return file
}

// createTestSymbols creates mock symbols for testing
func createTestSymbols(pkg *typesys.Package, file *typesys.File) {
	// Create interface symbols
	authenticator := createInterface("Authenticator", pkg, file)
	validator := createInterface("Validator", pkg, file)

	// Create method symbols for interfaces
	login := createMethod("Login", authenticator, pkg, file)
	logout := createMethod("Logout", authenticator, pkg, file)
	validate := createMethod("Validate", validator, pkg, file)

	// Add methods to their interfaces (to create a relationship)
	authenticator.References = append(authenticator.References,
		&typesys.Reference{Symbol: login, File: file, Context: authenticator},
		&typesys.Reference{Symbol: logout, File: file, Context: authenticator},
	)

	validator.References = append(validator.References,
		&typesys.Reference{Symbol: validate, File: file, Context: validator},
	)

	// Reference Validator from Authenticator (embedded interface)
	authenticator.References = append(authenticator.References, &typesys.Reference{
		Symbol:  validator,
		File:    file,
		Context: authenticator,
	})

	// Create User struct
	user := createStruct("User", pkg, file)

	// Create User methods that implement both interfaces
	userLogin := createMethod("Login", user, pkg, file)
	userLogout := createMethod("Logout", user, pkg, file)
	userValidate := createMethod("Validate", user, pkg, file)

	// Add method implementations to User
	user.References = append(user.References,
		&typesys.Reference{Symbol: userLogin, File: file, Context: user},
		&typesys.Reference{Symbol: userLogout, File: file, Context: user},
		&typesys.Reference{Symbol: userValidate, File: file, Context: user},
	)

	// Create Functions
	newUser := createFunction("NewUser", pkg, file)
	formatUser := createFunction("FormatUser", pkg, file)

	// Add references between symbols to create a call graph
	newUser.References = append(newUser.References, &typesys.Reference{
		Symbol: user,
		File:   file,
	})

	formatUser.References = append(formatUser.References,
		&typesys.Reference{Symbol: user, File: file},
		&typesys.Reference{Symbol: newUser, File: file, IsWrite: false}, // Call to NewUser
	)

	userLogin.References = append(userLogin.References, &typesys.Reference{
		Symbol:  userValidate,
		File:    file,
		IsWrite: false, // This is a call
	})
}

// Helper functions to create different kinds of symbols

func createInterface(name string, pkg *typesys.Package, file *typesys.File) *typesys.Symbol {
	sym := typesys.NewSymbol(name, typesys.KindInterface)
	sym.Package = pkg
	sym.File = file
	sym.Exported = name[0] >= 'A' && name[0] <= 'Z'

	pkg.Symbols[sym.ID] = sym
	if sym.Exported {
		pkg.Exported[name] = sym
	}

	return sym
}

func createStruct(name string, pkg *typesys.Package, file *typesys.File) *typesys.Symbol {
	sym := typesys.NewSymbol(name, typesys.KindStruct)
	sym.Package = pkg
	sym.File = file
	sym.Exported = name[0] >= 'A' && name[0] <= 'Z'

	pkg.Symbols[sym.ID] = sym
	if sym.Exported {
		pkg.Exported[name] = sym
	}

	return sym
}

func createMethod(name string, parent *typesys.Symbol, pkg *typesys.Package, file *typesys.File) *typesys.Symbol {
	sym := typesys.NewSymbol(name, typesys.KindMethod)
	sym.Package = pkg
	sym.File = file
	sym.Parent = parent
	sym.Exported = name[0] >= 'A' && name[0] <= 'Z'

	pkg.Symbols[sym.ID] = sym

	// Add method definition position
	sym.AddDefinition(file.Path, token.Pos(0), 0, 0)

	return sym
}

func createFunction(name string, pkg *typesys.Package, file *typesys.File) *typesys.Symbol {
	sym := typesys.NewSymbol(name, typesys.KindFunction)
	sym.Package = pkg
	sym.File = file
	sym.Exported = name[0] >= 'A' && name[0] <= 'Z'

	pkg.Symbols[sym.ID] = sym
	if sym.Exported {
		pkg.Exported[name] = sym
	}

	// Add function definition position
	sym.AddDefinition(file.Path, token.Pos(0), 0, 0)

	return sym
}

// FindSymbolByName finds a symbol by name in the module
func FindSymbolByName(module *typesys.Module, name string) *typesys.Symbol {
	for _, pkg := range module.Packages {
		// Check exported symbols first (faster lookup)
		if sym, ok := pkg.Exported[name]; ok {
			return sym
		}

		// Check all symbols
		for _, sym := range pkg.Symbols {
			if sym.Name == name {
				return sym
			}
		}
	}
	return nil
}
