# Testing Infrastructure for echo9llama

**Version:** 1.0  
**Date:** December 19, 2025  
**Purpose:** This document outlines the comprehensive testing framework implemented for the `echo9llama` project to ensure code quality, system stability, and production readiness.

---

## 1. Testing Philosophy

The testing strategy for `echo9llama` is built upon a multi-layered approach, ensuring that every component of the system is validated, from individual functions to the full end-to-end behavior of the autonomous agent. The framework is designed to be automated, comprehensive, and deeply integrated into the development lifecycle via GitHub Actions.

## 2. Layers of Testing

The framework is divided into three primary layers:

| Layer | Scope | Purpose | Location |
|:---|:---|:---|:---|
| **Unit Tests** | Individual functions and packages | Verify the correctness of isolated components and business logic. | `core/` (alongside source) |
| **Integration Tests** | Interactions between components | Ensure that different parts of the system work together as expected (e.g., cognitive loop, memory, LLM providers). | `test/integration/` |
| **End-to-End (E2E) Tests** | Full system behavior | Validate the complete application from the perspective of an external user, interacting with the live server API. | `test/e2e/` |

### 2.1. Unit Tests

Unit tests are written using Go's standard `testing` package and the `testify` suite for assertions. They are placed in `_test.go` files alongside the code they are testing. These tests are fast, run in isolation, and do not require external dependencies like databases or APIs.

### 2.2. Integration Tests

Integration tests validate the collaboration between different modules. They are located in the `test/integration` directory and are tagged with `//go:build integration`. These tests may require external services, such as a running Dgraph instance, which are automatically provisioned by the CI workflow using Docker services.

### 2.3. End-to-End (E2E) Tests

E2E tests simulate real-world usage by making HTTP requests to a running instance of the `echoself` server. They are located in `test/e2e` and tagged with `//go:build e2e`. These tests are crucial for verifying that the entire system is functioning correctly from the API level down to the core cognitive components.

## 3. Continuous Integration (CI) Pipeline

A robust CI pipeline has been established using GitHub Actions to automate the entire testing and validation process. Two main workflows have been created:

- **`ci.yaml`**: This workflow runs on every push and pull request to the `main` and `develop` branches. It executes a comprehensive suite of jobs, including:
  - **Unit, Integration, and E2E Tests** across multiple platforms.
  - **Code Quality Checks** using `golangci-lint` and `go mod tidy`.
  - **Security Scans** with `govulncheck` and `gosec`.
  - **Build Verification** for multiple OS/architecture combinations.
  - **Docker Image Build** to ensure the containerization is working.

- **`nightly.yaml`**: This workflow runs daily and performs more intensive, long-running tests that are not practical for every commit. This includes:
  - **Full Cognitive Loop Tests** over extended periods.
  - **Extended Benchmarks** to track performance over time.
  - **Stress and Chaos Tests** to ensure system resilience.

## 4. How to Run Tests Locally

A `Makefile.test` has been created to simplify the process of running tests locally.

### 4.1. Prerequisites

- Go 1.24+
- Docker and Docker Compose (for integration/E2E tests)
- `golangci-lint` (for linting)

### 4.2. Common Commands

| Command | Description |
|:---|:---|
| `make -f Makefile.test unit` | Run all unit tests. |
| `make -f Makefile.test integration` | Run integration tests (requires Docker). |
| `make -f Makefile.test e2e` | Run E2E tests (requires a running server). |
| `make -f Makefile.test test` | Run both unit and integration tests. |
| `make -f Makefile.test lint` | Run the linter. |
| `make -f Makefile.test security` | Run security scans. |
| `make -f Makefile.test coverage` | Generate HTML coverage reports. |
| `make -f Makefile.test dgraph-up` | Start a Dgraph instance for testing. |
| `make -f Makefile.test dgraph-down` | Stop the Dgraph instance. |
| `make -f Makefile.test help` | Display all available commands. |

### 4.3. Example Workflow for Integration Testing

```bash
# 1. Start the required services
make -f Makefile.test dgraph-up

# 2. Run the integration tests
make -f Makefile.test integration

# 3. Stop the services
make -f Makefile.test dgraph-down
```

## 5. Mocks and Fixtures

To facilitate testing, a set of mocks and test data fixtures have been created:

- **`test/mocks/`**: Contains mock implementations of external dependencies, such as `MockLLMProvider` and `MockMemoryStore`, allowing for isolated unit testing of components that rely on them.
- **`test/fixtures/`**: Provides sample data structures (`SampleMemoryNodes`, `SampleCognitiveSteps`) to ensure consistent and repeatable test inputs.

---

This comprehensive testing framework provides a strong foundation for the continued evolution of `echo9llama`, ensuring that new features can be added with confidence and that the system remains robust, secure, and production-ready.
