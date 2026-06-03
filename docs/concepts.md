# Concepts

## Event
An **Event** is the fundamental unit of information in the system. It represents a specific occurrence or state change. Each event contains:
- **ID**: A unique identifier (UUID) for the specific event instance.
- **Type**: A string that categorizes the event, used for filtering and matching.
- **Payload**: The actual data or state associated with the occurrence.
- **Context**: A Go context used for cancellation and carrying metadata across asynchronous boundaries.
- **Ref**: A reference identifier used to logically link related events together.

## Bus
The **Bus** is the central communication hub of the library. It follows the publish-subscribe pattern, allowing decoupled components to interact without direct dependencies.
- **Publish**: Broadcasts an event to all interested subscribers.
- **Subscribe**: Creates a subscription that filters events using **Matchers**.

## Followup events
A **Followup event** is an event that is logically connected to a previous event. They are typically used for:
- Request-response patterns.
- Multi-step workflows.
- Signaling the result of an action.

Technically, a followup event is created with the same **Ref** as its predecessor. This shared reference allows subscribers to track the "conversation" or "chain" of events even as they are processed asynchronously.

## Events carrier
An **Events carrier** is a specialized event payload that acts as an orchestrator for multiple events. Instead of representing a single occurrence, a carrier manages a group of events and monitors their lifecycle.

The library provides two primary carrier types:
- **All**: Dispatches a set of events in parallel (respecting concurrency limits) and waits for all of them to be completed.
- **Sequence**: Dispatches events one by one, waiting for each event to be resolved before proceeding to the next one.

### Outcome Factory
The **Outcome Factory** (referred to as `DoneEventFactory` in the code) is a critical concept for carriers. It is a function provided to the carrier that determines what "summary" event should be published once the carrier's mission is complete. It receives all the followup events collected during the process and produces a single final event that represents the cumulative result of the carrier's execution.


