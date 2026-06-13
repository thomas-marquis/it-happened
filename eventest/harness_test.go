package eventest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
	runtimepkg "github.com/thomas-marquis/it-happened/eventest/internal/engine/runtime"
	"github.com/thomas-marquis/it-happened/inmemory"
)

type testPayload string

func (testPayload) Type() event.Type {
	return "test"
}

// Custom payload types for testing with real event types
type customPayloadA struct {
	ID   string
	Data string
}

func (p customPayloadA) Type() event.Type {
	return "custom"
}

type customPayloadB struct {
	ID   string
	Data string
}

func (p customPayloadB) Type() event.Type {
	return "custom"
}

type payloadType1 struct{ TypeID string }

func (payloadType1) Type() event.Type { return "type1" }

type payloadType2 struct{ TypeID string }

func (payloadType2) Type() event.Type { return "type2" }

// Helper to create default payloads using runtime.DefaultPayload
func dp(name string) event.Payload {
	return runtimepkg.DefaultPayload(name)
}

func TestHarness_Publish_SimpleEventSequence(t *testing.T) {
	t.Run("should pass when events match in order", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)
		bus := inmemory.NewBus(done)

		tt := &testing.T{}

		// When
		eventest.NewHarness(bus, "^abc",
			eventest.WithSideEffect("-abc")).
			PublishAndWait(tt, event.New(dp("init event")))

		// Then
		assert.False(t, tt.Failed())
	})

	t.Run("should not pass when events don't match", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)
		bus := inmemory.NewBus(done)

		tt := &testing.T{}

		// When
		eventest.NewHarness(bus, "^abc",
			eventest.WithSideEffect("abc")).
			PublishAndWait(tt, event.New(dp("init event")))

		// Then
		assert.True(t, tt.Failed())
	})

	//t.Run("should pass with single event", func(t *testing.T) {
	//	// Given
	//	done := make(chan struct{})
	//	defer close(done)
	//	bus := inmemory.NewBus(done)
	//
	//	// When & Then
	//	eventest.NewHarness(bus, "^abc",
	//		eventest.WithSideEffect("-abc")).
	//		Publish(t, event.New(dp("init event")))
	//
	//
	//	// Given
	//	bus := inmemory.NewBus(nil, nil)
	//	harness := eventest.NewHarness(bus, "a")
	//
	//	// When & Then - test passes
	//	harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
	//		testBus.Publish(event.New(dp("a")))
	//	})
	//})
}

