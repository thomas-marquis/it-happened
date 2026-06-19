# Quickstart: Test Coverage Improvement

**Feature**: Test Coverage Improvement  
**Date**: 2026-06-19  
**Spec**: [spec.md](./spec.md)  
**Plan**: [plan.md](./plan.md)  

## Overview

This guide provides runnable validation scenarios to verify that the Test Coverage Improvement feature works end-to-end. It covers the key components: testing infrastructure, coverage measurement, and documentation updates.

## Prerequisites

1. Go 1.25+ installed
2. Git repository cloned
3. Dependencies installed: `go mod download`
4. Mocks generated: `go generate ./...`

## Setup

```bash
# Clone the repository (if not already done)
git clone github.com/thomas-marquis/it-happened
cd it-happened

# Install dependencies
go mod download

# Generate mocks (if gen.go has directives)
go generate ./...

# Verify existing tests pass
go test ./...
```

## Validation Scenarios

### Scenario 1: Run Existing Tests

**Purpose**: Verify the existing test infrastructure works correctly.

**Steps**:
```bash
# Run all tests
go test ./...

# Expected output:
# PASS
# ok  	github.com/thomas-marquis/it-happened/event	0.001s
# ok  	github.com/thomas-marquis/it-happened/carrier	0.001s  
# ok  	github.com/thomas-marquis/it-happened/inmemory	0.001s
```

**Success Criteria**:
- [ ] All existing tests pass
- [ ] No race conditions detected (if running with `-race`)

---

### Scenario 2: Compute Current Coverage

**Purpose**: Establish baseline coverage before adding new tests.

**Steps**:
```bash
# Compute coverage for all packages
go test ./... -cover

# Get detailed coverage report
go test ./... -coverprofile=coverage.out

# View function-level coverage
go tool cover -func=coverage.out

# View coverage by file
go tool cover -func=coverage.out | grep -E "event/|carrier/|inmemory/"
```

**Expected Output**:
```
github.com/thomas-marquis/it-happened/event:	50.0%
github.com/thomas-marquis/it-happened/carrier:	0.0%
github.com/thomas-marquis/it-happened/inmemory:	0.0%
```

**Success Criteria**:
- [ ] Coverage percentages are displayed for each package
- [ ] Carrier and inmemory packages show 0% or low coverage (untested)

---

### Scenario 3: Verify Race Detection Works

**Purpose**: Ensure the race detector is functioning correctly.

**Steps**:
```bash
# Run tests with race detector
go test -race ./...
```

**Success Criteria**:
- [ ] All tests pass with `-race` flag
- [ ] No race conditions are detected

---

### Scenario 4: Verify Mock Generation

**Purpose**: Ensure mocks can be generated for interfaces.

**Steps**:
```bash
# Check if gen.go exists and has mock generation directives
cat gen.go

# Generate/update mocks
go generate ./...

# Verify mocks were generated
ls -la internal/mocks/
```

**Expected Output**:
- [ ] gen.go contains `//go:generate mockgen` directives
- [ ] Mocks are generated in internal/mocks/ directory
- [ ] Mock files have _mock.go suffix

**Success Criteria**:
- [ ] Mocks are generated without errors
- [ ] Mock files exist for required interfaces

---

## Package-Specific Validation

### Event Package Tests

**Validate existing matcher tests**:
```bash
# Run only event package tests
go test ./event/... -v

# Expected: matcher_test.go tests pass
```

**Success Criteria**:
- [ ] All existing event tests pass

### Carrier Package Tests

**Validate carrier package has no tests (baseline)**:
```bash
# List test files in carrier package
ls -la ./carrier/*_test.go 2>/dev/null || echo "No test files found"

# Expected: No test files or only newly created ones
```

**Success Criteria**:
- [ ] No existing test files in carrier/ (or only new ones we create)

### Inmemory Package Tests

**Validate inmemory package has no tests (baseline)**:
```bash
# List test files in inmemory package
ls -la ./inmemory/*_test.go 2>/dev/null || echo "No test files found"

# Expected: No test files or only newly created ones
```

**Success Criteria**:
- [ ] No existing test files in inmemory/ (or only new ones we create)

---

## End-to-End Validation

### Full Test Suite with Coverage

**Purpose**: Run complete validation of the test infrastructure.

**Steps**:
```bash
# Clean any previous coverage files
rm -f coverage.out

# Run all tests with coverage and race detection
go test -race -cover ./... -coverprofile=coverage.out

# View coverage summary
go tool cover -func=coverage.out

# Clean up
rm -f coverage.out
```

**Expected Output**:
```
# All tests pass
PASS
ok      github.com/thomas-marquis/it-happened/event        0.002s  coverage: 50.0% of statements
ok      github.com/thomas-marquis/it-happened/carrier      0.001s  coverage: 0.0% of statements
ok      github.com/thomas-marquis/it-happened/inmemory     0.001s  coverage: 0.0% of statements
```

