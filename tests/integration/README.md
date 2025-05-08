# Integration Tests

This directory contains integration tests that span multiple packages in the go-tree project.

## Purpose

Integration tests verify that different components work correctly together, testing across package boundaries and ensuring compatibility between tightly coupled packages.

## Running Tests

Run the integration tests with:

```bash
go test -v ./tests/integration/...
```

## Test Files

- `loadersaver_test.go`: Tests the integration between the `pkg/loader` and `pkg/saver` packages, ensuring that:
  - Modules can be loaded with the loader
  - Modified in memory
  - Saved with the saver
  - Reloaded again with the loader

## Writing New Integration Tests

When writing integration tests:

1. Create a new test file in this directory
2. Focus on testing the interaction between two or more packages
3. Use public APIs only (don't access package internals)
4. Set up realistic test data that exercises both packages
5. Clean up temporary files and resources in tests 