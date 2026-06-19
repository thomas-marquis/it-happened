# Data Model: Test Coverage Improvement

**Feature**: Test Coverage Improvement  
**Date**: 2026-06-19  
**Spec**: [spec.md](./spec.md)  
**Plan**: [plan.md](./plan.md)

## Overview

This document describes the data entities, their relationships, and validation rules for the Test Coverage Improvement feature. Since this is a testing-focused feature, the "data model" primarily consists of test artifacts, coverage metrics, and documentation components.

## Test Artifacts

### Test File Entity

**Description**: A Go test file containing unit tests for a specific package.

**Structure**:
```go
type TestFile struct {
    Package     string    // e.g., "event", "carrier", "inmemory"
    Name        string    // e.g., "bus_test.go", "carrier_test.go"
    Functions   []TestCase
    Coverage    float64   // Percentage of package code covered
}
```

**Fields**:
- `Package`: The Go package being tested (required, one of: event, carrier, inmemory)
- `Name`: Filename following Go conventions (*_test.go) (required)
- `Functions`: Collection of test cases in the file (required, at least one)
- `Coverage`: Percentage of package code covered by tests (computed, 0-100)

**Validation Rules**:
- Must follow Go naming conventions
- Must be placed in the same directory as the package being tested
- Must import "testing" package
- Must use testify/assert or testify/require for assertions
- Must use t.Run() for subtests
- Must include Given/When/Then comments

**Relationships**:
- Belongs to one Package
- Contains many TestCase entities
- Contributes to Package Coverage Metric

---

### Test Case Entity

**Description**: An individual test function that verifies a specific behavior.

**Structure**:
```go
type TestCase struct {
    ID          string   // e.g., "TestBus_Publish"
    Name        string   // Human-readable name
    Description string   // What behavior is being tested
    Package     string   // Package under test
    Subtests    []Subtest
    Tags        []string // e.g., ["unit", "inmemory", "concurrent"]
}
```

**Fields**:
- `ID`: Unique identifier (Go function name) (required)
- `Name`: Descriptive name of the test (required)
- `Description`: Plain-language description of what's being tested (required)
- `Package`: The package containing the code under test (required)
- `Subtests`: Nested test cases using t.Run() (optional)
- `Tags`: Categorization tags for filtering (optional)

**Validation Rules**:
- Name must start with "Test" followed by PascalCase
- Must have exactly one `*testing.T` parameter
- Must call t.Run() for each scenario if using subtests
- Must include Given/When/Then comments
- Must use testify assertions

**Relationships**:
- Belongs to one TestFile
- Tests one or more Components
- May use Mock entities
- Contains zero or more Subtest entities

---

### Subtest Entity

**Description**: A nested test case created with t.Run() for better test organization.

**Structure**:
```go
type Subtest struct {
    Name        string // e.g., "publishes to all subscribers"
    Description string // What scenario this subtest covers
    Given       string // Initial state setup
    When        string // Action being tested
    Then        string // Expected outcome
}
```

**Fields**:
- `Name`: Descriptive name of the subtest scenario (required)
- `Description`: Additional context (optional)
- `Given`: Setup conditions (from Given comment) (required)
- `When`: Action under test (from When comment) (required)
- `Then`: Expected outcome (from Then comment) (required)

**Validation Rules**:
- Name must be descriptive of the scenario
- Must have all three: Given, When, Then
- Must be independent of other subtests

**Relationships**:
- Belongs to one TestCase
- Tests specific behavior of a Component

---

### Component Entity

**Description**: A code component that needs to be tested.

**Entities**:

#### Event Package Components

1. **Bus Interface** (`event/bus.go`)
   - Methods: `Publish(evt Event)`, `Subscribe() *Subscriber`
   - Test Focus: Interface contract verification
   - Mock: May need mock for testing subscribers

2. **Event** (`event/event.go`)
   - Test Focus: Event creation, properties, ChainRef, ChainPosition
   - Already has some coverage via matcher tests

3. **Matcher** (`event/matcher.go`)
   - Status: Already has tests (matcher_test.go)
   - Coverage: Existing tests to be verified

