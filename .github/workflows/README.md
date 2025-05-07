# Go Project CI/CD Workflows

This directory contains GitHub Actions workflows that implement a comprehensive CI/CD pipeline for this Go project.

## Workflow Overview

### PR Check (`pr-check.yml`)
Triggered on pull requests to `main` and `dev` branches.
- Fast validation with short tests
- Basic code quality checks
- Dependency security review
- Optimized for quick feedback

### Development Pipeline (`dev-pipeline.yml`)
Triggered on pushes to the `dev` branch.
- Full test suite with race detection
- Code coverage reporting
- Advanced linting
- Security scanning
- Build verification

### Main Pipeline (`main-pipeline.yml`)
Triggered on pushes to the `main` branch and on release creation.
- Comprehensive tests (highest standards)
- Strict linting and security analysis
- Cross-platform builds (Linux, Windows, macOS)
- Release automation
- Documentation deployment

## Architecture Decisions

1. **Environment-Based Separation:** Different workflows for different environments ensure appropriate levels of testing at each stage.
2. **Progressive Validation:** Tests become more comprehensive as code moves toward production.
3. **Parallel Jobs:** Tests, linting, and security scans run in parallel to minimize workflow duration.
4. **Matrix Builds:** Cross-platform compilation ensures compatibility across operating systems.
5. **Artifact Management:** Build artifacts are preserved and published with releases.

## Setup Requirements

- GitHub repository with appropriate branch protection rules
- CodeCov account for test coverage monitoring (optional)
- Go modules for dependency management

## Customizing

Each workflow can be customized as needed:
1. Adjust Go version in each workflow as required
2. Modify the matrix build settings for your specific targets
3. Update documentation generation for your project needs
4. Configure branch names to match your development workflow 