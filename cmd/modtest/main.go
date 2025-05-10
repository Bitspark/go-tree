package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/common"
	"bitspark.dev/go-tree/pkg/run/testing/runner"
	"bitspark.dev/go-tree/pkg/service"
)

// TestResult extends common.TestResult with timing information
type TestResult struct {
	*common.TestResult
	Name      string
	Package   string
	Duration  time.Duration
	StartTime time.Time
}

func main() {
	// Setup command-line flags
	verbose := flag.Bool("v", false, "Verbose output")
	failFast := flag.Bool("failfast", false, "Stop on first test failure")
	coverage := flag.Bool("coverage", false, "Calculate test coverage")
	specificPkg := flag.String("package", "", "Test only a specific package (default is all packages)")
	_ = flag.Duration("timeout", 10*time.Minute, "Test timeout duration") // Parsed but handled by the Go test command internally
	flag.Parse()

	// Get module path from argument, default to current directory
	modulePath := "."
	if flag.NArg() > 0 {
		modulePath = flag.Arg(0)
	}

	// Get test function prefix filter (if provided)
	testFuncPrefix := ""
	if flag.NArg() > 1 {
		testFuncPrefix = flag.Arg(1)
	}

	// Convert to absolute path for better error messages
	absPath, err := filepath.Abs(modulePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error converting to absolute path: %v\n", err)
		os.Exit(1)
	}

	// Check if the directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "❌ Directory does not exist: %s\n", absPath)
		os.Exit(1)
	}

	// Check if go.mod exists
	goModPath := filepath.Join(absPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "❌ No go.mod file found in: %s\n", absPath)
		os.Exit(1)
	}

	// Create service configuration
	config := &service.Config{
		ModuleDir:       absPath,
		IncludeTests:    true, // Important for test discovery
		WithDeps:        true, // May be needed for tests that use dependencies
		DependencyDepth: 1,
		DownloadMissing: false,
		Verbose:         *verbose,
	}

	// Create service to load the module
	fmt.Println("🔍 Loading module...")
	svc, err := service.NewService(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading module: %v\n", err)
		os.Exit(1)
	}

	// Get the main module
	mainModule := svc.GetMainModule()
	if mainModule == nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load main module\n")
		os.Exit(1)
	}

	// Print module information
	fmt.Printf("📦 Module: %s\n", mainModule.Path)
	fmt.Printf("📂 Directory: %s\n", mainModule.Dir)

	// Set package path to test
	pkgPath := "./..." // Default to all packages
	if *specificPkg != "" {
		pkgPath = *specificPkg
	}

	// Initialize the test runner with default settings
	testRunner := runner.DefaultRunner()

	// Find all test functions in the module
	testFunctions := findTestFunctions(mainModule, testFuncPrefix)

	if len(testFunctions) == 0 {
		fmt.Printf("\n😞 No test functions found")
		if testFuncPrefix != "" {
			fmt.Printf(" matching prefix '%s'", testFuncPrefix)
		}
		fmt.Println()
		os.Exit(0)
	}

	fmt.Printf("\n🧪 Found %d test functions", len(testFunctions))
	if testFuncPrefix != "" {
		fmt.Printf(" matching prefix '%s'", testFuncPrefix)
	}
	fmt.Println()

	// Organize test functions by package
	testsByPackage := organizeTestsByPackage(testFunctions)

	// Track detailed test results for statistics
	testResults := make([]*TestResult, 0, len(testFunctions))

	// Track overall results
	overallResult := &common.TestResult{
		Tests:  make([]string, 0),
		Passed: 0,
		Failed: 0,
		Output: "",
	}

	// Record start time for performance measurement
	startTime := time.Now()

	// Run tests package by package, function by function
	failedTests := false
	pkgIndex := 0
	pkgCount := len(testsByPackage)

	for pkg, tests := range testsByPackage {
		pkgIndex++
		fmt.Printf("\n📦 Running tests in package (%d/%d): %s\n", pkgIndex, pkgCount, pkg)

		for i, testFunc := range tests {
			testName := testFunc.Name
			shortName := testName
			if len(shortName) > 25 {
				shortName = shortName[:22] + "..."
			}

			fmt.Printf("  [%3d/%-3d] 🧪 %-25s ", i+1, len(tests), shortName)

			// Track start time for this test
			testStartTime := time.Now()

			// Start a goroutine to show progress dots while the test is running
			progressChan := make(chan bool)
			go showProgress(progressChan)

			// Create test options with this specific test only
			testOptions := &common.RunOptions{
				Verbose: *verbose,
				Tests:   []string{testName},
			}

			// Run this specific test
			result, err := testRunner.RunTests(mainModule, pkg, testOptions)

			// Stop the progress indicator
			progressChan <- true

			// Calculate test duration
			testDuration := time.Since(testStartTime)

			// Store detailed test results
			detailedResult := &TestResult{
				TestResult: result,
				Name:       testName,
				Package:    pkg,
				Duration:   testDuration,
				StartTime:  testStartTime,
			}
			testResults = append(testResults, detailedResult)

			// Update overall results
			if result != nil {
				overallResult.Tests = append(overallResult.Tests, testName)

				if err != nil || result.Failed > 0 {
					overallResult.Failed++
					failedTests = true
					fmt.Printf("\r  [%3d/%-3d] ❌ %-25s %s\n", i+1, len(tests), shortName, formatDuration(testDuration))

					// Print failure details if we're not in verbose mode
					// (verbose mode will show this in the output)
					if !*verbose && result.Output != "" {
						lines := strings.Split(result.Output, "\n")
						for _, line := range lines {
							if strings.Contains(line, testName) && (strings.Contains(line, "FAIL") || strings.Contains(line, "Error")) {
								fmt.Printf("      💥 %s\n", line)
							}
						}
					}

					// Exit early if failFast is set
					if *failFast {
						fmt.Println("\n🛑 Stopping due to test failure (-failfast flag)")
						break
					}
				} else {
					overallResult.Passed++
					fmt.Printf("\r  [%3d/%-3d] ✅ %-25s %s\n", i+1, len(tests), shortName, formatDuration(testDuration))
				}

				// Append test output to overall output
				if *verbose {
					overallResult.Output += result.Output + "\n"
				}
			}
		}

		// If failFast and we had failures, don't continue to next package
		if *failFast && failedTests {
			break
		}
	}

	// Calculate execution time
	totalDuration := time.Since(startTime)

	// Print overall test results
	printOverallResults(overallResult, totalDuration)

	// Print test statistics
	printTestStatistics(testResults, totalDuration)

	// Run coverage analysis if requested
	if *coverage {
		fmt.Println("\n📊 Running coverage analysis...")
		coverageResult, err := testRunner.AnalyzeCoverage(mainModule, pkgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error analyzing coverage: %v\n", err)
		} else if coverageResult != nil {
			printCoverageResults(coverageResult, totalDuration)
		}
	}

	// Exit with appropriate code based on test results
	if failedTests {
		fmt.Println("\n❌ Some tests failed")
		os.Exit(1)
	} else {
		fmt.Println("\n🎉 All tests passed successfully!")
		os.Exit(0)
	}
}

