# Testing Guide

This document describes how to run tests for this project.

## Test Types

The project has two types of tests:

1. **Unit Tests**: Fast-running tests that verify individual components in isolation
2. **Integration Tests**: Slower tests that verify components working together (require DB)

## Running Tests

### Run all tests
Runs both unit and integration tests (requires DB):
```
make test
```

### Run only unit tests
```
make test-unit
```

### Run only integration tests
```
make test-integration
```
