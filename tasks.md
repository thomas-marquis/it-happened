# Marble Testing Framework - Task List

## Overview

This file contains the complete todo list for completing the marble testing framework implementation.
Each task corresponds to a specific actionable item from the implementation plan.

---

## Step 1: Verify and Complete Core Functionality

### Core Verification
- [x] **Task 1.1**: Review and verify start event (`^`) handling in runtime
  - File: `eventest/internal/engine/runtime/runtime.go`
  - Check: Start event is parsed but actually used in runtime
  - Action: Test and fix if needed

- [x] **Task 1.2**: Complete TODO test cases in runtime_test.go
  - File: `eventest/internal/engine/runtime/runtime_test.go`
  - Lines: 190-200 (TestRuntime_Run test cases for start event)
  - Action: Implement the two empty test cases

- [x] **Task 1.3**: Fix package name issues in gomockevent tests
  - File: `eventest/gomockevent/matchers_test.go`
  - Issue: Package name in test expectations doesn't match actual package name
  - Action: Update expected strings to match actual package name

- [x] **Task 1.4**: Run all existing tests and fix any failures
  - Command: `go test ./...`
  - Action: Fix any failing tests

- [x] **Task 1.5**: Verify all marble syntax features work
  - Test: Simple events, waits, groups (ordered/unordered), followups, start event
  - Action: Create manual verification tests if needed

---

## Step 2: Write Harness API Tests

### Basic Functionality Tests
- [ ] **Task 2.1**: Create harness_test.go file
  - File: `eventest/harness_test.go`
  - Action: Create file with package declaration and imports

- [ ] **Task 2.2**: Test simple event sequence matching
  - Test: `NewHarness(bus, "abc").Run(t, func() { bus.Publish(...) })`
  - Expected: Test passes when events match in order

- [ ] **Task 2.3**: Test event sequence with waits
  - Test: `NewHarness(bus, "a-b-c").Run(t, ...)`
  - Expected: Test accounts for wait ticks

- [ ] **Task 2.4**: Test with ordered groups
  - Test: `NewHarness(bus, "[ab]").Run(t, ...)`
  - Expected: Events a and b must occur in order within same tick

- [ ] **Task 2.5**: Test with unordered groups
  - Test: `NewHarness(bus, "(ab)").Run(t, ...)`
  - Expected: Events a and b can occur in any order within same tick

- [ ] **Task 2.6**: Test with nested groups
  - Test: `NewHarness(bus, "[a(bc)]").Run(t, ...)`
  - Expected: Complex nested group structure works

- [ ] **Task 2.7**: Test with followup events
  - Test: `NewHarness(bus, "b<-a").Run(t, ...)`
  - Expected: Followup relationship verified

- [ ] **Task 2.8**: Test with start event
  - Test: `NewHarness(bus, "^abc").Run(t, ...)`
  - Expected: Start event handled correctly

### Options Tests
- [ ] **Task 2.9**: Test WithPayloads option
  - Test: `NewHarness(bus, "a", WithPayloads(map[string]event.Payload{"a": myPayload}))`
  - Expected: Event a matches payload

- [ ] **Task 2.10**: Test WithEvents option
  - Test: `NewHarness(bus, "a", WithEvents(map[string]event.Event{"a": myEvent}))`
  - Expected: Event a matches exact event

- [ ] **Task 2.11**: Test WithMatchers option
  - Test: `NewHarness(bus, "a", WithMatchers(map[string]event.Matcher{"a": myMatcher}))`
  - Expected: Event a matches custom matcher

- [ ] **Task 2.12**: Test WithSideEffect option
  - Test: `NewHarness(bus, "abc", WithSideEffect("xy"))`
  - Expected: Side effect marble executed, then expected sequence verified

- [ ] **Task 2.13**: Test WithTickDuration option
  - Test: `NewHarness(bus, "a-b", WithTickDuration(100*time.Millisecond))`
  - Expected: Custom tick duration used

### Error Case Tests
- [ ] **Task 2.14**: Test missing event
  - Test: Expected "abc", actual "ab"
  - Expected: Test fails with clear error message

- [ ] **Task 2.15**: Test extra event
  - Test: Expected "ab", actual "abc"
  - Expected: Test fails with clear error message

- [ ] **Task 2.16**: Test wrong event order
  - Test: Expected "abc", actual "acb"
  - Expected: Test fails with clear error message

- [ ] **Task 2.17**: Test wrong event in group
  - Test: Expected "(ab)", actual "ac"
  - Expected: Test fails (unordered group still has expected events)

- [ ] **Task 2.18**: Test empty marble string
  - Test: `NewHarness(bus, "")`
  - Expected: Returns error or handles gracefully

- [ ] **Task 2.19**: Test invalid marble syntax
  - Test: `NewHarness(bus, "a!!b")`
  - Expected: Returns parse error