// showProgress displays a progress indicator (dots) while a test is running
func showProgress(done chan bool) {
	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			fmt.Print(".")
		}
	}
}

// findTestFunctions finds all test functions in the module that match the given prefix
func findTestFunctions(module *typesys.Module, prefix string) []*typesys.Symbol {
	var testFunctions []*typesys.Symbol

	// Search through all packages in the module
	for _, pkg := range module.Packages {
		// Skip non-test packages if we're looking for tests
		if !strings.HasSuffix(pkg.Name, "_test") && !containsTestFiles(pkg) {
			continue
		}

		// Look for function symbols that start with "Test"
		for _, symbol := range pkg.Symbols {
			if symbol.Kind == typesys.KindFunction {
				// Must start with "Test" and have an uppercase letter after that
				if strings.HasPrefix(symbol.Name, "Test") && len(symbol.Name) > 4 {
					// Apply additional prefix filter if provided
					if prefix == "" || strings.HasPrefix(symbol.Name, prefix) {
						testFunctions = append(testFunctions, symbol)
					}
				}
			}
		}
	}

	return testFunctions
}

// containsTestFiles checks if a package contains test files
func containsTestFiles(pkg *typesys.Package) bool {
	for _, file := range pkg.Files {
		if strings.HasSuffix(file.Name, "_test.go") {
			return true
		}
	}
	return false
}

