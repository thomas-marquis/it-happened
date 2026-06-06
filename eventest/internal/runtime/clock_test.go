package runtime_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/runtime"
)

func TestVirtualClock(t *testing.T) {
	t.Run("should have current time at start", func(t *testing.T) {
		// Given
		vc := runtime.NewVirtualClock()

		// When
		vc.Start()

		// Then
		assert.NotZero(t, vc.Now())
		assert.LessOrEqual(t, time.Now().Sub(vc.Now()), 10*time.Millisecond)
	})

	t.Run("should forward time", func(t *testing.T) {
		// Given
		vc := runtime.NewVirtualClock()
		vc.Start()
		start := vc.Now()

		// When
		vc.Forward(1 * time.Second)
		vc.Forward(2 * time.Second)
		vc.Forward(1 * time.Second)

		// Then
		assert.Equal(t, start.Add(4*time.Second), vc.Now())
	})

	t.Run("should return elapsed time", func(t *testing.T) {
		// Given
		vc := runtime.NewVirtualClock()
		vc.Start()

		// When
		vc.Forward(1 * time.Second)
		vc.Forward(2 * time.Second)

		// Then
		assert.Equal(t, 3*time.Second, vc.Elapsed())
	})

	t.Run("should execute scheduled events once", func(t *testing.T) {
		// Given
		clock := runtime.NewVirtualClock()
		timer := clock.(runtime.Timer)
		scheduler := clock.(runtime.Scheduler)
		count := 0
		scheduler.Schedule(1*time.Second, func() {
			count++
		})
		timer.Start()

		// When
		clock.Forward(1 * time.Second)
		assert.Equal(t, 1, count, "should have run once after first forward")

		clock.Forward(1 * time.Second)

		// Then
		assert.Equal(t, 1, count, "should NOT have run again after second forward")
	})

	t.Run("should execute events in order", func(t *testing.T) {
		// Given
		vc := runtime.NewVirtualClock()
		results := make([]int, 0)
		vc.Schedule(2*time.Second, func() { results = append(results, 2) })
		vc.Schedule(1*time.Second, func() { results = append(results, 1) })
		vc.Start()

		// When
		vc.Forward(3 * time.Second)

		// Then
		assert.Equal(t, []int{1, 2}, results)
	})

	t.Run("should reset on stop", func(t *testing.T) {
		// Given
		vc := runtime.NewVirtualClock()
		vc.Start()
		vc.Forward(1 * time.Second)

		// When
		vc.Stop()

		// Then
		assert.Zero(t, vc.Elapsed())
	})

	t.Run("should execute event when reached by multiple forwards", func(t *testing.T) {
		// Given
		vc := runtime.NewVirtualClock()
		count := 0
		vc.Schedule(1*time.Second, func() {
			count++
		})
		vc.Schedule(3*time.Second, func() {
			count++
		})
		vc.Start()

		// When & Then
		vc.Forward(500 * time.Millisecond)
		assert.Equal(t, 0, count)
		vc.Forward(500 * time.Millisecond)
		assert.Equal(t, 1, count)
		vc.Forward(2 * time.Second)
		assert.Equal(t, 2, count)
	})
}
