# Research: Test Coverage Improvement

**Feature**: Test Coverage Improvement  
**Date**: 2026-06-19  
**Spec**: [spec.md](./spec.md)  
**Plan**: [plan.md](./plan.md)

## Technical Decisions

### Decision 1: Mock Generation Tool

**Decision**: Use gomock (go.uber.org/mock/gomock) for mock generation

**Rationale**: 
- Already specified in the project's constitution (Quality Standards section)
- Industry-standard for Go projects
- Integrates well with testify for assertions
- User explicitly confirmed this choice in planning input
- Consistent with existing project setup (gen.go file mentioned for mock generation)

**Alternatives considered**:
- testifymock: Less commonly used, would require new dependency
- mockery: Another popular option, but gomock is already in constitution
- Manual mocks: Error-prone, violates constitution principle of using mockgen

**Implementation notes**:
- Mocks will be generated via `go generate ./...` command
- Mocks will be stored in `internal/mocks/` directory
- Mocks must NOT be edited manually
- gen.go file will contain mock generation directives

---

### Decision 2: Assertion Library

**Decision**: Use testify (github.com/stretchr/testify) for assertions

**Rationale**:
- Already specified in the project's constitution (Quality Standards section)
- Industry-standard for Go projects
- Provides both `assert` and `require` packages
- User explicitly confirmed this choice in planning input

**Alternatives considered**:
- Standard library testing: Lacks rich assertion helpers
- testify only: Already the decision
- Other assertion libraries: Would violate constitution

**Implementation notes**:
- Use `testify/assert` for assertions
- Use `testify/require` for setup preconditions
- All tests must use t.Run() subtests with Given/When/Then comments

---

### Decision 3: Mock Generation Location

**Decision**: Generate mocks in gen.go file (if needed only)

**Rationale**:
- User explicitly specified this in planning input
- Centralized location for all code generation directives
- Follows Go project conventions
- Consistent with existing project structure (gen.go already exists)

**Alternatives considered**:
- Per-package mock generation: Could lead to duplication
- Separate mock generation file: gen.go already serves this purpose
- Inline generation in test files: Harder to maintain

**Implementation notes**:
- Mocks will only be generated if needed (for interfaces that require mocking in tests)
- gen.go will contain `//go:generate` directives for mockgen
- Mocks will be regenerated using `go generate ./...`

---

### Decision 4: Carrier Test Readability

**Decision**: Carriers' tests must be as readable as possible

**Rationale**:
- User explicitly specified this as a priority in planning input
- Carrier implementations (All, Sequence) have complex asynchronous behavior
- Readable tests are essential for maintainability and debugging
- Aligns with constitution principle of Simplicity and Composability

**Implementation Strategy**:
- Use descriptive test names that explain the scenario being tested
- Leverage table-driven tests for similar test cases
- Use subtests (t.Run) to group related test scenarios
- Add clear Given/When/Then comments in each test
- Keep test logic simple and focused on one behavior at a time
- Use helper functions to reduce test boilerplate
- Document complex test setups with comments

**Readability Techniques**:
```go
// Example of readable test structure
func TestCarrierDispatch(t *testing.T) {
    t.Run("All carrier dispatches events in parallel", func(t *testing.T) {
        // Given: a carrier with multiple events and a bus with subscribers
        carrier := NewAllCarrier(events...)
        bus := inmemory.NewBus()
        
        // When: the carrier dispatches events
        carrier.Dispatch(bus)
        
        // Then: all events are published to the bus
        // ... assertions
    })
}
```

---

### Decision 5: Coverage Badge Implementation

**Decision**: Generate coverage badge via GitHub Actions workflow

**Rationale**:
- Clarified during /speckit-clarify session
- No external service dependencies required
- Maintains control over badge generation
- Consistent with existing CI workflow

**Alternatives considered**:
- codecov.io: External service, would add dependency
- coveralls.io: External service, would add dependency
- Manual badge updates: Error-prone, not automated

**Implementation notes**:
- Workflow will compute coverage using `go test -cover ./...`
- Badge SVG will be committed to the repository
- README.md will be updated to display the badge
- Workflow will run on pushes to main branch

---

### Decision 6: Local Coverage Computation

**Decision**: Add Makefile target for local coverage computation

**Rationale**:
- Clarified during /speckit-clarify session
- Provides simple, documented method for maintainers
- Consistent with project conventions (Makefile already exists)
- Easy to remember and use (`make coverage`)

**Alternatives considered**:
- Direct command only: Less discoverable
- Shell script in tools/: Would work but Makefile is more conventional for this use case
- Multiple methods: Unnecessary complexity

**Implementation notes**:
- Makefile will include a `coverage` target
- Target will run `go test -cover ./...` and display coverage report
- Will be documented in CONTRIBUTE.md Testing Strategy section

---

## Best Practices for Testing Event-Driven Components

### Testing Asynchronous Components

**Challenge**: Event buses and carriers operate asynchronously, making it difficult to verify event delivery timing.

**Best Practices**:
1. **Use synchronization primitives**: sync.WaitGroup, channels, or mutexes to coordinate test goroutines
2. **Avoid timing-based assertions**: Tests should be deterministic, not rely on timeouts
3. **Use subtests for isolation**: Each test scenario should be independent
4. **Verify side effects**: Instead of testing timing, verify that expected state changes occurred
5. **Test concurrent scenarios**: Use `-race` flag to detect race conditions

### Testing Event Carriers

**All Carrier**:
- Verify all events are dispatched
- Test concurrent dispatch behavior
- Verify error handling for individual event failures

**Sequence Carrier**:
- Verify events are dispatched in order
- Test that each event completes before the next begins
- Verify error handling stops the sequence

### Testing Event Bus

**Inmemory Bus**:
- Verify event publication to all subscribers
- Test event filtering with matchers
- Verify thread-safety under concurrent publish/subscribe
- Test subscriber registration/unregistration

### Mocking Strategy

**When to Mock**:
- Mock external dependencies (not part of the library)
- Mock interfaces to isolate components under test
- DO NOT mock components within the same package being tested

**When NOT to Mock**:
- Concrete implementations within the same package
- Simple data structures or value types
- Components that are the focus of the test

## Quality Assurance Checklist

- [ ] All tests use testify/assert or testify/require
- [ ] All tests use t.Run() for subtests
- [ ] All tests include Given/When/Then comments
- [ ] All mocks are generated via gen.go using mockgen
- [ ] All mocks are stored in internal/mocks/
- [ ] Mocks are NOT edited manually
- [ ] Carrier tests are highly readable
- [ ] All tests pass `go test -race ./...`
- [ ] Coverage meets 80% threshold for each package
- [ ] Makefile includes coverage target
- [ ] CONTRIBUTE.md includes Testing Strategy section
- [ ] README.md displays coverage badge