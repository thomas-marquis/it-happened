# Concepts

This section explains the core concepts of the it-happened library. Each concept is described in simple, non-technical language to help you understand the building blocks of event-driven applications.

## Event

An Event represents something that happened in your application. It is the fundamental building block of event-driven architecture. Each event has a unique identifier, a type that categorizes it, and carries data in its payload. Events are immutable once created, ensuring consistency throughout their lifecycle.

## Type

A Type is a label or category for an event. It allows you to classify and identify different kinds of events in your system. Types make it possible to filter events and route them to the appropriate handlers based on their purpose or domain.

## Payload

A Payload is the data carried by an event. It contains the actual information about what happened. Every payload knows its event type, which helps the system understand how to process the event's data.

## Chainable

Chainable is the ability of an event to be part of a sequence or chain. When events are chainable, they can be linked together to represent a series of related occurrences. This allows tracking workflows that span multiple steps across your application.


## Chain

A Chain is a sequence of related events that share a common reference. Think of it as a conversation or workflow where each step produces a new event. Chains allow you to track the progression of a process from start to finish, even when the steps are processed asynchronously.

An event can create a followup event in a way to build a chain. All events within the same chain share the same ChainRef. Furthermore, the ChainPosition tracks its position within the chain.
This makes it possible to build complex, multi-step processes from individual events.

## ChainRef

ChainRef is the unique identifier that links all events in a chain together. Every event in the same chain shares the same ChainRef, which is typically the ID of the first event in that chain. This reference allows the system to correlate related events across time and space.

## ChainPosition

ChainPosition indicates where an event sits within its chain. The first event has position 0, the next event in the same chain has position 1, and so on. This helps you understand the order of events in a multi-step process and track progress through a chain.

## Followup

A Followup is a new event created as a direct result of a previous event in a chain. It shares the same ChainRef as its parent but has an incremented position. Followups allow you to build event sequences where each step naturally leads to the next.

## Bus

The Bus is the central communication hub of the library. It enables different parts of your application to communicate without knowing about each other. Components publish events to the bus, and other components subscribe to receive events they're interested in, creating a decoupled architecture.

## Subscriber

A Subscriber receives and processes events from the bus. It allows you to register callback functions that will be invoked when specific types of events occur.
Subscribers can match events using different criteria to filter only the events they care about.

A subscriber is persisted in the bus until it has been unsubscribed.

## Matcher

A Matcher is a filter that determines which events a subscriber should receive. It examines each event and decides whether it matches the subscriber's criteria. Matchers enable precise event routing and allow subscribers to focus only on relevant events.

## Option

An Option is a way to configure how events and carriers are created. Using the functional options pattern, you can specify settings like context, timeout, or concurrency limits. Options provide flexibility without requiring complex constructors with many parameters.

## Notifier

A Notifier receives notifications whenever an event is published to the bus. It allows you to monitor event activity without subscribing to specific events. The default implementation is a no-operation notifier that discards all notifications.

## Carrier

A Carrier is a special type of event that can dispatch multiple other events to the bus. It acts as an orchestrator, managing a group of events and their lifecycle. Carriers are useful when you need to publish several related events as a single unit.

## CompletionCondition

A CompletionCondition is a rule that determines when an event dispatched by a carrier is considered complete. It's a function that compares the original sent event with received events and returns true when the appropriate completion criteria are met. This allows carriers to track the progress of their dispatched events.