// organizeTestsByPackage groups test functions by their package import path
func organizeTestsByPackage(tests []*typesys.Symbol) map[string][]*typesys.Symbol {
	result := make(map[string][]*typesys.Symbol)

	for _, test := range tests {
		pkgPath := test.Package.ImportPath
		result[pkgPath] = append(result[pkgPath], test)
	}

	return result
}

// printTestStatistics displays statistics about the test run
func printTestStatistics(results []*TestResult, totalDuration time.Duration) {
	fmt.Println("\n📊 Test Statistics 📊")
	fmt.Println(strings.Repeat("─", 80))

	// Get total count of tests
	totalTests := len(results)
	if totalTests == 0 {
		fmt.Println("No tests were executed")
		return
	}

	// Calculate statistics
	var totalTestDuration time.Duration
	for _, result := range results {
		totalTestDuration += result.Duration
	}

	// Sort tests by duration (slowest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Duration > results[j].Duration
	})

	// Average duration
	avgDuration := totalTestDuration / time.Duration(totalTests)

	// Calculate median duration
	medianDuration := results[totalTests/2].Duration

	// Print general statistics
	fmt.Printf("Total Tests:        %d\n", totalTests)
	fmt.Printf("Average Duration:   %s\n", formatDuration(avgDuration))
	fmt.Printf("Median Duration:    %s\n", formatDuration(medianDuration))
	fmt.Printf("Execution Overhead: %s\n", formatDuration(totalDuration-totalTestDuration))

	// Print top 5 slowest tests or fewer if there are less than 5 tests
	numSlowTests := 5
	if totalTests < numSlowTests {
		numSlowTests = totalTests
	}

	fmt.Printf("\n⏱️  Top %d Slowest Tests:\n", numSlowTests)
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("%-40s %-30s %s\n", "Test", "Package", "Duration")
	fmt.Println(strings.Repeat("─", 80))

	for i := 0; i < numSlowTests; i++ {
		testName := results[i].Name
		if len(testName) > 38 {
			testName = testName[:35] + "..."
		}

		pkgName := results[i].Package
		if len(pkgName) > 28 {
			pkgName = "..." + pkgName[len(pkgName)-25:]
		}

		fmt.Printf("%-40s %-30s %s\n", testName, pkgName, formatDuration(results[i].Duration))
	}

	// Distribution of test durations
	fmt.Println("\n⏱️  Duration Distribution:")
	fmt.Println(strings.Repeat("─", 80))

	// Define duration buckets
	buckets := []struct {
		name  string
		upper time.Duration
	}{
		{"< 10ms", 10 * time.Millisecond},
		{"10-50ms", 50 * time.Millisecond},
		{"50-100ms", 100 * time.Millisecond},
		{"100-500ms", 500 * time.Millisecond},
		{"500ms-1s", 1 * time.Second},
		{"> 1s", time.Hour}, // Effectively unlimited upper bound
	}

	// Count tests in each bucket
	counts := make([]int, len(buckets))
	for _, result := range results {
		for i, bucket := range buckets {
			if result.Duration < bucket.upper {
				counts[i]++
				break
			}
		}
	}

	// Calculate max count for scaling
	maxCount := 0
	for _, count := range counts {
		if count > maxCount {
			maxCount = count
		}
	}

	// Display histogram
	maxBarWidth := 40
	for i, bucket := range buckets {
		barWidth := 0
		if maxCount > 0 {
			barWidth = counts[i] * maxBarWidth / maxCount
		}

		// Create the bar
		bar := strings.Repeat("█", barWidth)

		// Display the histogram line
		fmt.Printf("%-10s | %-40s %d\n", bucket.name, bar, counts[i])
	}
}

// printOverallResults displays the overall test results
func printOverallResults(result *common.TestResult, duration time.Duration) {
	fmt.Println()
	fmt.Println(strings.Repeat("═", 80))
	fmt.Printf("✨ Overall Test Results ✨\n")
	fmt.Println(strings.Repeat("─", 80))

	// Determine emoji based on pass/fail
	statusEmoji := "🎉"
	if result.Failed > 0 {
		statusEmoji = "❌"
	}

	fmt.Printf("%s Total Tests: %d\n", statusEmoji, result.Passed+result.Failed)
	fmt.Printf("✅ Passed: %d\n", result.Passed)

	if result.Failed > 0 {
		fmt.Printf("❌ Failed: %d\n", result.Failed)
	} else {
		fmt.Printf("❌ Failed: %d\n", result.Failed)
	}

	fmt.Printf("⏱️  Total Time: %s\n", formatDuration(duration))
	fmt.Println(strings.Repeat("═", 80))

	// Print verbose output if available and requested
	if result.Output != "" {
		fmt.Println("\n📝 Detailed Test Output:")
		fmt.Println(result.Output)
	}
}

