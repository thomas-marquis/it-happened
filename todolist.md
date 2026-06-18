# Implementation Todo List: Marble Testing Refactoring

## Overview

This document provides a detailed, actionable implementation checklist for refactoring the marble testing feature according to the TECHNICAL_SPEC.md. All tasks must be completed in the specified order to ensure a clean transition.

---

## Phase 0: Preparation (Do NOT code yet - spec is written)

- [x] Read and understand `specs/constitution.md`
- [x] Read all documentation under `docs/` folder
- [x] Read all source code under `eventest/` folder and subfolders
- [x] Write detailed technical specification in `TECHNICAL_SPEC.md`
- [x] Update spec to use "initEvent" terminology
- [x] Clarify timing: both expectation and side effect start at same tick 0
- [x] Clarify: side effect MUST NOT contain initEvent (`^`)

---

## Phase 1: Core Marble Package - Rename Types

### Task 1.1: Rename PlaceholderNode to InitEventNode

**Goal:** Standardize the node type name throughout the codebase.

- [x] `eventest/internal/marble/node.go`
  - [x] Rename `type PlaceholderNode struct` to `type InitEventNode struct`
  - [x] Update `Accept` method: `v.VisitInitEvent(n)` instead of `v.VisitPlaceholder(n)`
  - [x] Update `Position()` method (no change needed, just rename type)

- [x] `eventest/internal/marble/parser.go`
  - [x] Line ~39: Change `&PlaceholderNode{pos:...}` to `&InitEventNode{pos:...}`

- [x] `eventest/internal/marble/op.go`
  - [x] Rename `PlaceholderEventOpType` to `InitEventOpType`
  - [x] Rename `type PlaceholderEventOp struct{}` to `type InitEventOp struct{}`
  - [x] Update `Type()` method to return `InitEventOpType`
  - [x] Keep `String()` returning `"^"` (marble syntax unchanged)

- [x] `eventest/internal/marble/node_helpers.go`
  - [x] Update `VisitPlaceholder` to `VisitInitEvent` in `opListBuilder`
  - [x] Update `VisitPlaceholder` to `VisitInitEvent` in `stringBuilder`

- [x] `eventest/internal/marble/visitor.go`
  - [x] Update interface: `VisitInitEvent(*InitEventNode)` instead of `VisitPlaceholder(*PlaceholderNode)`
  - [x] Update `BaseVisitor`: `VisitInitEvent(*InitEventNode) {}`

- [x] Update all test files that reference these types:
  - [x] `eventest/internal/marble/parser_test.go`
  - [x] `eventest/internal/marble/node_test.go`
  - [x] `eventest/internal/marble/semantic_test.go`

**Verification:** All references to `PlaceholderNode` and `PlaceholderEventOp` are replaced with `InitEventNode` and `InitEventOp`.

---

## Phase 2: Semantic Rules - Add New Validation

### Task 2.1: Add MandatoryInitEventRule

**Goal:** Ensure expectation chain starts with exactly one initEvent at position 0.

- [x] `eventest/internal/marble/semantic.go`
  - [x] Add new rule type: `MandatoryInitEventRule struct{}`
  - [x] Add helper type: `initEventVisitor struct { BaseVisitor; count int }`
  - [x] Implement `VisitInitEvent(*InitEventNode)` to increment count
  - [x] Implement `VisitSequence` and `VisitGroup` to recursively visit children
  - [x] Implement `isFirstNodeInitEvent(n Node) bool` helper function
  - [x] Implement `Validate(node Node) error` method:
    - Count initEvent occurrences (must be exactly 1)
    - Verify first node is initEvent (directly or nested in group)
    - Return appropriate error messages

### Task 2.2: Add NoInitEventInSideEffectRule

**Goal:** Ensure side effect chain never contains initEvent.

- [x] `eventest/internal/marble/semantic.go`
  - [x] Add new rule type: `NoInitEventInSideEffectRule struct{}`
  - [x] Reuse `initEventVisitor` from MandatoryInitEventRule
  - [x] Implement `Validate(node Node) error` method:
    - Count initEvent occurrences
    - Return error if count > 0

### Task 2.3: Add SideEffectDurationRule

**Goal:** Ensure side effect duration does not exceed expectation duration.

- [x] `eventest/internal/marble/semantic.go`
  - [x] Add new rule type: `SideEffectDurationRule struct { ExpectedDuration int }`
  - [x] Implement `Validate(node Node) error` method:
    - Calculate side effect timeline ticks
    - Compare with expectedDuration
    - Return error if side effect ticks > expectedDuration

### Task 2.4: Remove Old Rules

