package main

import (
	"context"
	"fmt"
	"time"

	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

// PipelinePayload is used for all pipeline events.
type PipelinePayload struct {
	Message string
}

func (p PipelinePayload) EventType() event.Type {
	return "pipeline.demo"
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := inmemory.NewBus(ctx)

	// Subscribe to pipeline events
	sub := bus.Subscribe()

	sub.On(event.Is("pipeline.demo"), func(e event.Event) {
		payload, ok := e.Payload().(PipelinePayload)
		if ok {
			fmt.Printf("  %s\n", payload.Message)
		} else {
			return
		}
		// Emit followup so pipeline can continue to next stage
		// Only emit for original events (not followups) to avoid infinite loop
		if e.ChainPosition() == 0 {
			bus.Publish(e.NewFollowup(PipelinePayload{Message: payload.Message + " ->"}))
		}
	})

	sub.ListenWithWorkers(1)

	fmt.Println("=== Pipeline Carrier (Function-Based Transformation) ===")
	fmt.Println("Publishing pipeline carrier with 3 stages...")

	// Create the initial event for the pipeline
	initEvent := event.New(PipelinePayload{Message: "Stage 0: Initial data"})

	// Define the pipeline stages as transformation functions
	pipelineStages := []func(prev event.Event) event.Event{
		// Stage 1: Transform initial data
		func(prev event.Event) event.Event {
			if p, ok := prev.Payload().(PipelinePayload); ok {
				return event.New(PipelinePayload{
					Message: p.Message + " Stage 1 processed",
				})
			}
			return prev
		},
		// Stage 2: Further processing
		func(prev event.Event) event.Event {
			if p, ok := prev.Payload().(PipelinePayload); ok {
				return event.New(PipelinePayload{
					Message: p.Message + " Stage 2 processed",
				})
			}
			return prev
		},
		// Stage 3: Final processing
		func(prev event.Event) event.Event {
			if p, ok := prev.Payload().(PipelinePayload); ok {
				return event.New(PipelinePayload{
					Message: p.Message + " Stage 3 finalized",
				})
			}
			return prev
		},
	}

	pipelineCarrier := carrier.NewPipeline(
		initEvent,
		pipelineStages,
		event.New(PipelinePayload{Message: "Pipeline timed out"}),
		carrier.WithTimeout(2*time.Second),
	)

	bus.Publish(pipelineCarrier)

	// Wait for pipeline processing
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("Note: Pipeline carrier executes a sequence of transformation functions.")
}
