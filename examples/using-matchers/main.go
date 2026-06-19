package main

import (
	"fmt"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

// UserCreatedPayload represents a user creation event.
type UserCreatedPayload struct {
	UserID   string
	Username string
	Email    string
}

func (p UserCreatedPayload) EventType() event.Type {
	return "user.created"
}

// UserUpdatedPayload represents a user update event.
type UserUpdatedPayload struct {
	UserID   string
	Username string
	Email    string
}

func (p UserUpdatedPayload) EventType() event.Type {
	return "user.updated"
}

// UserDeletedPayload represents a user deletion event.
type UserDeletedPayload struct {
	UserID string
}

func (p UserDeletedPayload) EventType() event.Type {
	return "user.deleted"
}

// SystemEventPayload represents a system event.
type SystemEventPayload struct {
	Type    string
	Message string
}

func (p SystemEventPayload) EventType() event.Type {
	return event.Type(p.Type)
}

func main() {
	done := make(chan struct{})
	defer close(done)

	bus := inmemory.NewBus(done, nil)

	// Create a subscriber
	sub := bus.Subscribe()

	// Example 1: Match by specific event type
	sub.On(event.Is("user.created"), func(e event.Event) {
		if payload, ok := e.Payload().(UserCreatedPayload); ok {
			fmt.Printf("[Is] User created: %s (%s)\n", payload.Username, payload.Email)
		}
	})

	// Example 2: Match any of multiple types using IsOneOf
	sub.On(event.IsOneOf("user.updated", "user.deleted"), func(e event.Event) {
		switch payload := e.Payload().(type) {
		case UserUpdatedPayload:
			fmt.Printf("[IsOneOf] User updated: %s (%s)\n", payload.Username, payload.Email)
		case UserDeletedPayload:
			fmt.Printf("[IsOneOf] User deleted: %s\n", payload.UserID)
		}
	})

	// Example 3: Match all events using IsAny
	sub.On(event.IsAny(), func(e event.Event) {
		fmt.Printf("[IsAny] Event received: %s\n", e.Type())
	})

	// Start listening with 1 worker
	sub.ListenWithWorkers(1)

	// Publish various events
	bus.Publish(event.New(UserCreatedPayload{
		UserID:   "user-001",
		Username: "alice",
		Email:    "alice@example.com",
	}))

	bus.Publish(event.New(UserUpdatedPayload{
		UserID:   "user-001",
		Username: "alice.smith",
		Email:    "alice.smith@example.com",
	}))

	bus.Publish(event.New(UserDeletedPayload{
		UserID: "user-002",
	}))

	bus.Publish(event.New(SystemEventPayload{
		Type:    "system.health_check",
		Message: "All systems operational",
	}))

	// Wait for all events to be processed
	time.Sleep(200 * time.Millisecond)

	fmt.Println("Matcher demonstration complete!")
}