- [x] `eventest/internal/marble/semantic.go`
  - [x] Remove `StartEventAtBeginningRule`
  - [x] Remove `StartEventAnywhereRule`
  - [x] Remove `UniqueStartEventRule`
  - [x] Remove `placeholderEventVisitor` (replaced by `initEventVisitor`)
  - [x] Remove `isFirstNodeStart` function (replaced by `isFirstNodeInitEvent`)

**Verification:** Run existing tests to ensure new rules work correctly.

---

## Phase 3: Timeline Builder - Handle InitEventNode

### Task 3.1: Update TimelineBuilder for InitEventNode

**Goal:** Ensure timeline builder can process InitEventNode.

- [x] `eventest/internal/engine/timeline/timeline_builder.go`
  - [x] Replace `VisitPlaceholder(n *marble.PlaceholderNode)` with `VisitInitEvent(n *marble.InitEventNode)`
  - [x] Update to use `marble.InitEventOp{}` instead of `marble.PlaceholderEventOp{}`
  - [x] Logic remains the same (create new tick or add to current ops)

**Verification:** Timeline building works for marble strings with `^`.

---

## Phase 4: Runtime - Update Validation

### Task 4.1: Remove Old Start Event Rules from Runtime

- [x] `eventest/internal/engine/runtime/runtime.go`
  - [x] In `RunAllFromNode` or `RunFromNode`:
    - [x] Remove `marble.StartEventAnywhereRule{}` from validation
    - [x] Keep only `marble.WaitlessGroupsRule{}`
    - [x] Add `marble.NoInitEventInSideEffectRule{}` for side effect validation

### Task 4.2: Ensure InitEventOp Handling

- [x] `eventest/internal/engine/runtime/runtime.go`
  - [x] In `Next()` method of `RunningSession`:
    - [x] Add case for `marble.InitEventOp`:
      - [x] Pop placeholder from list
      - [x] Publish to bus
      - [x] Error if no placeholder available (updated error message)

**Verification:** Runtime can execute side effect marble sequences without errors.

---

## Phase 5: Interceptor - Update Validation

### Task 5.1: Add MandatoryInitEventRule to Interceptor

- [x] `eventest/internal/engine/interceptor/interceptor.go`
  - [x] In `FromMarble()` method:
    - [x] Add `marble.MandatoryInitEventRule{}` to validation rules
    - [x] Keep `marble.WaitlessGroupsRule{}`

### Task 5.2: Update Validator for InitEventNode

- [x] `eventest/internal/engine/interceptor/validator.go`
  - [x] Replace `VisitPlaceholder(n *marble.PlaceholderNode)` with `VisitInitEvent(n *marble.InitEventNode)`
  - [x] Update comment to reference initEvent instead of start event/placeholder
  - [x] Call `validateInitEvent(v.currentTick)`
  - [x] Add `validateInitEvent(tickIdx int)` method:
    - [x] Error if tickIdx != 0 (initEvent must be at tick 0)

**Verification:** Interceptor properly validates expectation marble sequences.

---

## Phase 6: Harness - Update API

### Task 6.1: Rename PublishAndWait to RunAndWait

- [x] `eventest/harness.go`
  - [x] Rename method: `func (h *Harness) PublishAndWait(t *testing.T, placeholders ...event.Event)`
  - [x] To: `func (h *Harness) RunAndWait(t *testing.T)`
  - [x] Remove `placeholders ...event.Event` parameter
  - [x] Keep PublishAndWait as deprecated wrapper for backward compatibility

### Task 6.2: Update RunAndWait Implementation

- [x] `eventest/harness.go` - `RunAndWait` method:
  - [x] Create clock and interceptor
  - [x] Parse and validate expectation:
    - [x] Parse with `marble.ParseAsNode(h.expected)`
    - [x] Validate with `marble.MandatoryInitEventRule{}`
  - [x] Create recorder with `intercept.EXPECT().FromMarble(h.expected)`
  - [x] Apply matchers from payloadMap, eventMap, matchers (existing logic)
  - [x] If side effect is provided:
    - [x] Parse side effect: `marble.ParseAsNode(h.sideEffect)`
    - [x] Validate side effect with:
      - [x] `marble.NoInitEventInSideEffectRule{}`
      - [x] `marble.WaitlessGroupsRule{}`
    - [x] Calculate expectation timeline and duration
    - [x] Calculate side effect timeline and duration
    - [x] Validate: side effect duration <= expectation duration
      - [x] If violation, `t.Fatalf("side effect duration (%d ticks) exceeds expectation duration (%d ticks)")`
    - [x] Create runtime with clock, payloadMap, eventMap, tickDuration
    - [x] Run side effect: `rt.RunAll(h.sideEffect)`
    - [x] Note: Side effect starts at tick 0 (by marble definition)
  - [x] Else (no side effect):
    - [x] Start clock
    - [x] Defer clock stop
  - [x] Wait for full expectation duration:
    - [x] Calculate total duration from timeline
    - [x] Sleep for totalDuration
    - [x] Stop clock
  - [x] Call `intercept.Finish()`

