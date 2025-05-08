package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"bitspark.dev/go-tree/pkg/core/loader"
	"bitspark.dev/go-tree/pkg/execute"
)

type executeOptions struct {
	// Execution options
	ForceColor    bool
	DisableCGO    bool
	Timeout       string
	TestsOnly     bool
	TestBenchmark bool
	TestVerbose   bool
	TestShort     bool
	TestRace      bool
	TestCover     bool
	ExtraEnv      string
}

var executeOpts executeOptions

// newExecuteCmd creates the execute command
func newExecuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute commands on a Go module",
		Long:  `Executes tests and other commands within the context of a Go module.`,
	}

	// Common execution flags
	cmd.PersistentFlags().BoolVar(&executeOpts.ForceColor, "color", false, "Force colorized output")
	cmd.PersistentFlags().BoolVar(&executeOpts.DisableCGO, "disable-cgo", false, "Disable CGO")
	cmd.PersistentFlags().StringVar(&executeOpts.Timeout, "timeout", "", "Timeout for command execution")
	cmd.PersistentFlags().StringVar(&executeOpts.ExtraEnv, "env", "", "Additional environment variables (comma-separated KEY=VALUE pairs)")

	// Add subcommands
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newRunCmd())

	return cmd
}

// newTestCmd creates the test execution command
func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [packages]",
		Short: "Run tests in the module",
		Long:  `Runs Go tests for the specified packages in the module.`,
		RunE:  runTestCmd,
	}

	// Test-specific flags
	cmd.Flags().BoolVar(&executeOpts.TestVerbose, "verbose", false, "Enable verbose test output")
	cmd.Flags().BoolVar(&executeOpts.TestBenchmark, "bench", false, "Run benchmarks")
	cmd.Flags().BoolVar(&executeOpts.TestShort, "short", false, "Run short tests")
	cmd.Flags().BoolVar(&executeOpts.TestRace, "race", false, "Enable race detection")
	cmd.Flags().BoolVar(&executeOpts.TestCover, "cover", false, "Enable test coverage")

	return cmd
}

// newRunCmd creates the run command execution
func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [command]",
		Short: "Run a Go command in the module",
		Long:  `Runs a Go command (build, run, get, etc.) in the context of the module.`,
		Args:  cobra.MinimumNArgs(1),
		RunE:  runGoCmd,
	}

	return cmd
}

// runTestCmd executes tests on the module
func runTestCmd(cmd *cobra.Command, args []string) error {
	// Create a loader to load the module
	modLoader := loader.NewGoModuleLoader()

	// Configure load options
	loadOpts := loader.DefaultLoadOptions()

	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	mod, err := modLoader.LoadWithOptions(GlobalOptions.InputDir, loadOpts)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Create executor
	executor := execute.NewGoExecutor()

	// Configure executor
	executor.EnableCGO = !executeOpts.DisableCGO

	// Set additional environment variables
	if executeOpts.ExtraEnv != "" {
		executor.AdditionalEnv = parseEnvVars(executeOpts.ExtraEnv)
	}

	// Determine packages to test
	pkgPath := "./..."
	if len(args) > 0 {
		pkgPath = args[0]
	}

	// Build test flags
	var testFlags []string

	if executeOpts.TestVerbose {
		testFlags = append(testFlags, "-v")
	}

	if executeOpts.TestBenchmark {
		testFlags = append(testFlags, "-bench=.")
	}

	if executeOpts.TestShort {
		testFlags = append(testFlags, "-short")
	}

	if executeOpts.TestRace {
		testFlags = append(testFlags, "-race")
	}

	if executeOpts.TestCover {
		testFlags = append(testFlags, "-cover")
	}

	if executeOpts.Timeout != "" {
		testFlags = append(testFlags, "-timeout="+executeOpts.Timeout)
	}

	// Run tests
	fmt.Fprintf(os.Stderr, "Running tests for %s\n", pkgPath)
	result, err := executor.ExecuteTest(mod, pkgPath, testFlags...)
	if err != nil {
		return fmt.Errorf("failed to execute tests: %w", err)
	}

	// Report results
	fmt.Printf("Test Results:\n")
	fmt.Printf("  Package: %s\n", result.Package)
	fmt.Printf("  Tests Run: %d\n", len(result.Tests))
	fmt.Printf("  Passed: %d\n", result.Passed)
	fmt.Printf("  Failed: %d\n", result.Failed)

	// Print test output
	if GlobalOptions.Verbose || executeOpts.TestVerbose {
		fmt.Println("\nTest Output:")
		fmt.Println(result.Output)
	} else if result.Failed > 0 {
		// Always show output if tests failed
		fmt.Println("\nTest Output (failures):")
		fmt.Println(result.Output)
	}

	// Return error if any tests failed
	if result.Failed > 0 {
		return fmt.Errorf("tests failed")
	}

	return nil
}

// runGoCmd executes a Go command on the module
func runGoCmd(cmd *cobra.Command, args []string) error {
	// Create a loader to load the module
	modLoader := loader.NewGoModuleLoader()

	// Configure load options
	loadOpts := loader.DefaultLoadOptions()

	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	mod, err := modLoader.LoadWithOptions(GlobalOptions.InputDir, loadOpts)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Create executor
	executor := execute.NewGoExecutor()

	// Configure executor
	executor.EnableCGO = !executeOpts.DisableCGO

	// Set additional environment variables
	if executeOpts.ExtraEnv != "" {
		executor.AdditionalEnv = parseEnvVars(executeOpts.ExtraEnv)
	}

	// Run command
	fmt.Fprintf(os.Stderr, "Running go %s\n", strings.Join(args, " "))
	result, err := executor.Execute(mod, args...)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	// Print output
	if result.StdOut != "" {
		fmt.Print(result.StdOut)
	}

	if result.StdErr != "" {
		fmt.Fprint(os.Stderr, result.StdErr)
	}

	// Return error if command failed
	if result.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", result.ExitCode)
	}

	return nil
}

// parseEnvVars parses comma-separated KEY=VALUE pairs into environment variables
func parseEnvVars(envString string) []string {
	if envString == "" {
		return nil
	}

	parts := strings.Split(envString, ",")
	envVars := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			envVars = append(envVars, part)
		}
	}

	return envVars
}