**Success Criteria**:
- [ ] All tests pass
- [ ] Coverage is computed for each package
- [ ] No race conditions detected

---

## Documentation Validation

### Scenario 5: Verify CONTRIBUTE.md Structure

**Purpose**: Ensure CONTRIBUTE.md has the expected structure for adding Testing Strategy.

**Steps**:
```bash
# Check if CONTRIBUTE.md exists
ls -la CONTRIBUTE.md

# View the file structure
grep "^## " CONTRIBUTE.md

# Expected sections:
# ## Contributing
# ## Development Workflow
# ## Quality Standards
```

**Success Criteria**:
- [ ] CONTRIBUTE.md exists
- [ ] File has proper Markdown structure
- [ ] Quality Standards section exists (where Testing Strategy will be added)

---

### Scenario 6: Verify README.md Structure

**Purpose**: Ensure README.md has the expected structure for adding coverage badge.

**Steps**:
```bash
# View the badges section
head -10 README.md

# Expected: Existing badges are present
```

**Success Criteria**:
- [ ] README.md exists
- [ ] README.md has existing badges (Go Reference, CI, License)
- [ ] Badge section is at the top of the file

---

## Feature-Specific Validation

### Validate New Test Files

**After implementation**, verify new test files exist:

```bash
# List all test files
dind . -name "*_test.go" -type f | sort

# Expected: New test files in event/, carrier/, inmemory/
```

**Success Criteria**:
- [ ] `event/bus_test.go` exists (or similar)
- [ ] `event/notifier_test.go` exists
- [ ] `event/option_test.go` exists
- [ ] `event/subscriber_test.go` exists
- [ ] `carrier/all_test.go` exists
- [ ] `carrier/carrier_test.go` exists
- [ ] `carrier/sequence_test.go` exists
- [ ] `inmemory/bus_test.go` exists
- [ ] `inmemory/options_test.go` exists

### Validate Coverage Improvement

**After implementation**, verify coverage has improved:

```bash
# Run coverage
go test ./... -coverprofile=coverage.out

# Check coverage for each target package
go tool cover -func=coverage.out | grep -E "event/|carrier/|inmemory/"

# Clean up
rm -f coverage.out
```

**Success Criteria**:
- [ ] event package coverage >= 80%
- [ ] carrier package coverage >= 80%
- [ ] inmemory package coverage >= 80%

### Validate Makefile Coverage Target

**After implementation**, verify Makefile has coverage target:

```bash
# Check Makefile
cat Makefile

# Run the coverage target (if implemented)
make coverage 2>&1 || echo "Coverage target not yet implemented"
```

**Success Criteria**:
- [ ] Makefile has a `coverage` target
- [ ] `make coverage` runs without errors
- [ ] Coverage report is displayed

### Validate README.md Badge

**After implementation**, verify README.md has coverage badge:

```bash
# Check for coverage badge
grep -i "coverage" README.md

# Expected: Line containing coverage badge markdown
```

**Success Criteria**:
- [ ] README.md contains a coverage badge
- [ ] Badge uses shields.io or similar service
- [ ] Badge links to coverage reports or workflow

---

## Troubleshooting

### Issue: Tests fail with race conditions

**Diagnosis**: Concurrent operations may have race conditions.

**Solution**:
```bash
# Run with race detector
go test -race ./...

# Fix any detected race conditions in the implementation
```

### Issue: Coverage not computed

**Diagnosis**: Tests may not be running or coverage profile not generated.

**Solution**:
```bash
# Verify tests pass first
go test ./...

# Then compute coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Issue: Mocks not generated

**Diagnosis**: gen.go may be missing directives or mockgen not installed.

**Solution**:
```bash
# Install mockgen
go install go.uber.org/mock/mockgen@latest

# Verify gen.go has directives
cat gen.go

# Generate mocks
go generate ./...
```

---

## Completion Checklist

Before considering this feature complete, verify all scenarios pass:

- [ ] Scenario 1: Run existing tests - PASS
- [ ] Scenario 2: Compute current coverage - PASS
- [ ] Scenario 3: Verify race detection - PASS
- [ ] Scenario 4: Verify mock generation - PASS
- [ ] Scenario 5: Verify CONTRIBUTE.md structure - PASS
- [ ] Scenario 6: Verify README.md structure - PASS
- [ ] New test files exist for all untested components - PASS
- [ ] Coverage >= 80% for all packages - PASS
- [ ] Makefile has coverage target - PASS
- [ ] README.md has coverage badge - PASS

**Feature is ready for implementation when all prerequisites and scenarios pass.**