### Task 6.3: Update Harness Options (if needed)

- [x] Verify all existing options still work:
  - [x] `WithSideEffect`
  - [x] `WithPayloads`
  - [x] `WithMatchers`
  - [x] `WithEvents`
  - [x] `WithTickDuration`

**Verification:** Harness can be created and RunAndWait executes correctly.

---

## Phase 7: Documentation Updates

### Task 7.1: Update marble.md

- [x] `docs/marble.md`
  - [x] Update terminology: "Start Event" → "initEvent"
  - [x] Update symbol reference: `^` is the initEvent indicator
  - [x] Update examples to use consistent terminology
  - [x] Clarify: initEvent is mandatory and must be at beginning
  - [x] Add section about side effects: must not contain initEvent

### Task 7.2: Update getting-started.md

- [x] `docs/getting-started.md`
  - [x] Update all examples to use new Harness API
  - [x] Replace `PublishAndWait` with `RunAndWait`
  - [x] Remove placeholder event passing
  - [x] Ensure all examples start expectation with `^`
  - [x] Ensure all side effects do NOT contain `^`

### Task 7.3: Update advanced.md

- [x] `docs/advanced.md`
  - [x] Update terminology to "initEvent"
  - [x] Update code examples to use new API

---

## Phase 8: Test Updates

### Task 8.1: Update harness_test.go

- [x] `eventest/harness_test.go`
  - [x] Replace all `PublishAndWait(t, placeholderEvent)` with `RunAndWait(t)`
  - [x] Ensure all expectation strings start with `^`
  - [x] Ensure all side effect strings do NOT contain `^`
  - [x] Update examples:
    - [x] Side effects start with `-` to align with expectation's initEvent at tick 0
    - [x] Verify test intent is preserved

### Task 8.2: Update Marble Package Tests

- [x] `eventest/internal/marble/*_test.go`
  - [x] Update all references to `PlaceholderNode` → `InitEventNode`
  - [x] Update all references to `PlaceholderEventOp` → `InitEventOp`
  - [x] Update visitor implementations
  - [x] Add tests for new semantic rules:
    - [x] `MandatoryInitEventRule`
    - [x] `NoInitEventInSideEffectRule`
    - [x] `SideEffectDurationRule`

### Task 8.3: Update Engine Package Tests

- [x] `eventest/internal/engine/runtime/*_test.go`
  - [x] Update references to old rule names
  - [x] Update test cases to use new validation (removed `^` from side effect test)

- [x] `eventest/internal/engine/interceptor/*_test.go`
  - [x] Update references to PlaceholderNode → InitEventNode
  - [x] Update test cases for new validation (added initEvent tick to all tests)

- [x] `eventest/internal/engine/timeline/*_test.go`
  - [x] Update references to PlaceholderNode → InitEventNode
  - [x] Tests already work with InitEventNode in timeline

---

## Phase 9: Error Message Updates

### Task 9.1: Update All Error Messages

Search entire codebase for error messages containing:
- [x] "placeholder" → replace with "initEvent"
- [x] "start event" → replace with "initEvent"
- [x] "Start Event" → replace with "initEvent"
- [x] "PlaceholderNode" → replace with "InitEventNode" (in error messages only)

Files to check:
- [x] `eventest/internal/marble/semantic.go`
- [x] `eventest/internal/engine/interceptor/validator.go`
- [x] `eventest/harness.go`
- [x] All other files with error returns

---

## Phase 10: Final Verification

### Task 10.1: Run All Tests

- [x] Run `go test ./...` from project root
- [x] Fix any compilation errors
- [x] Fix any test failures
- [x] Ensure all tests pass

### Task 10.2: Manual Testing

- [x] Create simple test cases for each requirement:
  - [x] Expectation without initEvent → should fail
  - [x] Side effect with initEvent → should fail
  - [x] Side effect longer than expectation → should fail
  - [x] Valid case: expectation with initEvent, side effect without → should pass
  - [x] Side effect starting with wait ("-abc") → should pass
  - [x] Side effect in group: expected: "(^abc)", se: "(ab)" → should pass

### Task 10.3: Code Review

- [x] Verify all TODOs are removed or addressed
- [x] Verify no references to old terminology remain (except in migration docs)
- [x] Verify consistent style and formatting
- [x] Verify all imports are correct