//func TestHarness_EventSequenceWithWaits(t *testing.T) {
//	t.Run("should handle wait ticks correctly", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// "a-b-c" means: event a (tick 0), wait (tick 1), event b (tick 2), wait (tick 3), event c (tick 4)
//		harness := eventest.NewHarness(bus, "a-b-c")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 0
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 1 (wait)
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 2
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 3 (wait)
//			testBus.Publish(event.New(dp("c")))
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 4
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle multiple consecutive waits", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// "a--b" means: event a (tick 0), wait (tick 1), wait (tick 2), event b (tick 3)
//		harness := eventest.NewHarness(bus, "a--b")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 0
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 1 (wait)
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 2 (wait)
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 3
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle underscore wait syntax", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// "a___b" means: event a (tick 0), wait (tick 1), event b (tick 2)
//		// Note: ___ is parsed as a single WaitNode, same as -
//		harness := eventest.NewHarness(bus, "a___b")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 0
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 1 (wait)
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 2
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_OrderedGroups(t *testing.T) {
//	t.Run("should require events in order within ordered group", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// [ab] means ordered group: a must come before b in the same tick
//		harness := eventest.NewHarness(bus, "[ab]")
//
//		// When - both events in same tick, in order
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			testBus.Publish(event.New(dp("b")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle ordered group at start", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// [ab]c means: ordered group [ab] in first tick, then c in second tick
//		harness := eventest.NewHarness(bus, "[ab]c")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("c")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle ordered group at end", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// a[bc] means: a in first tick, then ordered group [bc] in second tick
//		harness := eventest.NewHarness(bus, "a[bc]")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_UnorderedGroups(t *testing.T) {
//	t.Run("should accept events in any order within unordered group", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// (ab) means unordered group: a and b can come in any order in the same tick
//		harness := eventest.NewHarness(bus, "(ab)")
//
//		// When - publish b before a (both in same tick)
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("a")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should accept events in order within unordered group", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		harness := eventest.NewHarness(bus, "(ab)")
//
//		// When - publish a before b (both in same tick)
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			testBus.Publish(event.New(dp("b")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle unordered group with more events", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		harness := eventest.NewHarness(bus, "(abc)")
//
//		// When - publish in different order (all in same tick)
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//			testBus.Publish(event.New(dp("a")))
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_NestedGroups(t *testing.T) {
//	t.Run("should handle nested ordered groups", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// [a(bc)] means: ordered group containing a, then unordered group (bc)
//		// All in the same tick
//		harness := eventest.NewHarness(bus, "[a(bc)]")
//
//		// When - all events in same tick, in order
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle nested unordered groups", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// (a[bc]) means: unordered group containing a and ordered group [bc]
//		// All in the same tick
//		harness := eventest.NewHarness(bus, "(a[bc])")
//
//		// When - all events in same tick, any order
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle deeply nested groups", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// [(a(bc))] means: ordered group containing unordered group with a and unordered group (bc)
//		// All in the same tick
//		harness := eventest.NewHarness(bus, "[(a(bc))]")
//
//		// When - all events in same tick
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle complex nested structure", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// a-[bc]-(de) means:
//		// tick 0: a
//		// tick 1: wait
//		// tick 2: ordered group [bc]
//		// tick 3: wait
//		// tick 4: unordered group (de)
//		harness := eventest.NewHarness(bus, "a-[bc]-(de)")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration) // end tick 0
//			clk.Forward(timeline.DefaultTickDuration) // end tick 1 (wait)
//			// Ordered group [bc] in tick 2
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//			clk.Forward(timeline.DefaultTickDuration) // end tick 2
//			clk.Forward(timeline.DefaultTickDuration) // end tick 3 (wait)
//			// Unordered group (de) in tick 4
//			testBus.Publish(event.New(dp("d")))
//			testBus.Publish(event.New(dp("e")))
//			clk.Forward(timeline.DefaultTickDuration) // end tick 4
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_FollowupEvents(t *testing.T) {
//	t.Run("should verify followup relationship", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// a b<-a means: event a in tick 0, then event b as followup of a in tick 1
//		harness := eventest.NewHarness(bus, "a b<-a")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			// First publish the original event in tick 0
//			original := event.New(dp("a"))
//			testBus.Publish(original)
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 0
//			// Then publish the followup in tick 1
//			followup := event.NewFollowup(original, dp("b"))
//			testBus.Publish(followup)
//			clk.Forward(timeline.DefaultTickDuration) // end of tick 1
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle followup in sequence", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// a c b<-c d means: a, then c, then b is followup of c, then d
//		harness := eventest.NewHarness(bus, "a c b<-c d")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration)
//			original := event.New(dp("c"))
//			testBus.Publish(original)
//			clk.Forward(timeline.DefaultTickDuration)
//			followup := event.NewFollowup(original, dp("b"))
//			testBus.Publish(followup)
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("d")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_StartEvent(t *testing.T) {
//	t.Run("should handle start event at beginning", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// ^abc means: start event, then a, b, c
//		harness := eventest.NewHarness(bus, "^abc")
//
//		// When - start event in first tick, then a, b, c in subsequent ticks
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("^")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("c")))
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should work without explicit start event", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// abc without ^ should still work
//		harness := eventest.NewHarness(bus, "abc")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("c")))
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_WithPayloads(t *testing.T) {
//	t.Run("should match events by payload", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		payloadA1 := dp("specific-a")
//		payloadB1 := dp("specific-b")
//
//		harness := eventest.NewHarness(
//			bus, "ab",
//			eventest.WithPayloads(map[string]event.Payload{
//				"a": payloadA1,
//				"b": payloadB1,
//			}),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(payloadA1))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(payloadB1))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_WithEvents(t *testing.T) {
//	t.Run("should match exact events", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		eventA := event.New(dp("a"))
//		eventB := event.New(dp("b"))
//
//		harness := eventest.NewHarness(
//			bus, "ab",
//			eventest.WithEvents(map[string]event.Event{
//				"a": eventA,
//				"b": eventB,
//			}),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(eventA)
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(eventB)
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_WithMatchers(t *testing.T) {
//	t.Run("should use custom matchers", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		// Create custom matchers that match any DefaultPayload
//		// This matches the runtime.DefaultPayloadType
//		harness := eventest.NewHarness(
//			bus, "ab",
//			eventest.WithMatchers(map[string]event.Matcher{
//				"a": event.Is(runtimepkg.DefaultPayloadType),
//				"b": event.Is(runtimepkg.DefaultPayloadType),
//			}),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("any-a")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("any-b")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should work with event.Is matcher", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		harness := eventest.NewHarness(
//			bus, "ab",
//			eventest.WithMatchers(map[string]event.Matcher{
//				"a": event.Is("test"),
//				"b": event.Is("test"),
//			}),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(testPayload("a")))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(testPayload("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_WithSideEffect(t *testing.T) {
//	t.Run("should execute side effect before expected sequence", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// Side effect: "x" is executed first, then expected sequence "a"
//		// The interceptor will see both x and a
//		// The side effect runtime publishes x using the clock
//		harness := eventest.NewHarness(
//			bus, "xa",
//			eventest.WithSideEffect("x"),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			// Side effect already published x, now publish a in next tick
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle side effect with multiple events", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		harness := eventest.NewHarness(
//			bus, "abc",
//			eventest.WithSideEffect("ab"),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			// Side effect already published a and b in ticks 0 and 1
//			// Now publish c in tick 2
//			clk.Forward(timeline.DefaultTickDuration)
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("c")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle side effect with groups", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// Side effect "(ab)" then expected "c" -> total sequence "(ab)c"
//		harness := eventest.NewHarness(
//			bus, "(ab)c",
//			eventest.WithSideEffect("(ab)"),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			// Side effect already published a and b (in any order) in tick 0
//			// Now publish c in tick 1
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(dp("c")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_WithTickDuration(t *testing.T) {
//	t.Run("should use custom tick duration", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		customDuration := 100 * time.Millisecond
//
//		// Note: WithTickDuration affects side effects, but for the main sequence
//		// we use the clock's Forward method
//		harness := eventest.NewHarness(
//			bus, "a-b",
//			eventest.WithTickDuration(customDuration),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			// Use clock.Forward with default tick duration for main sequence
//			// The custom tick duration is used by runtime for side effects
//			clk.Forward(timeline.DefaultTickDuration * 2) // skip wait tick
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle very short tick duration", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		shortDuration := 1 * time.Millisecond
//
//		harness := eventest.NewHarness(
//			bus, "a-b",
//			eventest.WithTickDuration(shortDuration),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration * 2) // skip wait tick
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_ErrorMissingEvent(t *testing.T) {
//	t.Run("should fail when expected event is missing", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.14)
//		// The harness reports errors via t.Error(), not panic
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//
//	t.Run("should fail when first event is missing", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.14)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//}
//
//func TestHarness_ErrorExtraEvent(t *testing.T) {
//	t.Run("should fail when extra event is published", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.15)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//}
//
//func TestHarness_ErrorWrongOrder(t *testing.T) {
//	t.Run("should fail when events come in wrong order", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.16)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//
//	t.Run("should fail when ordered group has wrong order", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.16)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//}
//
//func TestHarness_ErrorWrongEventInGroup(t *testing.T) {
//	t.Run("should fail when unordered group has wrong event", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.17)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//
//	t.Run("should fail when ordered group has wrong event", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.17)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//}
//
//func TestHarness_ErrorEmptyMarble(t *testing.T) {
//	t.Run("should handle empty marble string gracefully", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.18)
//		// This actually panics in the parser, so it would panic
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//}
//
//func TestHarness_ErrorInvalidSyntax(t *testing.T) {
//	t.Run("should fail with invalid syntax - unknown character", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.19)
//		// This panics in NewHarness, so assert.Panics would work but we're skipping for now
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//
//	t.Run("should fail with unclosed group", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.19)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//
//	t.Run("should fail with unclosed unordered group", func(t *testing.T) {
//		// TODO: Error case tests need mock testing.T (Task 2.19)
//		t.Skip("Error case tests require verification via mock testing.T")
//	})
//}
//
//func TestHarness_CompleteWorkflow(t *testing.T) {
//	t.Run("should work with complete workflow", func(t *testing.T) {
//		// Given
//		done := make(chan struct{})
//		defer close(done)
//
//		bus := inmemory.NewBus(done, nil)
//		// a-[bc]-(de) means: a, wait, [bc], wait, (de)
//		// tick 0: a
//		// tick 1: wait
//		// tick 2: [bc] (ordered group)
//		// tick 3: wait
//		// tick 4: (de) (unordered group)
//		harness := eventest.NewHarness(
//			bus, "a-[bc]-(de)",
//			eventest.WithTickDuration(10*time.Millisecond),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration) // end tick 0
//			clk.Forward(timeline.DefaultTickDuration) // end tick 1 (wait)
//			// Ordered group [bc] in tick 2
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//			clk.Forward(timeline.DefaultTickDuration) // end tick 2
//			clk.Forward(timeline.DefaultTickDuration) // end tick 3 (wait)
//			// Unordered group (de) in tick 4
//			testBus.Publish(event.New(dp("d")))
//			testBus.Publish(event.New(dp("e")))
//			clk.Forward(timeline.DefaultTickDuration) // end tick 4
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_MultipleHarnessesSameBus(t *testing.T) {
//	t.Run("should allow multiple harnesses on same bus", func(t *testing.T) {
//		// Given
//		done := make(chan struct{})
//		defer close(done)
//
//		bus := inmemory.NewBus(done, nil)
//
//		harness1 := eventest.NewHarness(bus, "a")
//		harness2 := eventest.NewHarness(bus, "b")
//
//		// When
//		harness1.Run(t, func(testBus event.Bus, _ clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//		})
//
//		harness2.Run(t, func(testBus event.Bus, _ clock.Clock) {
//			testBus.Publish(event.New(dp("b")))
//		})
//
//		// Then - both tests pass independently
//	})
//}
//
//func TestHarness_ClockSynchronization(t *testing.T) {
//	t.Run("should verify clock advances correctly through ticks", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//		// Simple sequence with known timing
//		harness := eventest.NewHarness(bus, "a-b-c")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			// Publish a, then advance through ticks
//			testBus.Publish(event.New(dp("a")))
//			clk.Forward(timeline.DefaultTickDuration)
//			clk.Forward(timeline.DefaultTickDuration) // wait tick
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//			clk.Forward(timeline.DefaultTickDuration) // wait tick
//			testBus.Publish(event.New(dp("c")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes (clock advanced correctly)
//	})
//}
//
//func TestHarness_WithRealEventTypes(t *testing.T) {
//	t.Run("should work with custom event payload types", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		// Create matchers for custom payloads
//		customA := customPayloadA{ID: "a", Data: "test-a"}
//		customB := customPayloadB{ID: "b", Data: "test-b"}
//
//		harness := eventest.NewHarness(
//			bus, "ab",
//			eventest.WithMatchers(map[string]event.Matcher{
//				"a": event.HasPayload(customA),
//				"b": event.HasPayload(customB),
//			}),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(customA))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(customB))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should work with different payload types in same sequence", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		payload1 := payloadType1{TypeID: "t1"}
//		payload2 := payloadType2{TypeID: "t2"}
//
//		// Use type matchers instead of specific payloads
//		harness := eventest.NewHarness(
//			bus, "ab",
//			eventest.WithMatchers(map[string]event.Matcher{
//				"a": event.Is("type1"),
//				"b": event.Is("type2"),
//			}),
//		)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(payload1))
//			clk.Forward(timeline.DefaultTickDuration)
//			testBus.Publish(event.New(payload2))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_LongMarbleSequence(t *testing.T) {
//	t.Run("should handle 50+ events in sequence", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		// Build a marble string with 50 events
//		var sb strings.Builder
//		for i := 0; i < 50; i++ {
//			sb.WriteRune(rune('a' + (i % 26)))
//		}
//		longMarble := sb.String()
//
//		harness := eventest.NewHarness(bus, longMarble)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			for i := 0; i < 50; i++ {
//				eventName := string(rune('a' + (i % 26)))
//				testBus.Publish(event.New(dp(eventName)))
//				clk.Forward(timeline.DefaultTickDuration)
//			}
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle sequence with groups and waits", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		// Build a long marble with mixed features: events, waits, groups
//		// Pattern: [abc]-[abc]-[abc]... (repeating abc pattern)
//		var sb strings.Builder
//		for i := 0; i < 10; i++ {
//			if i > 0 {
//				sb.WriteRune('-')
//			}
//			sb.WriteRune('[')
//			sb.WriteRune('a')
//			sb.WriteRune('b')
//			sb.WriteRune('c')
//			sb.WriteRune(']')
//		}
//		longMarble := sb.String()
//
//		harness := eventest.NewHarness(bus, longMarble)
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			for i := 0; i < 10; i++ {
//				if i > 0 {
//					clk.Forward(timeline.DefaultTickDuration) // wait
//				}
//				// Ordered group [abc]
//				testBus.Publish(event.New(dp("a")))
//				testBus.Publish(event.New(dp("b")))
//				testBus.Publish(event.New(dp("c")))
//				clk.Forward(timeline.DefaultTickDuration)
//			}
//		})
//
//		// Then - test passes
//	})
//}
//
//func TestHarness_MixedFeatures(t *testing.T) {
//	t.Run("should handle complex mixed features: start, events, waits, groups, followups", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		harness := eventest.NewHarness(bus, "^a-(bc)[de]")
//
//		// When
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			// Start event
//			testBus.Publish(event.New(dp("^")))
//			clk.Forward(timeline.DefaultTickDuration)
//
//			// Event a
//			originalA := event.New(dp("a"))
//			testBus.Publish(originalA)
//			clk.Forward(timeline.DefaultTickDuration)
//
//			// Wait
//			clk.Forward(timeline.DefaultTickDuration)
//
//			// Unordered group (bc) - can be in any order
//			testBus.Publish(event.New(dp("c")))
//			testBus.Publish(event.New(dp("b")))
//			clk.Forward(timeline.DefaultTickDuration)
//
//			// Ordered group [de] - must be in order
//			testBus.Publish(event.New(dp("d")))
//			testBus.Publish(event.New(dp("e")))
//			clk.Forward(timeline.DefaultTickDuration)
//		})
//
//		// Then - test passes
//	})
//
//	t.Run("should handle deeply nested groups with events and waits", func(t *testing.T) {
//		// Given
//		bus := inmemory.NewBus(nil, nil)
//
//		harness := eventest.NewHarness(bus, "[a(b[c]d)e]")
//
//		// When - all in same tick, in order
//		harness.Run(t, func(testBus event.Bus, clk clock.Clock) {
//			testBus.Publish(event.New(dp("a")))
//			testBus.Publish(event.New(dp("b")))
//			testBus.Publish(event.New(dp("c")))
//			testBus.Publish(event.New(dp("d")))
//			testBus.Publish(event.New(dp("e")))
//		})
//
//		// Then - test passes
//	})
//}
