package integration

import (
	"bitspark.dev/go-tree/pkg/io/materialize"
	"bitspark.dev/go-tree/pkg/run/execute/materializeinterface"
	"bitspark.dev/go-tree/pkg/testutil/materializehelper"
)

func init() {
	// Initialize the materializehelper with a function to create materializers
	materializehelper.Initialize(func() materializeinterface.ModuleMaterializer {
		return materialize.NewModuleMaterializer()
	})
}
