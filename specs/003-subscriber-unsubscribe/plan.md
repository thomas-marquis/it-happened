# Implementation Plan: Subscriber Unsubscribe Capability

**Feature ID**: 003 | **Branch**: `feat/subscriber-unsubscribe` | **Date**: 2026-06-19 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-subscriber-unsubscribe/spec.md`

## Summary

This plan implements the ability to unsubscribe event callbacks from a Subscriber, addressing memory leak issues in dynamic subscription scenarios (particularly in carrier/Sequence). The implementation consists of two minimal changes: (1) modifying `Detach()` to clear registered callbacks, and (2) adding a new `OnWithCancel()` method for fine-grained subscription cancellation. The solution maintains full backward compatibility with existing code.

## Technical Context

**Language/Version**: Go 1.25+

**Primary Dependencies**: None (uses existing event types)

**Storage**: N/A (library project, no persistent storage)

**Testing**: go test, testify/assert, testify/require, race detector

**Target Platform**: Any platform supporting Go 1.25+

**Project Type**: library

**Performance Goals**: No degradation to existing event dispatch performance; callback removal O(n) for matcher's callback slice

**Constraints**: 
- No breaking changes to public API
- All existing tests must continue to pass
- Must pass race detector (`-race` flag)
- Follows test-first development (Red-Green-Refactor)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle Compliance Evaluation

| Principle | Requirement | Compliance | Justification |
|-----------|-------------|------------|---------------|
| **I. Event-First Design** | All features designed around event-driven patterns | ✅ PASS | Feature enhances event-driven pattern by enabling proper resource cleanup for subscriptions |
| **II. Test-First Development** (NON-NEGOTIABLE) | Tests MUST be written before implementation | ✅ PASS | All tasks include test-first approach; Red-Green-Refactor cycle explicitly followed |
| **III. Clean Interface Design** | Public APIs defined through clear, minimal interfaces | ✅ PASS | Only 2 minimal changes: Detach behavior + OnWithCancel method; no new interfaces |
| **IV. Type Safety and Contracts** | All types strongly typed | ✅ PASS | Uses existing Matcher and callback types; no type assertions |
| **V. Observability and Traceability** | Events support tracing via ChainRef and ChainPosition | ✅ PASS | No impact on observability; feature is about resource management |
| **VI. Simplicity and Composability** | Components small, focused, composable | ✅ PASS | Minimal, focused changes; composable with existing API |
| **VII. Quality Gates** (NON-NEGOTIABLE) | All contributions pass quality gates | ✅ PASS | Plan includes: code clean & documented, tests pass, linting passes, CI passes |

### Development Workflow Compliance

| Workflow Rule | Compliance | Notes |
|---------------|------------|-------|
| Test-first approach | ✅ PASS | Tests written before implementation for all tasks |
| Mocks generated using mockgen | ✅ PASS | N/A - no new interfaces requiring mocks |
| Mocks stored in mocks/ directory | ✅ PASS | N/A - no new mocks needed |
| Mocks NOT edited manually | ✅ PASS | N/A - no mock changes |
| Code review verifies constitution compliance | ✅ PASS | Plan references specific constitution principles |

### Quality Standards Compliance

| Standard | Requirement | Compliance | Notes |
|----------|-------------|------------|-------|
| Testing | testify/assert, testify/require | ✅ PASS | All tests use testify |
| Testing | t.Run() subtests with Given/When/Then | ✅ PASS | All new tests follow this structure |
| Documentation | Go doc comments for all public APIs | ✅ PASS | New OnWithCancel method will have full Go doc comments |
| Code Style | Standard Go conventions, ./tools/lint.sh | ✅ PASS | All code will pass linting |

**GATE STATUS**: ✅ **ALL GATES PASS** - Proceed to Phase 0 research

## Project Structure

### Documentation (this feature)

```text
specs/003-subscriber-unsubscribe/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── subscriber.md    # API contract for Subscriber changes
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
.
├── event/
│   ├── subscriber.go    # MODIFIED: Detach() behavior + OnWithCancel() method
│   └── subscriber_test.go # MODIFIED: New tests for unsubscribe functionality
├── carrier/
│   └── sequence.go      # MODIFIED: Update to use new API (optional optimization)
└── gen.go               # No changes (no new mocks needed)
```

**Structure Decision**: Minimal changes to existing files. Only `event/subscriber.go` requires implementation changes. Test file `event/subscriber_test.go` will be created/updated. No new packages needed.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. All constitution principles are satisfied by the feature design.

## Implementation Phases

### Phase 0: Research & Analysis (COMPLETED)
- Analyzed current Subscriber implementation
- Identified memory leak in carrier/Sequence
- Designed minimal API solution

### Phase 1: Design & Specification (COMPLETED)
- Created spec.md with functional requirements
- Defined API changes (Detach behavior + OnWithCancel)
- Documented edge cases and constraints

### Phase 2: Task Breakdown

#### Core Implementation Tasks

| ID | Task | Priority | Estimated Complexity | Dependencies |
|----|------|----------|---------------------|--------------|
| T001 | Write tests for Detach() callback cleanup | High | Low | None |
| T002 | Implement Detach() to clear registered map | High | Low | T001 |
| T003 | Write tests for OnWithCancel() | High | Medium | None |
| T004 | Implement OnWithCancel() method | High | Medium | T003 |
| T005 | Update On() documentation | Medium | Low | None |

#### Test Tasks

| ID | Task | Priority | Estimated Complexity | Dependencies |
|----|------|----------|---------------------|--------------|
| T101 | Test Detach() clears all callbacks | High | Low | T002 |
| T102 | Test Detach() is idempotent | Medium | Low | T002 |
| T103 | Test OnWithCancel() returns working cancel function | High | Medium | T004 |
| T104 | Test cancel function removes specific callback | High | Medium | T004 |
| T105 | Test cancel function is idempotent | Medium | Medium | T004 |
| T106 | Test concurrent cancellation safety | High | Medium | T004 |
| T107 | Test Detach() after cancel works correctly | Medium | Medium | T002, T004 |
| T108 | Test cancel after Detach() is safe (no-op) | Medium | Medium | T002, T004 |
| T109 | Test existing On() usage unaffected | High | Low | T002, T004 |

#### Integration Tasks

| ID | Task | Priority | Estimated Complexity | Dependencies |
|----|------|----------|---------------------|--------------|
| T201 | Run all existing tests to verify no regressions | High | Low | T002, T004 |
| T202 | Run tests with -race flag | High | Low | T201 |
| T203 | Run lint.sh to verify code style | Medium | Low | T002, T004 |
| T204 | Update carrier/sequence.go to use new API (optional) | Low | Low | T002, T004 |

## Detailed Task Descriptions

### Core Implementation

#### T001: Write tests for Detach() callback cleanup
**Description**: Write failing tests that verify Detach() clears the registered callbacks map.

**Given/When/Then**:
```
Given: A Subscriber with multiple registered callbacks
When: Detach() is called
Then: The registered map is empty AND no callbacks are invoked for subsequent events
```

**Files Modified**: `event/subscriber_test.go` (new or existing)

**Acceptance Criteria**:
- [ ] Test fails before implementation (Red)
- [ ] Test verifies map is cleared
- [ ] Test verifies callbacks don't fire after Detach()

---

#### T002: Implement Detach() to clear registered map
**Description**: Modify the Detach() method to clear the registered callbacks map.

**Implementation**:
```go
func (s *Subscriber) Detach() {
    s.Lock()
    s.registered = make(map[Matcher][]func(Event))
    s.Unlock()
    close(s.done)
}
```

**Files Modified**: `event/subscriber.go`

**Acceptance Criteria**:
- [ ] All T001 tests pass (Green)
- [ ] Existing Detach() behavior (closing done channel) preserved
- [ ] Thread-safe implementation using existing RWMutex

---

#### T003: Write tests for OnWithCancel()
**Description**: Write failing tests for the new OnWithCancel() method.

**Given/When/Then Scenarios**:
1. ```
   Given: A Subscriber
   When: OnWithCancel() is called with a matcher and callback
   Then: The callback is invoked for matching events AND a cancel function is returned
   ```
2. ```
   Given: A Subscriber with OnWithCancel() callback registered
   When: The cancel function is called
   Then: The callback is NOT invoked for subsequent matching events
   ```
3. ```
   Given: A Subscriber with multiple OnWithCancel() callbacks
   When: One cancel function is called
   Then: Only that specific callback is removed, others remain active
   ```

**Files Modified**: `event/subscriber_test.go`

**Acceptance Criteria**:
- [ ] All tests fail before implementation (Red)
- [ ] Tests cover all specified scenarios

---

#### T004: Implement OnWithCancel() method
**Description**: Implement the new OnWithCancel() method on Subscriber.

**Implementation**:
```go
// OnWithCancel registers a callback for events matching the given matcher
// and returns a function to cancel/unregister that specific callback.
//
// The callback will be invoked when an event matching the matcher is received.
// Unlike On(), this method allows fine-grained removal of individual callbacks
// without detaching the entire subscriber.
//
// Parameters:
//
//	matcher - The matcher that determines which events trigger the callback
//	callback - The function to invoke when a matching event is received
//
// Returns:
//
//	A function that, when called, removes this specific callback
func (s *Subscriber) OnWithCancel(matcher Matcher, callback func(Event)) func() {
    if s.started {
        panic("cannot register callback after listening started")
    }

    s.Lock()
    defer s.Unlock()
    
    if _, exists := s.registered[matcher]; !exists {
        s.registered[matcher] = make([]func(Event), 0)
    }
    s.registered[matcher] = append(s.registered[matcher], callback)

    // Return cancellation function
    return func() {
        s.Lock()
        defer s.Unlock()
        if callbacks, exists := s.registered[matcher]; exists {
            for i, cb := range callbacks {
                if cb == callback {
                    // Remove by swapping with last element and slicing
                    callbacks[i] = callbacks[len(callbacks)-1]
                    s.registered[matcher] = callbacks[:len(callbacks)-1]
                    // Clean up empty matcher entries
                    if len(s.registered[matcher]) == 0 {
                        delete(s.registered, matcher)
                    }
                    break
                }
            }
        }
    }
}
```

**Files Modified**: `event/subscriber.go`

**Acceptance Criteria**:
- [ ] All T003 tests pass (Green)
- [ ] Thread-safe implementation
- [ ] Proper cleanup of empty matcher entries
- [ ] Panics if called after listening started (consistent with On())

---

#### T005: Update On() documentation
**Description**: Add documentation to the existing On() method clarifying that callbacks persist until Detach() and recommending OnWithCancel() for subscriptions requiring individual cleanup.

**Files Modified**: `event/subscriber.go`

**Acceptance Criteria**:
- [ ] On() method has updated Go doc comment
- [ ] Clearly states callbacks persist until Detach()
- [ ] References OnWithCancel() as alternative

### Test Tasks

All test tasks (T101-T109) follow the project's testing standards:
- Use testify/assert and testify/require
- Use t.Run() subtests with Given/When/Then comments
- Cover edge cases: double cancellation, concurrent access, etc.

### Integration Tasks

#### T201: Verify no regressions
**Description**: Run all existing tests to ensure backward compatibility.

**Command**: `go test ./...`

**Acceptance Criteria**:
- [ ] All existing tests pass
- [ ] No failures or panics

---

#### T202: Race detector verification
**Description**: Run all tests with race detector enabled.

**Command**: `go test -race ./...`

**Acceptance Criteria**:
- [ ] No race conditions detected
- [ ] All tests pass

---

#### T203: Lint verification
**Description**: Run the project's lint script.

**Command**: `./tools/lint.sh`

**Acceptance Criteria**:
- [ ] No lint errors
- [ ] Code follows Go conventions

---

#### T204: Update carrier/sequence.go (Optional)
**Description**: Update the sequence carrier to use the new API for cleaner code.

**Note**: This is optional since the current code already calls Detach(), which will now clean up callbacks. However, using OnWithCancel() could make the code more explicit.

**Files Modified**: `carrier/sequence.go`

**Acceptance Criteria**:
- [ ] Sequence carrier tests continue to pass
- [ ] No memory leaks (verified by existing tests)

## Testing Strategy

### Unit Tests
- All new functionality tested in isolation
- Edge cases covered (double cancel, concurrent access, etc.)
- Mock usage: None required (testing Subscriber directly)

### Integration Tests
- Verify existing code continues to work
- Verify new methods work with existing Subscriber infrastructure

### Race Condition Testing
- All tests run with `-race` flag
- Specific tests for concurrent cancellation scenarios

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking existing On() usage | Low | High | Extensive test coverage; no signature changes |
| Race conditions in OnWithCancel | Medium | High | Thread-safe implementation; race detector testing |
| Performance degradation | Low | Medium | O(n) removal for callback slice; no changes to hot paths |
| Memory leaks in new code | Low | Medium | Test with memory profiling; Detach() clears map |

## Rollback Plan

If issues are discovered after implementation:
1. Revert changes to `event/subscriber.go`
2. Revert changes to `event/subscriber_test.go`
3. All changes are isolated to these files, making rollback simple

## Definition of Done

The feature is complete when:
- [ ] All tasks (T001-T203) are completed
- [ ] All tests pass (including race detector)
- [ ] Lint script passes
- [ ] No breaking changes to existing code
- [ ] Documentation is updated
- [ ] Code follows all constitution principles