---

## Step 3: Integration Testing

### End-to-End Tests
- [ ] **Task 3.1**: Test complete workflow with inmemory bus
  - Test: Create bus, harness, publish events, verify
  - Expected: Full integration works

- [ ] **Task 3.2**: Test multiple harnesses on same bus
  - Test: Two harnesses verify different sequences on same bus
  - Expected: Both can verify independently

- [ ] **Task 3.3**: Test clock synchronization
  - Test: Verify clock advances correctly through ticks
  - Expected: Timing is consistent

- [ ] **Task 3.4**: Test with real event types
  - Test: Use application-specific event types
  - Expected: Works with custom payload types

### Complex Scenario Tests
- [ ] **Task 3.5**: Test long marble sequence
  - Test: 50+ events in sequence
  - Expected: Handles large sequences correctly

- [ ] **Task 3.6**: Test deeply nested groups
  - Test: `"[(a[b(cd)])]"`
  - Expected: Handles deep nesting correctly

- [ ] **Task 3.7**: Test mixed features
  - Test: `"^a-(bc)[de]f<-a"`
  - Expected: All features work together

---

## Step 4: Code Quality Improvements

### String() Methods
- [ ] **Task 4.1**: Add String() to EventOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns event name

- [ ] **Task 4.2**: Add String() to WaitOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns "-"

- [ ] **Task 4.3**: Add String() to StartEventOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns "^"

- [ ] **Task 4.4**: Add String() to EventWithFollowupOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns "new<-of" format

- [ ] **Task 4.5**: Add String() to OrderedGroupStartOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns "[" or descriptive string

- [ ] **Task 4.6**: Add String() to OrderedGroupEndOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns "]" or descriptive string

- [ ] **Task 4.7**: Add String() to UnorderedGroupStartOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns "(" or descriptive string

- [ ] **Task 4.8**: Add String() to UnorderedGroupEndOp
  - File: `eventest/internal/marble/op.go`
  - Expected: Returns ")" or descriptive string

- [ ] **Task 4.9**: Add String() to Tick
  - File: `eventest/internal/engine/timeline/timeline.go`
  - Expected: Returns descriptive string of tick contents

### Error Message Improvements
- [ ] **Task 4.10**: Improve parser error messages
  - File: `eventest/internal/marble/parser.go`
  - Action: Add context (position, surrounding text) to error messages

- [ ] **Task 4.11**: Improve semantic validator error messages
  - File: `eventest/internal/marble/semantic.go`
  - Action: Add more descriptive error messages for each rule

- [ ] **Task 4.12**: Improve runtime error messages
  - File: `eventest/internal/engine/runtime/runtime.go`
  - Action: Add context to execution errors

- [ ] **Task 4.13**: Improve interceptor error messages
  - File: `eventest/internal/engine/interceptor/validator.go`
  - Action: Add expected vs actual details to mismatch errors

### Documentation
- [ ] **Task 4.14**: Add package-level doc comment to eventest
  - File: `eventest/doc.go` (NEW)
  - Content: Package overview, usage examples

- [ ] **Task 4.15**: Add doc comments to all exported functions in harness.go
  - File: `eventest/harness.go`
  - Action: Add // Function comments for NewHarness, Run, and all Option functions

- [ ] **Task 4.16**: Add doc comments to marble package
  - File: `eventest/internal/marble/doc.go` (NEW)
  - Content: Package overview, relationship to other packages

- [ ] **Task 4.17**: Add doc comments to engine packages
  - Files: runtime, interceptor, timeline, clock package files
  - Action: Add package-level documentation

---

## Step 5: Documentation and Examples

### Project Documentation
- [ ] **Task 5.1**: Update README.md with usage examples
  - File: `README.md`
  - Content: Basic usage, examples, installation

- [ ] **Task 5.2**: Add architecture documentation
  - File: `docs/architecture.md` (NEW)
  - Content: High-level architecture, component relationships

- [ ] **Task 5.3**: Add getting started guide
  - File: `docs/getting-started.md` (NEW)
  - Content: First steps, basic examples

- [ ] **Task 5.4**: Add advanced usage guide
  - File: `docs/advanced.md` (NEW)
  - Content: Side effects, custom matchers, tick durations

### Code Examples
- [ ] **Task 5.5**: Add Example() functions to harness.go
  - File: `eventest/harness.go`
  - Content: Runnable examples using `go test -run=Example`

- [ ] **Task 5.6**: Add example tests
  - File: `eventest/harness_test.go`
  - Content: ExampleTest functions for documentation

- [ ] **Task 5.7**: Add marble syntax examples to docs/marble.md
  - File: `docs/marble.md`
  - Action: Add more examples, use cases

---

## Code Quality Tasks (Cross-Cutting)

