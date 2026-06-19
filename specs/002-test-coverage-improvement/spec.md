# Feature Specification: Test Coverage Improvement

**Feature Branch**: `TBD`

**Created**: 2026-06-19

**Status**: Draft

**Input**: User description: "I want to increase the project's test coverage. Implementing unit tests for untested components and remaining edge cases. In particular, I would like to find a good, readable and reliable way to tests the event carriers and event bus. These components may be challenging to test due to their asynchronous nature. Furthermore, the testing strategy must be detailed in the contributor documentation."

## Clarifications

### Session 2026-06-19

- Q: How should the coverage badge be implemented? → A: Generate badge via GitHub Actions
- Q: What local coverage computation method should be documented for maintainers? → A: Add a Makefile target

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developers can write effective tests for asynchronous event bus components (Priority: P1)

Developers need to be able to write reliable, readable unit tests for the event bus interface and its implementations. The testing approach must handle the asynchronous nature of event publishing and subscription, ensuring that tests can verify event delivery without race conditions. This includes testing the inmemory bus implementation which currently has no test coverage.

**Why this priority**: The event bus is the core abstraction of the library. Without proper tests, regressions in this critical component could break the entire library for users. This is foundational for all other testing efforts.

**Independent Test**: Can be fully tested by having developers write unit tests that verify event publishing, subscription, and delivery work correctly, including edge cases like multiple subscribers, event filtering, and concurrent operations.

**Acceptance Scenarios**:

1. **Given** a test for the inmemory bus, **When** an event is published, **Then** all registered subscribers receive the event
2. **Given** a test with multiple subscribers, **When** an event is published, **Then** each subscriber's callback is invoked with the correct event
3. **Given** a test with concurrent event publishing, **When** multiple events are published simultaneously, **Then** all events are delivered correctly without data races
4. **Given** a test with event matchers, **When** an event is published, **Then** only subscribers with matching criteria receive the event

---

### User Story 2 - Developers can write effective tests for event carrier components (Priority: P1)

Developers need to be able to write reliable unit tests for event carrier implementations (All, Sequence, and the Carrier interface). The testing approach must handle the asynchronous dispatching of multiple events and verify that all events in a carrier are properly delivered to the bus. This includes testing carrier configuration options and completion conditions.

**Why this priority**: Event carriers are a key feature for batch event dispatching. Without proper tests, the carrier functionality could have subtle bugs that are hard to detect in production.

**Independent Test**: Can be fully tested by having developers write unit tests that verify carrier creation, event dispatching, configuration options, and completion conditions work as expected.

**Acceptance Scenarios**:

1. **Given** a test for the All carrier, **When** Dispatch is called, **Then** all events in the carrier are published to the bus
2. **Given** a test for the Sequence carrier, **When** Dispatch is called, **Then** events are published sequentially to the bus
3. **Given** a test with carrier options, **When** a carrier is created with timeout/concurrency options, **Then** the carrier respects these configuration values
4. **Given** a test with completion conditions, **When** a custom completion condition is set, **Then** the carrier uses it to determine event completion

---

### User Story 3 - Developers can test remaining untested event package components (Priority: P2)

Developers need comprehensive test coverage for all event package components that currently lack tests, including: notifier.go, option.go, and subscriber.go. Each component should have unit tests covering its public API and edge cases.

**Why this priority**: Complete test coverage of the event package ensures reliability. These components support the core event-driven functionality and need to be tested to prevent regressions.

**Independent Test**: Can be fully tested by writing unit tests for each untested component in the event package, verifying their behavior matches their documented contracts.

**Acceptance Scenarios**:

1. **Given** a test for the notifier, **When** it notifies subscribers, **Then** all registered callbacks are invoked
2. **Given** a test for event options, **When** options are applied, **Then** they correctly configure the event properties
3. **Given** a test for the subscriber, **When** it registers handlers, **Then** the handlers are invoked for matching events
4. **Given** a test for the subscriber, **When** it unsubscribes, **Then** its handlers are no longer invoked

---

### User Story 4 - Contributors have clear testing guidelines in documentation (Priority: P2)

Contributors need clear, comprehensive documentation on how to test the library's components, especially asynchronous patterns for event-driven code. The CONTRIBUTE.md file must be updated with a dedicated testing strategy section that explains approaches, patterns, and best practices for testing event buses, carriers, and other asynchronous components.

**Why this priority**: Good documentation ensures that all contributors (not just the original authors) can write effective tests. This is essential for maintaining high test coverage over time.

**Independent Test**: Can be fully tested by having a new contributor read the testing documentation and successfully write tests for a new component or feature.

**Acceptance Scenarios**:

1. **Given** the CONTRIBUTE.md documentation, **When** a contributor reads the testing strategy section, **Then** they understand how to test asynchronous components
2. **Given** the CONTRIBUTE.md documentation, **When** a contributor needs to test an event bus implementation, **Then** they find clear examples and patterns to follow
3. **Given** the CONTRIBUTE.md documentation, **When** a contributor needs to test event carriers, **Then** they find specific guidance for carrier testing
4. **Given** the CONTRIBUTE.md documentation, **When** a contributor looks for testing best practices, **Then** they find recommendations for test structure, mocking, and assertions

---

### User Story 5 - Achieve minimum test coverage threshold across all packages (Priority: P3)

The project must achieve a minimum test coverage threshold (to be determined) across all packages, with specific focus on the currently untested packages: carrier, inmemory, and the remaining components in event. This provides measurable assurance of code quality.

**Why this priority**: While not as critical as the core functionality tests, having a measurable coverage target helps ensure comprehensive testing and provides a metric for code quality.