4. **Notifier** (`event/notifier.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: Notification delivery to subscribers

5. **Subscriber** (`event/subscriber.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: Handler registration, matching, unsubscription

6. **Option** (`event/option.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: All configuration options

#### Carrier Package Components

1. **Carrier Interface** (`carrier/carrier.go`)
   - Methods: `Dispatch(bus event.Bus)`, `EventType() string`
   - Test Focus: Interface contract verification

2. **All Carrier** (`carrier/all.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: Parallel event dispatching, error handling
   - Behavior: Dispatches all events concurrently

3. **Sequence Carrier** (`carrier/sequence.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: Sequential event dispatching, order verification
   - Behavior: Dispatches events one at a time in order

4. **CompletionCondition** (`carrier/carrier.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: Custom completion conditions

#### Inmemory Package Components

1. **Inmemory Bus** (`inmemory/bus.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: Event publishing, subscription, concurrent operations
   - Behavior: Thread-safe in-memory event bus implementation

2. **Options** (`inmemory/options.go`)
   - Status: **Untested** - requires test coverage
   - Test Focus: Bus configuration options

---

### Mock Entity

**Description**: A generated test double for an interface.

**Structure**:
```go
type Mock struct {
    InterfaceName string   // e.g., "Bus", "Subscriber"
    Package       string   // Package where mock is used
    Location      string   // File path: internal/mocks/<package>_mock.go
    Methods       []string // Methods mocked from the interface
}
```

**Fields**:
- `InterfaceName`: Name of the interface being mocked (required)
- `Package`: Package where the mock will be used (required)
- `Location`: Generated file location (computed)
- `Methods`: List of interface methods that have mock implementations (computed)

**Validation Rules**:
- Must be generated using mockgen (gomock)
- Must NOT be edited manually
- Must be placed in internal/mocks/ directory
- Must be regenerated using `go generate ./...`

**Generation Directives** (in gen.go):
```go
//go:generate mockgen -source=event/bus.go -destination=internal/mocks/bus_mock.go -package=mocks
//go:generate mockgen -source=event/subscriber.go -destination=internal/mocks/subscriber_mock.go -package=mocks
```

**Relationships**:
- Implements one Interface
- Used by TestCase entities
- Generated from gen.go directives

---

## Coverage Metrics

### Package Coverage Entity

**Description**: Coverage statistics for a Go package.

**Structure**:
```go
type PackageCoverage struct {
    Package     string  // e.g., "event", "carrier", "inmemory"
    Coverage    float64 // Percentage (0-100)
    Functions   int     // Number of functions
    Covered    int     // Number of functions covered
    Timestamp   string  // When coverage was measured
    MinTarget  float64 // Minimum acceptable coverage (80%)
}
```

**Fields**:
- `Package`: Go package name (required)
- `Coverage`: Percentage of statements covered (computed, 0-100)
- `Functions`: Total number of functions in package (computed)
- `Covered`: Number of functions with at least one test (computed)
- `Timestamp`: When coverage measurement was taken (computed)
- `MinTarget`: Minimum acceptable coverage percentage (default: 80)

**Validation Rules**:
- Coverage must be >= MinTarget for feature completion
- Measured using `go test -cover ./<package>`

**Relationships**:
- Aggregates coverage from all TestFile entities in the package
- Contributes to overall project coverage

---

### Project Coverage Entity

**Description**: Aggregate coverage statistics for the entire project.

**Structure**:
```go
type ProjectCoverage struct {
    Packages     []PackageCoverage
    TotalCoverage float64
    AllPass      bool
}
```

**Validation Rules**:
- All packages must meet MinTarget coverage
- Measured using `go test -cover ./...`

---

## Documentation Entities

### Testing Strategy Documentation

**Description**: Documentation section in CONTRIBUTE.md explaining how to test library components.

**Required Sections**:
1. **Testing Framework**: testify, gomock, go test
2. **Test Structure**: t.Run, Given/When/Then, subtests
3. **Mocking**: When to mock, mock generation, mock usage
4. **Asynchronous Testing**: Testing event-driven components
5. **Coverage**: How to compute and interpret coverage
6. **Best Practices**: General testing recommendations

**Validation Rules**:
- Must be added to CONTRIBUTE.md
- Must be clear and actionable
- Must include examples for testing event buses and carriers

---

### Coverage Badge Entity

**Description**: Visual indicator in README.md showing current test coverage.

**Structure**:
```markdown
![Coverage](https://img.shields.io/badge/Coverage-XX%25-brightgreen)
```

**Fields**:
- `URL`: Link to coverage report or workflow
- `Percentage`: Current coverage percentage (computed from ProjectCoverage)
- `Color`: Badge color based on coverage (green >= 80%, yellow >= 60%, red < 60%)

**Validation Rules**:
- Must be added to README.md
- Must link to coverage reports or workflow
- Must be automatically updated

**Implementation**:
- Generated by GitHub Actions workflow
- Badge SVG committed to repository
- README.md updated with badge markdown

---

## State Transitions

### Test Development Lifecycle

```
[Unimplemented] --(Write test that fails)--> [Red]
[Red] --(Implement code)--> [Green]
[Green] --(Refactor)--> [Refactored]
[Refactored] --(All tests pass)--> [Complete]
```

**States**:
1. **Unimplemented**: No test exists for the component
2. **Red**: Test exists and fails (expected before implementation)
3. **Green**: Test exists and passes
4. **Refactored**: Test passes after code improvements
5. **Complete**: All acceptance criteria met

**Validation**:
- Must follow Red-Green-Refactor cycle (constitution principle II)
- Cannot skip from Unimplemented to Green

---

### Coverage Lifecycle

```
[Unknown] --(Run tests with -cover)--> [Measured]
[Measured] --(Evaluate against threshold)--> [Pass/Fail]
[Fail] --(Add more tests)--> [Measured]
[Pass] --(All packages >= 80%)--> [Acceptable]
```

**States**:
1. **Unknown**: Coverage not yet measured
2. **Measured**: Coverage percentage computed
3. **Pass**: Coverage >= MinTarget (80%)
4. **Fail**: Coverage < MinTarget (80%)
5. **Acceptable**: All packages pass minimum threshold

---

## Validation Rules Summary

### Test Validation
- [ ] All tests use testify/assert or testify/require
- [ ] All tests use t.Run() for subtests
- [ ] All tests include Given/When/Then comments
- [ ] All tests pass `go test -race ./...`
- [ ] No manual mock edits
- [ ] Mocks generated via gen.go

### Coverage Validation
- [ ] Coverage measured using `go test -cover ./...`
- [ ] Each package achieves >= 80% coverage
- [ ] Coverage badge in README.md is up-to-date

### Documentation Validation
- [ ] CONTRIBUTE.md contains Testing Strategy section
- [ ] Testing Strategy includes asynchronous testing guidance
- [ ] Testing Strategy includes carrier testing examples
- [ ] Testing Strategy includes bus testing examples
- [ ] Testing Strategy documents coverage computation

### Quality Gates Validation
- [ ] All tests pass
- [ ] All linting passes (./tools/lint.sh)
- [ ] CI workflows pass
- [ ] Code is clean and documented