- [ ] **Task CQ.1**: Run linter and fix issues
  - Command: `golangci-lint run` or `go vet ./...`
  - Action: Fix any linting issues

- [ ] **Task CQ.2**: Check for unused imports
  - Command: `go mod tidy`
  - Action: Clean up dependencies

- [ ] **Task CQ.3**: Verify test coverage
  - Command: `go test -cover ./...`
  - Target: > 80% coverage for all packages

- [ ] **Task CQ.4**: Check for race conditions
  - Command: `go test -race ./...`
  - Action: Fix any race conditions found

- [ ] **Task CQ.5**: Verify godoc generation
  - Command: `go doc eventest`
  - Action: Ensure documentation renders correctly

---

## Test Infrastructure Tasks

- [ ] **Task TI.1**: Set up CI/CD pipeline (if not exists)
  - File: `.github/workflows/test.yml` (NEW)
  - Content: Run tests on push/PR

- [ ] **Task TI.2**: Add test coverage reporting
  - File: `.github/workflows/test.yml`
  - Action: Add coverage upload to codecov or similar

- [ ] **Task TI.3**: Add linting to CI
  - File: `.github/workflows/lint.yml` (NEW)
  - Content: Run linter on push/PR

---

## Task Dependencies

```
Step 1 Tasks (1.1-1.9) → No dependencies, can start immediately
    ↓
Step 2 Tasks (2.1-2.19) → Depends on Step 1 completion
    ↓
Step 3 Tasks (3.1-3.7) → Depends on Step 2 completion
    ↓
Step 4 Tasks (4.1-4.17) → Can start after Step 1, but best after Step 2
    ↓
Step 5 Tasks (5.1-5.7) → Can start after Step 3
    ↓
Code Quality Tasks (CQ.1-CQ.5) → Can run at any time, but best at end
    ↓
Test Infrastructure Tasks (TI.1-TI.3) → Can be done in parallel
```

---

## Task Statistics

| Category | Total Tasks | Estimated Time |
|----------|-------------|----------------|
| Step 1: Core Verification | 5 | 1-2 hours |
| Step 2: Harness Tests | 15 | 2-4 hours |
| Step 3: Integration Tests | 7 | 2-3 hours |
| Step 4: Code Quality | 13 | 1-2 hours |
| Step 5: Documentation | 8 | 1-2 hours |
| Code Quality | 5 | 1 hour |
| Test Infrastructure | 3 | 1 hour |
| **Total** | **56** | **8-13 hours** |

---

## Recommended Execution Order

### Phase 1: Foundation (High Priority)
1. Task 1.4: Run all existing tests and fix failures
2. Tasks 1.1-1.3: Verify core functionality
3. Tasks 2.1-2.8: Basic harness tests
4. Tasks 2.9-2.13: Options tests

### Phase 2: Robustness (High Priority)
1. Tasks 2.14-2.19: Error case tests
2. Tasks 3.1-3.7: Integration tests
3. Tasks 4.10-4.13: Error message improvements

### Phase 3: Polish (Medium Priority)
1. Tasks 4.1-4.9: String() methods
2. Tasks 4.14-4.17: Documentation
3. Tasks CQ.1-CQ.5: Code quality checks

### Phase 4: Documentation (Low Priority)
1. Tasks 5.1-5.7: Project documentation
2. Tasks TI.1-TI.3: CI/CD setup

---

## Quick Start

To get started immediately:

```bash
# Run existing tests
cd /home/thomas/Documents/projects/opensource/it-happened
go test ./...

# Fix the gomockevent test issues
vim eventest/gomockevent/matchers_test.go

# Create harness tests
vim eventest/harness_test.go
```

---

## Task Owners

| Task Prefix | Recommended Owner | Notes |
|-------------|-------------------|-------|
| Step 1 | Core Team | Requires deep knowledge of runtime |
| Step 2 | QA/Test Engineer | Focus on test coverage |
| Step 3 | Integration Specialist | End-to-end perspective |
| Step 4 | All | Everyone can contribute |
| Step 5 | Technical Writer | Documentation focus |

---

## Milestones

### Milestone 1: Core Complete
- All Step 1 tasks complete
- All existing tests pass
- Target: End of Week 1

### Milestone 2: Test Coverage Complete
- All Step 2 tasks complete
- All Step 3 tasks complete
- Code coverage > 80%
- Target: End of Week 2

### Milestone 3: Production Ready
- All Step 4 tasks complete
- All Step 5 tasks complete
- All code quality tasks complete
- Target: End of Week 3

---

## Definition of Done

A task is considered **Done** when:
1. The code change is implemented
2. All tests pass (`go test ./...`)
3. No linting issues (`go vet ./...`)
4. Code is reviewed and approved
5. Documentation is updated (if applicable)
6. The change is committed to the repository