**Independent Test**: Can be fully tested by running `go test -cover ./...` and verifying that the coverage percentage meets or exceeds the defined threshold for each package.

**Acceptance Scenarios**:

1. **Given** the test suite, **When** `go test -cover ./...` is run, **Then** the event package achieves at least 80% coverage
2. **Given** the test suite, **When** `go test -cover ./...` is run, **Then** the carrier package achieves at least 80% coverage
3. **Given** the test suite, **When** `go test -cover ./...` is run, **Then** the inmemory package achieves at least 80% coverage
4. **Given** the test suite, **When** `go test ./...` is run, **Then** all tests pass successfully

---

### Edge Cases

- What happens when testing concurrent event publishing with the inmemory bus? (Tests must handle synchronization properly)
- How do we test that event carriers correctly dispatch events when the bus has no subscribers? (Events should still be published without errors)
- What happens when a carrier's timeout is exceeded during dispatch? (Should be tested with controlled timing)
- How do we test that subscribers correctly handle events with nil payloads? (Edge case in event creation)
- What happens when multiple goroutines publish and subscribe simultaneously? (Tests must be thread-safe and detect race conditions)
- How do we verify that the All carrier dispatches events in parallel? (Need to test concurrent execution)
- How do we verify that the Sequence carrier dispatches events sequentially? (Need to test ordered execution)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The inmemory bus MUST have comprehensive unit tests covering Publish, Subscribe, and concurrent operations
- **FR-002**: The inmemory bus tests MUST verify that published events are delivered to all matching subscribers
- **FR-003**: The inmemory bus tests MUST verify thread-safety under concurrent publish and subscribe operations
- **FR-004**: The All carrier MUST have unit tests covering Dispatch with multiple events
- **FR-005**: The Sequence carrier MUST have unit tests covering Dispatch with ordered event delivery
- **FR-006**: All carrier implementations MUST have tests for configuration options (timeout, concurrency, completion conditions)
- **FR-007**: The notifier component MUST have unit tests for notification delivery
- **FR-008**: Event options MUST have unit tests for all configuration possibilities
- **FR-009**: The subscriber component MUST have unit tests for handler registration, matching, and unsubscription
- **FR-010**: All new tests MUST follow the project's existing test conventions (testify/assert, t.Run subtests, Given/When/Then comments)
- **FR-011**: All new tests MUST be able to run with `go test ./...` without errors
- **FR-012**: CONTRIBUTE.md MUST be updated with a new "Testing Strategy" section
- **FR-013**: The Testing Strategy section MUST include specific guidance for testing asynchronous components
- **FR-014**: The Testing Strategy section MUST include examples for testing event buses
- **FR-015**: The Testing Strategy section MUST include examples for testing event carriers
- **FR-016**: The Testing Strategy section MUST document best practices for mocking with gomock
- **FR-017**: The Testing Strategy section MUST explain how to test concurrent operations safely
- **FR-018**: All tests MUST achieve at least 80% code coverage for their respective packages
- **FR-019**: All tests MUST pass the existing linting checks in ./tools/lint.sh
- **FR-020**: A GitHub Actions workflow MUST be created to generate a coverage badge
- **FR-021**: README.md MUST be updated to display the coverage badge
- **FR-022**: CONTRIBUTE.md Testing Strategy section MUST document how maintainers can compute coverage locally
- **FR-023**: A Makefile target MUST be added for local coverage computation

### Key Entities *(include if feature involves data)*

- **Test File**: A Go test file (*_test.go) containing unit tests for a specific package
- **Test Case**: An individual test function that verifies a specific behavior
- **Test Coverage**: The percentage of code executed during test runs
- **Mock**: A generated test double for interfaces, created using gomock
- **Subtest**: A nested test case created with t.Run() for better test organization
- **Race Condition**: A concurrency issue where test results depend on timing of goroutine execution
- **Testing Strategy**: Documentation explaining how to approach testing for different component types
- **Coverage Badge**: A visual indicator in README.md showing current test coverage percentage

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of packages (event, carrier, inmemory) have at least 80% test coverage
- **SC-002**: All currently untested components have dedicated test files with comprehensive test cases
- **SC-003**: All tests pass when running `go test ./...`
- **SC-004**: All tests pass the race detector when running `go test -race ./...`
- **SC-005**: All new tests follow the project's established conventions (testify, t.Run, Given/When/Then)
- **SC-006**: CONTRIBUTE.md contains a complete Testing Strategy section with asynchronous testing guidance
- **SC-007**: At least 3 new contributors can successfully write tests after reading the Testing Strategy documentation
- **SC-008**: The testing approach for asynchronous components is documented and reproducible
- **SC-009**: README.md displays a working coverage badge linked to coverage reports
- **SC-010**: Maintainers can compute coverage locally following documented steps in CONTRIBUTE.md

## Assumptions

- The minimum acceptable test coverage threshold is 80% for each package
- The existing test framework (testify/assert, testify/require, gomock) will continue to be used
- Tests should run quickly enough for local development (aim for sub-second execution for individual test files)
- The inmemory bus implementation is the primary bus implementation that needs testing
- Carrier implementations (All, Sequence) are the primary carrier types that need testing
- The event matcher already has tests and does not need additional coverage in this feature
- Mocks will be generated using the existing `go generate ./...` command
- The testing strategy will be added as a new section in CONTRIBUTE.md, not as a separate document
- All tests must be deterministic and not rely on timing or external systems
- The race detector (`-race` flag) will be used to verify thread-safety in concurrent tests
- The coverage badge will be generated by a GitHub Actions workflow that commits the badge to the repository
- The Makefile will include a `coverage` target that runs tests with coverage and displays the results
