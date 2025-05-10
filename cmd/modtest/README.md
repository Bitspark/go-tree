# modtest

A command-line utility to run all tests in a Go module using the Go-Tree testing framework.

## Usage

```
go run cmd/modtest/main.go [flags] [module_path] [test_prefix]
```

Where:
- `module_path` is the path to a directory containing a Go module (with a go.mod file)
- If `module_path` is not provided, the current directory (`.`) is used
- `test_prefix` is an optional prefix to filter test functions (only runs tests starting with this prefix)

### Flags

- `-v`: Verbose output - shows detailed test results
- `-failfast`: Stop testing on first failure
- `-coverage`: Calculate and display test coverage information
- `-package <path>`: Test only a specific package (default is all packages with "./...")
- `-timeout <duration>`: Set a custom timeout for tests (default: 10m)

## Examples

### Run all tests in the current directory
```
go run cmd/modtest/main.go
```

### Run tests in a specific module
```
go run cmd/modtest/main.go /path/to/your/module
```

### Run only tests starting with "TestUser"
```
go run cmd/modtest/main.go . TestUser
```

### Run only tests starting with "TestAPI" in a specific module with verbose output
```
go run cmd/modtest/main.go -v /path/to/your/module TestAPI
```

### Run tests with coverage analysis
```
go run cmd/modtest/main.go -coverage /path/to/your/module
```

### Run tests for a specific package only
```
go run cmd/modtest/main.go -package=github.com/example/module/pkg1
```

## Output Format

The program runs each test function individually and reports results as they complete:

```
Loading module...
Module: github.com/example/module
Directory: /path/to/your/module

Found 10 test functions matching prefix 'TestUser'

Running tests in package: github.com/example/module/users
  Running test: TestUserCreate... PASSED
  Running test: TestUserUpdate... FAILED
    --- FAIL: TestUserUpdate (0.01s)
        users_test.go:42: Expected user to be updated, got unchanged user

Running tests in package: github.com/example/module/auth
  Running test: TestUserLogin... PASSED
  Running test: TestUserLogout... PASSED

--------------------------------------------------------------------------------
Overall Test Results:
  Total Tests: 4
  Passed: 3
  Failed: 1
  Time: 1.6 s
--------------------------------------------------------------------------------
```

When using the `-coverage` flag, additional coverage information is displayed:

```
--------------------------------------------------------------------------------
Coverage Results:
  Overall Coverage: 75.20%
  Time: 2.3 s
--------------------------------------------------------------------------------

Coverage by File:
  pkg1/file1.go: 80.00%
  pkg1/file2.go: 70.40%
  
Uncovered Functions:
  github.com/example/module/pkg1.UncoveredFunction
```

The command will exit with a non-zero status code if any tests fail. 