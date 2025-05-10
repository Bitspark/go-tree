package materialize

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute/materializeinterface"
)

// Ensure that Environment implements the materializeinterface.Environment interface
var _ materializeinterface.Environment = (*Environment)(nil)

// Ensure that ModuleMaterializer implements the materializeinterface.ModuleMaterializer interface
var _ materializeinterface.ModuleMaterializer = (*ModuleMaterializer)(nil)

// Materialize implements the materializeinterface.ModuleMaterializer interface
func (m *ModuleMaterializer) Materialize(module interface{}, options interface{}) (materializeinterface.Environment, error) {
	// Convert the generic module to a typesys.Module
	typedModule, ok := module.(*typesys.Module)
	if !ok {
		return nil, &MaterializeError{
			Message: "module must be a *typesys.Module",
		}
	}

	// Convert the generic options to MaterializeOptions
	var opts MaterializeOptions
	if options != nil {
		typedOpts, ok := options.(MaterializeOptions)
		if !ok {
			return nil, &MaterializeError{
				Message: "options must be a MaterializeOptions struct",
			}
		}
		opts = typedOpts
	} else {
		opts = DefaultMaterializeOptions()
	}

	// Call the real implementation
	return m.materializeModule(typedModule, opts)
}
