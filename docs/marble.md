# Marble Language Specification

## Overview

Marble is a domain-specific language for describing event sequences and timelines in a declarative, human-readable format. It is used by the `eventest` package to define expected event sequences for testing event-driven systems.

## Language Syntax

### Basic Characters

| Character | Name | Description |
|-----------|------|-------------|
| `^` | Start Event | Marks the beginning of a timeline (optional) |
| `-` | Wait | Represents a single time tick with no events |
| `_` | Multi-Wait | One or more consecutive underscores represent a single wait tick (treated same as `-`) |
| `a-z`, `A-Z` | Event Labels | Single character identifiers for events |
| `/` | Named Event Prefix | Allows multi-character event names (e.g., `/eventName`) |
| `(...)` | Unordered Group | Events inside parentheses can occur in any order |
| `[...]` | Ordered Group | Events inside square brackets must occur in the specified order |
| `<-` | Followup Operator | Creates a followup event relationship (e.g., `a<-b` means event `a` is a followup of event `b`) |

### Grammar

```
Marble       ::= (StartEvent? | Event | Wait | Group | EventWithFollowup | Whitespace)*
StartEvent   ::= '^'
Event        ::= Label
Label        ::= ('/' [a-zA-Z0-9]+) | [a-zA-Z]
Wait         ::= '-' | '_'+
Group        ::= OrderedGroup | UnorderedGroup
OrderedGroup ::= '[' Marble ']'
UnorderedGroup ::= '(' Marble ')'
EventWithFollowup ::= Label '<-' Label
Whitespace   ::= ' ' | '\t' | '\n' | '\r'
```

### Operators

#### Start Event (`^`)
- Must appear at the beginning of the marble string if present
- Represents the initialization of the timeline
- Only one start event is allowed per timeline
- Example: `^abc`

#### Event (`a`, `/eventName`)
- Single character labels: `a`, `b`, `c`, etc.
- Multi-character labels must be prefixed with `/`: `/start`, `/complete`, `/error`
- Each event label represents a distinct event in the timeline
- Example: `abc` or `/userCreated/deleteUser`

#### Wait (`-`, `_`)
- Single `-` represents one time tick with no events
- Multiple `_` characters are treated as a single wait (e.g., `____` = one wait tick)
- Used to create time gaps between events
- Example: `a-b` (event a, wait, event b)

#### Event with Followup (`a<-b`)
- Syntax: `<eventName><-<fromEvent>`
- Creates a followup event where `eventName` is a followup of `fromEvent`
- The `fromEvent` must have been published previously in the timeline
- Example: `a<-b` means event `a` is a followup of event `b`

#### Ordered Group (`[...]`)
- Events inside square brackets must occur in the exact order specified
- All events in the group occur within a single time tick
- Can be nested
- Example: `[ab]` means events a and b must occur in order within one tick

#### Unordered Group (`(...)`)
- Events inside parentheses can occur in any order
- All events in the group occur within a single time tick
- The runtime may shuffle the order of events for testing
- Can be nested
- Example: `(ab)` means events a and b can occur in any order within one tick

## Examples

### Simple Sequence
```
abc
```
Three events (a, b, c) each in their own time tick.

### With Waits
```
a-b-c
```
Event a, wait one tick, event b, wait one tick, event c.

### Groups
```
[ab]cd
```
Events a and b in order within one tick, then c, then d.

```
(ab)cd
```
Events a and b in any order within one tick, then c, then d.

### Nested Groups
```
[(ab)c]
```
Group containing an unordered group (a and b in any order) followed by c, all within one tick.

```
[ a (bc) d ]
```
Ordered group with nested unordered group: a, then b and c in any order, then d, all in one tick.

### Followup Events
```
a<-b
```
Event a is a followup of event b.

```
a b<-a c
```
Event a, then event b which is a followup of a, then event c.

### Complete Timeline
```
^a-(bc)[de]f
```
Start event, event a, wait, unordered group (b and c in any order), ordered group (d then e), event f.

## Semantic Rules

The following rules are enforced during validation:

1. **Waitless Groups**: Wait operators (`-`, `_`) cannot be used inside groups. Groups represent a single time tick, so all events must occur simultaneously.

2. **Start Event Rules** (configurable):
   - `StartEventAtBeginningRule`: If a start event is present, it must be at the beginning of the timeline
   - `StartEventAnywhereRule`: At most one start event can exist anywhere in the timeline
   - `UniqueStartEventRule`: Exactly one start event must be present

3. **Non-Empty Timeline**: A timeline must have at least one operation.

## Time Model

- Each operator (event, wait, group) represents one time tick
- The default tick duration is configurable (default: 10ms)
- Groups (ordered or unordered) represent a single time tick containing multiple events
- Events within a group are executed in the same time tick

## Integration with Event System

### Event Resolution
When a marble sequence is executed:

1. Each event label is resolved to an actual `event.Event`
2. If a payload mapping is provided, the event uses that payload
3. If an event mapping is provided, the event is used directly
4. Otherwise, a default payload is created from the label

### Followup Events
- Followup events use the `event.NewFollowup(from, to)` constructor
- The `From` label must reference a previously defined event
- The `EventName` label is the new event being created as a followup

## Use Cases

Marble is primarily used for:

1. **Testing Event Sequences**: Define expected event sequences in tests
2. **Timeline Simulation**: Simulate event-driven systems over time
3. **Intercepting Events**: Verify that actual events match expected marble sequences

## Comparison with Other Notations

Marble is inspired by:
- **RxJS Marble Diagrams**: Similar visual notation for observable sequences
- **Cucumber Given-When-Then**: Declarative test descriptions
- **Timing Diagrams**: Visual representation of events over time

Unlike RxJS marble diagrams which are visual, this marble language is text-based and designed for programmatic use in tests.
