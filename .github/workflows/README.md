# Go Project CI/CD Workflows

This directory contains GitHub Actions workflows that implement a comprehensive CI/CD pipeline for this Go project.

## Workflow Structure

We use a reusable workflow pattern to ensure consistency and prevent drift between environments:

- **Shared Go Checks** (`shared-go-checks.yml`) - Core reusable workflow with all tests, linting, and security checks
- **Feature Branch Check** - Minimal checks for all feature branches
- **PR Check** - Complete validation for pull requests to main/dev
- **Development Pipeline** - Full testing for the dev branch
- **Main Pipeline** - Production validation plus release management

## Workflow Overview

### Shared Go Checks (`shared-go-checks.yml`)
Reusable workflow that's called by all other workflows.
- Parameterized to allow different behavior based on environment
- Ensures identical testing methodology across all pipelines
- Prevents test configuration drift

### Feature Branch Check (`feature-check.yml`)
Triggered on pushes to all branches except main and dev.
- Basic validation using shared workflow
- Optimized for fast feedback during development

### PR Check (`pr-check.yml`)
Triggered on pull requests to `main` and `dev` branches.
- Complete validation using shared workflow
- Dependency security review
- Ensures PRs contain the same quality checks as main

### Development Pipeline (`dev-pipeline.yml`)
Triggered on pushes to the `dev` branch.
- Full test suite with race detection
- Code coverage reporting

### Main Pipeline (`main-pipeline.yml`)
Triggered on pushes to the `main` branch and on release creation.
- Complete validation using shared workflow with strictest settings
- Cross-platform builds (Linux, Windows, macOS)
- Release automation
- Documentation deployment

## Architecture Benefits

1. **Single Source of Truth:** Test logic exists in only one place
2. **Automatic Consistency:** All pipelines use identical validation rules
3. **Future-Proof:** Adding a new check automatically applies to all environments
4. **Parameterization:** Different levels of validation for different contexts
5. **Maintainability:** Easier to understand and modify the pipeline

## Setup Requirements

- GitHub repository with appropriate branch protection rules
- Go modules for dependency management
- CodeCov integration:
  1. Sign up at [codecov.io](https://codecov.io/)
  2. Connect your GitHub repository
  3. Get your CodeCov token
  4. Add the token as a GitHub secret named `CODECOV_TOKEN`

## Customizing

To modify the shared validation process:
1. Edit the `shared-go-checks.yml` file to add or modify checks
2. Adjust parameters in the calling workflows as needed

To add new environment-specific steps:
1. Add them directly to the relevant workflow file 