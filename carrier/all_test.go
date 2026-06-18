package carrier_test

import (
	"testing"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
	"github.com/thomas-marquis/it-happened/inmemory"
)

type fakePayload string

func (fakePayload) Type() event.Type {
	return "fakePayload"
}

func TestAll(t *testing.T) {
	t.Run("should emit all carried event", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		bus := inmemory.NewBus(done)

		a := event.New(fakePayload("aaa"))
		b := event.New(fakePayload("bbb"))
		c := event.New(fakePayload("ccc"))
		doneEvt := event.New(fakePayload("done"))

		th := eventest.NewHarness(bus, "^(abc)/aDone<-a-[(/bDone<-b /cDone<-c)/done]",
			eventest.WithSideEffect("-(abc)/aDone<-a-[(/cDone<-c/bDone<-b)/done]"),
			eventest.WithEvents(map[string]event.Event{
				"a":    a,
				"b":    b,
				"c":    c,
				"done": doneEvt,
			}))

		// When
		th.RunAndWait(t)
	})
}