// printCoverageResults displays coverage results in a human-readable format
func printCoverageResults(result *common.CoverageResult, duration time.Duration) {
	fmt.Println()
	fmt.Println(strings.Repeat("═", 80))
	fmt.Printf("📊 Coverage Results 📊\n")
	fmt.Println(strings.Repeat("─", 80))

	// Determine emoji based on coverage percentage
	coverageEmoji := "🔴"
	if result.Percentage >= 80 {
		coverageEmoji = "🟢" // Green for good coverage
	} else if result.Percentage >= 50 {
		coverageEmoji = "🟡" // Yellow for medium coverage
	} else if result.Percentage >= 30 {
		coverageEmoji = "🟠" // Orange for low coverage
	}

	fmt.Printf("%s Overall Coverage: %.2f%%\n", coverageEmoji, result.Percentage)
	fmt.Printf("⏱️  Analysis Time: %s\n", formatDuration(duration))
	fmt.Println(strings.Repeat("═", 80))

	// Print per-file coverage if available
	if len(result.Files) > 0 {
		fmt.Println("\n📁 Coverage by File:")
		fmt.Println(strings.Repeat("─", 80))

		// Convert map to slice for sorting
		type fileCoverage struct {
			file     string
			coverage float64
		}

		filesSlice := make([]fileCoverage, 0, len(result.Files))
		for file, cov := range result.Files {
			filesSlice = append(filesSlice, fileCoverage{file, cov})
		}

		// Sort by coverage (lowest first)
		sort.Slice(filesSlice, func(i, j int) bool {
			return filesSlice[i].coverage < filesSlice[j].coverage
		})

		// Print top 10 files with lowest coverage
		maxFiles := 10
		if len(filesSlice) < maxFiles {
			maxFiles = len(filesSlice)
		}

		fmt.Printf("🔍 Top %d files with lowest coverage:\n", maxFiles)
		for i := 0; i < maxFiles; i++ {
			file := filesSlice[i].file
			cov := filesSlice[i].coverage

			// Determine emoji based on file coverage
			emoji := "🔴"
			if cov >= 80 {
				emoji = "🟢"
			} else if cov >= 50 {
				emoji = "🟡"
			} else if cov >= 30 {
				emoji = "🟠"
			}

			// Truncate long filenames
			if len(file) > 60 {
				file = "..." + file[len(file)-57:]
			}

			fmt.Printf("  %s %-60s %.2f%%\n", emoji, file, cov)
		}
	}

	// Print uncovered functions if available
	if len(result.UncoveredFunctions) > 0 {
		fmt.Println("\n⚠️  Uncovered Functions:")
		fmt.Println(strings.Repeat("─", 80))

		// Only show top 20 uncovered functions to avoid overwhelming output
		maxUncovered := 20
		if len(result.UncoveredFunctions) < maxUncovered {
			maxUncovered = len(result.UncoveredFunctions)
		}

		for i := 0; i < maxUncovered; i++ {
			sym := result.UncoveredFunctions[i]
			fmt.Printf("  🔍 %s.%s\n", sym.Package.ImportPath, sym.Name)
		}

		// If there are more, show a count
		if len(result.UncoveredFunctions) > maxUncovered {
			fmt.Printf("  ... and %d more uncovered functions\n",
				len(result.UncoveredFunctions)-maxUncovered)
		}
	}
}

// formatDuration returns a human-friendly string for a duration
func formatDuration(d time.Duration) string {
	// Round to milliseconds for readability
	d = d.Round(time.Millisecond)

	if d < time.Millisecond {
		return fmt.Sprintf("%d µs", d.Microseconds())
	}

	if d < time.Second {
		return fmt.Sprintf("%d ms", d.Milliseconds())
	}

	seconds := d.Seconds()
	if seconds < 60 {
		return fmt.Sprintf("%.1f s", seconds)
	}

	minutes := int(seconds) / 60
	remainingSeconds := seconds - float64(minutes*60)
	return fmt.Sprintf("%d min %.1f s", minutes, remainingSeconds)
}
