package runner

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
	"bitspark.dev/go-tree/pkg/run/common"
	"bitspark.dev/go-tree/pkg/run/execute"
	"bitspark.dev/go-tree/pkg/run/testing"
)

// init registers the runner factory with the testing package
func init() {
	// Register our runner factory
	testing.RegisterRunnerFactory(createRunner)

	// Register our unified test executor to avoid import cycles
	unifiedRunner := NewUnifiedTestRunner(execute.NewGoExecutor(), nil, nil)
	testing.RegisterTestExecutor(func(env *materialize.Environment, module *typesys.Module,
		pkgPath string, testFlags ...string) (*common.TestResult, error) {
		return unifiedRunner.ExecuteTest(env, module, pkgPath, testFlags...)
	})
}

// createRunner creates a runner that implements the testing.TestRunner interface
func createRunner() testing.TestRunner {
	// Create the real runner
	runner := NewRunner(execute.NewGoExecutor())

	// Wrap it in an adapter to match the testing.TestRunner interface
	return &runnerAdapter{runner: runner}
}

// runnerAdapter adapts Runner to the testing.TestRunner interface
type runnerAdapter struct {
	runner *Runner
}

// RunTests implements testing.TestRunner.RunTests
func (a *runnerAdapter) RunTests(mod *typesys.Module, pkgPath string, opts *common.RunOptions) (*common.TestResult, error) {
	return a.runner.RunTests(mod, pkgPath, opts)
}

// AnalyzeCoverage implements testing.TestRunner.AnalyzeCoverage
func (a *runnerAdapter) AnalyzeCoverage(mod *typesys.Module, pkgPath string) (*common.CoverageResult, error) {
	return a.runner.AnalyzeCoverage(mod, pkgPath)
}
