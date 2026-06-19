---
description: "Task list for Subscriber Unsubscribe Capability feature implementation"
---

# Tasks: Subscriber Unsubscribe Capability

**Feature ID**: 003 | **Branch**: `feat/subscriber-unsubscribe` | **Date**: 2026-06-19

**Input**: Design documents from `/specs/003-subscriber-unsubscribe/`

**Prerequisites**: plan.md (required), spec.md (required)

**Tests**: Tests are INCLUDED per project's Test-First Development principle (II. NON-NEGOTIABLE)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No shared infrastructure needed - using existing project structure

- [ ] T001 Verify Go 1.25+ environment is available

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core changes that all user stories depend on

- [ ] T002 [P] Review existing Subscriber implementation in event/subscriber.go
- [ ] T003 [P] Review existing test patterns in event/*_test.go files

---

## Phase 3: User Story 1 - Clean up temporary subscriptions in Sequence carrier [US-001]

**Goal**: Fix memory leaks from dynamic subscriptions in carrier/Sequence by ensuring Detach() clears all callbacks

**Independent Test Criteria**: 
- Detach() clears the registered callbacks map
- No callbacks fire after Detach() is called
- Sequence carrier processes multiple events without memory leaks

### Tests
- [ ] T101 [P] [US-001] Write test: Detach() clears all registered callbacks in event/subscriber_test.go
- [ ] T102 [P] [US-001] Write test: No callbacks invoked after Detach() in event/subscriber_test.go
- [ ] T103 [P] [US-001] Write test: Detach() is idempotent in event/subscriber_test.go
- [ ] T104 [P] [US-001] Write test: Sequence carrier doesn't leak memory with repeated usage in carrier/sequence_test.go

### Implementation
- [ ] T105 [US-001] Implement Detach() to clear registered map in event/subscriber.go

**Dependencies**: T101, T102, T103 (tests must fail before implementation)

**Parallel Opportunities**: T101-T104 can run in parallel (different test scenarios)

---

## Phase 4: User Story 2 - Remove specific event handlers dynamically [US-002]

**Goal**: Enable fine-grained removal of individual callbacks via OnWithCancel() method

**Independent Test Criteria**: 
- OnWithCancel() returns a working cancellation function
- Calling cancel function stops the specific callback
- Other callbacks on the same Subscriber remain active
- Multiple callbacks can be independently cancelled

### Tests
- [ ] T201 [P] [US-002] Write test: OnWithCancel() returns cancel function in event/subscriber_test.go
- [ ] T202 [P] [US-002] Write test: Cancel function removes specific callback in event/subscriber_test.go
- [ ] T203 [P] [US-002] Write test: Multiple OnWithCancel() callbacks are independent in event/subscriber_test.go
- [ ] T204 [P] [US-002] Write test: Cancel function is idempotent in event/subscriber_test.go
- [ ] T205 [P] [US-002] Write test: Concurrent cancellation is thread-safe in event/subscriber_test.go
- [ ] T206 [P] [US-002] Write test: Detach() after cancel works correctly in event/subscriber_test.go
- [ ] T207 [P] [US-002] Write test: Cancel after Detach() is safe (no-op) in event/subscriber_test.go

### Implementation
- [ ] T208 [US-002] Implement OnWithCancel() method in event/subscriber.go

**Dependencies**: T201-T207 (tests must fail before implementation)

**Parallel Opportunities**: T201-T207 can run in parallel (different test scenarios)

---

## Phase 5: Documentation & Polish

**Purpose**: Final touches and cross-cutting concerns

- [ ] T301 [P] Update On() Go doc comment to document persistence until Detach() in event/subscriber.go
- [ ] T302 [P] Add Go doc comment for OnWithCancel() method in event/subscriber.go
- [ ] T303 [P] Add Go doc comment for modified Detach() behavior in event/subscriber.go

---

## Phase 6: Integration & Verification

**Purpose**: Ensure all components work together correctly

- [ ] T401 Run all existing tests to verify no regressions: `go test ./...`
- [ ] T402 Run all new tests to verify functionality: `go test ./event/... -run TestSubscriber`
- [ ] T403 Run all tests with race detector: `go test -race ./...`
- [ ] T404 Run lint script: `./tools/lint.sh`
- [ ] T405 Verify carrier/Sequence tests still pass in carrier/sequence_test.go

**Dependencies**: All Phase 3-5 tasks must be complete

---

## Dependencies Graph

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational) → T002, T003
    ↓
Phase 3 (US-001)
    T101─┬─ T105
    T102─┘
    T103─┘
    T104─┘
    ↓
Phase 4 (US-002)
    T201─┬─ T208
    T202─┘
    T203─┘
    T204─┘
    T205─┘
    T206─┘
    T207─┘
    ↓
Phase 5 (Documentation)
    T301, T302, T303
    ↓
Phase 6 (Integration)
    T401 → T402 → T403 → T404 → T405
```

---

## Parallel Execution Examples

### US-001 (Phase 3) - Parallel Test Writing
```bash
# All US-001 tests can be written in parallel
go test -run TestSubscriber/Detach ./event/...
```

### US-002 (Phase 4) - Parallel Test Writing
```bash
# All US-002 tests can be written in parallel
go test -run TestSubscriber/OnWithCancel ./event/...
```

### Documentation (Phase 5) - Parallel Documentation Updates
```bash
# All documentation tasks are independent
git checkout -b feat/subscriber-unsubscribe
do tasks T301, T302, T303 in any order
```

---

## Implementation Strategy

### MVP Scope (Phase 3 Only)
- Modify `Detach()` to clear callbacks
- Fixes the critical memory leak in carrier/Sequence
- All existing code continues to work
- **Estimated Value**: 80% of the problem solved with 20% of the effort

### Full Implementation (Phases 3-5)
- Adds fine-grained cancellation via `OnWithCancel()`
- Enables advanced use cases
- Complete API with documentation

---

## File Paths Summary

| Phase | File | Tasks |
|-------|------|-------|
| Phase 3 | event/subscriber_test.go | T101-T104 |
| Phase 3 | event/subscriber.go | T105 |
| Phase 4 | event/subscriber_test.go | T201-T207 |
| Phase 4 | event/subscriber.go | T208 |
| Phase 5 | event/subscriber.go | T301-T303 |
| Phase 6 | All packages | T401-T405 |

---

## Format Validation

✅ **All tasks follow the checklist format**:
- Every task starts with `- [ ]`
- Every task has a sequential ID (T001, T002, ...)
- Parallel tasks are marked with `[P]`
- User story tasks are marked with `[US-001]` or `[US-002]`
- Every task includes specific file paths

---

## Task Statistics

- **Total Tasks**: 22
- **Setup Phase**: 1 task
- **Foundational Phase**: 2 tasks
- **US-001 (Phase 3)**: 5 tasks (4 tests + 1 implementation)
- **US-002 (Phase 4)**: 8 tasks (7 tests + 1 implementation)
- **Documentation Phase**: 3 tasks
- **Integration Phase**: 5 tasks

- **Parallel Tasks**: 18 (82% of tasks)
- **Independent Test Criteria**: Defined for each user story

---

## Definition of Done

This feature is complete when:
- [ ] All 22 tasks are completed
- [ ] All tests pass (including race detector)
- [ ] Lint script passes without errors
- [ ] No breaking changes to existing code
- [ ] All Go doc comments are updated
- [ ] Both user stories are independently testable and verified