---

## Priority Levels

### P0 (Critical - Must Complete)
- Phase 1: Core type renaming
- Phase 2: Semantic rules
- Phase 6: Harness API update
- Phase 10: Verification

### P1 (High - Complete After P0)
- Phase 3: Timeline builder
- Phase 4: Runtime updates
- Phase 5: Interceptor updates
- Phase 8: Test updates

### P2 (Medium)
- Phase 7: Documentation updates
- Phase 9: Error message updates

---

## Estimated Implementation Order

1. **Phase 1** - Rename types (1-2 days)
2. **Phase 2** - Add semantic rules (1 day)
3. **Phase 8.2** - Update marble package tests (1 day)
4. **Phase 3** - Timeline builder (0.5 day)
5. **Phase 4** - Runtime updates (0.5 day)
6. **Phase 5** - Interceptor updates (0.5 day)
7. **Phase 6** - Harness API (1 day)
8. **Phase 8.1** - Update harness tests (1 day)
9. **Phase 8.3** - Update engine tests (0.5 day)
10. **Phase 7** - Documentation (1 day)
11. **Phase 9** - Error messages (0.5 day)
12. **Phase 10** - Final verification (1 day)

**Total Estimated Time: 8-10 days**

---

## Dependencies

- Phase 1 must be complete before Phase 2, 3, 4, 5, 6, 8
- Phase 2 must be complete before Phase 6 and Phase 8.2
- Phase 6 depends on Phase 1, 2, 3, 4, 5
- Phase 7 and 9 can be done in parallel with other phases
- Phase 10 must be last

---

## Checklist for Each Task

For each task above, follow this checklist:

- [ ] Read the current file(s) to understand existing code
- [ ] Make minimal, focused changes
- [ ] Preserve existing functionality
- [ ] Update comments and documentation
- [ ] Compile and test the changes
- [ ] Commit with descriptive message

---

## Migration Notes

### For Users of the Library

Breaking changes to communicate:
1. `PublishAndWait(t, placeholders...)` → `RunAndWait(t)`
2. Expectation marble MUST start with `^` (initEvent)
3. Side effect marble MUST NOT contain `^` (initEvent)
4. No placeholder events need to be passed

### Migration Example

**Before:**
```go
harness := eventest.NewHarness(bus, "^abc",
    eventest.WithSideEffect("-abc"))
harness.PublishAndWait(t, initEvent)
```

**After:**
```go
harness := eventest.NewHarness(bus, "^abc",
    eventest.WithSideEffect("-abc"))
harness.RunAndWait(t)
```

---

## Status Tracking

Use this format to track progress:

```markdown
- [x] Task completed
- [ ] Task pending
- [-] Task in progress
- [~] Task blocked
```

Update this file as you complete each task.

---

## Important Notes

1. **No Breaking Changes to Marble Syntax**: The `^` character remains the initEvent indicator in marble strings. Only the internal type names and API change.

2. **Side Effect Can Start with Wait**: A side effect like "-abc" is valid and starts at tick 0 (the first tick is a wait operation).

3. **initEvent is ONLY in Expectation**: The initEvent (`^`) must be in the expectation chain and must NOT appear in the side effect chain.

4. **Both Start at Tick 0**: By definition of how marble timelines work, both expectation and side effect start executing at tick 0.

5. **Test Duration**: The test runs for the full duration of the expectation chain, regardless of when the side effect completes (as long as it's within the expectation duration).

---

## Files Summary

**Files to modify (25 total):**

### Core Marble (7 files)
- `eventest/internal/marble/node.go`
- `eventest/internal/marble/op.go`
- `eventest/internal/marble/parser.go`
- `eventest/internal/marble/visitor.go`
- `eventest/internal/marble/node_helpers.go`
- `eventest/internal/marble/semantic.go`
- `eventest/internal/marble/parser_test.go`
- `eventest/internal/marble/semantic_test.go`
- `eventest/internal/marble/node_test.go`

### Engine (8 files)
- `eventest/internal/engine/timeline/timeline_builder.go`
- `eventest/internal/engine/timeline/timeline.go`
- `eventest/internal/engine/runtime/runtime.go`
- `eventest/internal/engine/interceptor/interceptor.go`
- `eventest/internal/engine/interceptor/validator.go`
- `eventest/internal/engine/interceptor/interceptor_test.go`
- `eventest/internal/engine/runtime/runtime_test.go`
- `eventest/internal/engine/timeline/timeline_test.go`

### Public API (2 files)
- `eventest/harness.go`
- `eventest/harness_test.go`

### Documentation (3 files)
- `docs/marble.md`
- `docs/getting-started.md`
- `docs/advanced.md`
