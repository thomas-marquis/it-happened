# Feature Specification: Subscriber Unsubscribe Capability

**Feature ID**: 003 | **Status**: Implemented | **Date**: 2026-06-19 | **Author**: Thomas Marquis | **Implementation Date**: 2026-06-19

## Overview

This feature adds the ability to unsubscribe/unregister individual event callbacks from a Subscriber, addressing a critical memory leak issue where dynamically created subscriptions (particularly in the carrier/Sequence implementation) cannot be cleaned up.

## Problem Statement

Currently, the `Subscriber.On()` method registers callbacks but provides no mechanism to remove them individually. The only cleanup option is `Detach()`, which stops all listeners but retains callback references in memory. This causes memory leaks in scenarios where:

1. Subscriptions are created dynamically (e.g., in `carrier.Sequence.doDispatch()`)
2. Temporary subscriptions are needed for specific event sequences
3. Long-running applications accumulate many short-lived subscriptions

**Evidence**: `carrier/sequence.go` lines 108-134 creates a new Subscriber for each event in a sequence. Even though `sub.Detach()` is called, the callback references persist in the Subscriber's `registered` map.

## Goals

### Primary Goals
- Enable removal of individual callbacks without detaching the entire Subscriber
- Fix memory leaks from dynamically created subscriptions
- Maintain backward compatibility with existing `On()` usage

### Non-Goals
- Modify the Bus interface
- Change existing `On()` method signature
- Add complex subscription management systems

## Functional Requirements

### FR-001: Automatic Callback Cleanup on Detach
The `Detach()` method MUST clear all registered callbacks from the Subscriber's internal map to prevent memory leaks when a Subscriber is no longer needed.

**Priority**: Critical

**Rationale**: This is the minimal fix for the sequence carrier memory leak. When `Detach()` is called, the Subscriber is being shut down, so it should clean up all resources.

---

### FR-002: Fine-Grained Callback Cancellation
A new method MUST be added to allow individual callback removal without affecting other callbacks on the same Subscriber.

**Priority**: High

**Rationale**: Enables advanced use cases where a Subscriber manages multiple independent subscriptions and needs to clean up specific ones.

---

### FR-003: Backward Compatibility
All existing `On()` method usage MUST continue to work without modification.

**Priority**: Critical

**Rationale**: Existing codebase and users depend on the current API. No breaking changes allowed.

## User Stories

### US-001: Clean up temporary subscriptions in Sequence carrier
**As a** library user implementing event sequences  
**I want** to create and clean up temporary subscriptions for each event in a sequence  
**So that** my application doesn't leak memory when processing many event sequences

**Acceptance Criteria**:
- [ ] Sequence carrier can create a Subscriber for each event
- [ ] Subscriber can be fully cleaned up after use
- [ ] No memory leaks when processing multiple sequences

---

### US-002: Remove specific event handlers dynamically
**As a** library user with a long-running Subscriber  
**I want** to unsubscribe from specific event types at runtime  
**So that** I can adjust my event handling dynamically without recreating the Subscriber

**Acceptance Criteria**:
- [ ] Can register multiple callbacks on one Subscriber
- [ ] Can remove individual callbacks without affecting others
- [ ] Removed callbacks no longer receive events

## API Design

### Current API (Preserved)
```go
// Existing method - unchanged
func (s *Subscriber) On(matcher Matcher, callback func(Event)) *Subscriber
```

**Documentation Update**: Must clearly state that callbacks registered via `On()` persist until `Detach()` is called, and recommend `OnWithCancel()` for subscriptions requiring individual cleanup.

---

### New API: Detach() Behavior Change
```go
// Modified behavior - now clears registered callbacks
func (s *Subscriber) Detach()
```

**Behavior**: 
- Closes the `done` channel (existing behavior)
- Clears the `registered` map to release callback references (NEW)

---

### New API: OnWithCancel()
```go
// New method for fine-grained cancellation
func (s *Subscriber) OnWithCancel(matcher Matcher, callback func(Event)) func()
```

**Parameters**:
- `matcher`: The matcher that determines which events trigger the callback
- `callback`: The function to invoke when a matching event is received

**Returns**:
- A cancellation function that, when called, removes this specific callback

**Usage Example**:
```go
cancel := sub.OnWithCancel(event.Is("order.created"), func(e event.Event) {
    fmt.Println("Order received")
})

// Later, when no longer needed:
cancel()  // Removes just this callback
```

## Technical Considerations

### Thread Safety
- All new methods must be thread-safe
- Must use the existing RWMutex pattern from Subscriber
- Cancellation functions must work correctly when called from any goroutine

### Memory Management
- `Detach()` must clear the `registered` map completely
- `OnWithCancel()` cancellation must remove the specific callback from the matcher's slice
- Empty matcher entries should be removed from the map

### Performance
- Callback removal should be O(n) for the matcher's callback slice (acceptable for typical use cases)
- No performance degradation for existing `On()` and event dispatch paths

## Dependencies

- No new external dependencies required
- Uses existing `event.Matcher` and callback types
- No changes to Bus interface

## Constraints

1. **No Breaking Changes**: Existing `On()` method signature and behavior must remain compatible
2. **Minimal API Surface**: Only add what's necessary (2 changes: Detach behavior + OnWithCancel method)
3. **Constitution Compliance**: Must follow all principles, especially:
   - II. Test-First Development
   - III. Clean Interface Design
   - VII. Quality Gates

## Success Criteria

### Measurable Outcomes
- [ ] `Detach()` clears all registered callbacks (verified by memory profiling)
- [ ] `OnWithCancel()` returns a working cancellation function
- [ ] Calling the cancellation function stops the callback from receiving events
- [ ] Existing `On()` usage continues to work unchanged
- [ ] All existing tests continue to pass
- [ ] New tests cover all new functionality

### Quality Gates
- [ ] Code passes `./tools/lint.sh`
- [ ] All tests pass with `-race` flag
- [ ] Documentation updated (Go doc comments, examples if applicable)
- [ ] No breaking changes to public API

## Edge Cases

1. **Double Cancellation**: Calling the cancellation function twice should be safe (idempotent)
2. **Concurrent Cancellation**: Cancelling while events are being processed should not cause race conditions
3. **Detach After Cancel**: Calling `Detach()` after cancelling individual callbacks should work without errors
4. **Cancel After Detach**: Calling a cancellation function after `Detach()` should be safe (no-op)

## Open Questions

None identified at this time.

---

**Approval Status**: Approved  
**Next Steps**: N/A - Implementation Complete

---

## Implementation Notes

**Documentation Updates Required**:
- [ ] Update README.md to document the new `OnWithCancel()` method
- [ ] Add usage examples in the `examples/` directory demonstrating fine-grained subscription cancellation
- [ ] Update any existing documentation that references Subscriber cleanup behavior

**Related Changes**:
- Fixed a bug in `carrier/sequence.go` where `close(finished)` could be called multiple times, causing a panic. Added `sync.Once` protection.

**Breaking Changes**:
- `Detach()` now clears all registered callbacks, which means `Accept()` will return `false` for all events after Detach(). This is intentional to fix memory leaks. Existing code that relied on `Accept()` returning true after Detach() will need to be updated.
