package main_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
	"github.com/thomas-marquis/it-happened/inmemory"
)

type customer struct {
	Name string
	Age  int
}

const (
	customerRegisteredEventType       = "customer.registered"
	customerRegistrationSucceededType = "customer.registration.succeeded"
)

type customerRegistrationRequested struct {
	Name string
	Age  int
}

func (customerRegistrationRequested) Type() event.Type {
	return customerRegisteredEventType
}

type customerRegistrationSucceeded struct {
	Customer customer
}

func (customerRegistrationSucceeded) Type() event.Type {
	return customerRegistrationSucceededType
}

type customerService struct {
	bus event.Bus

	registered []customer
}

func newCustomerService(bus event.Bus) *customerService {
	s := &customerService{bus: bus, registered: make([]customer, 0)}

	bus.Subscribe().On(event.Is(customerRegistrationSucceededType), func(evt event.Event) {
		pl := evt.Payload.(customerRegistrationSucceeded)
		s.registered = append(s.registered, pl.Customer)
	}).ListenWithWorkers(1)

	return s
}

func (s *customerService) List() []customer {
	return s.registered
}

func TestWithMarbleTesting(t *testing.T) {
	t.Run("should add a new customer", func(t *testing.T) {
		// Given
		c := customer{Name: "John", Age: 30}

		done := make(chan struct{})
		defer close(done)
		bus := inmemory.NewBus(done)

		svc := newCustomerService(bus)
		th := eventest.NewHarness(bus, "^---b",
			eventest.WithSideEffect("a---b"),
			eventest.WithPayloads(map[string]event.Payload{
				"a": customerRegistrationRequested{Name: "John", Age: 30},
				"b": customerRegistrationSucceeded{Customer: c},
			}),
			//eventest.WithMatchers(map[string]event.Matcher{
			//	"a": event.HasPayload(customerRegistrationRequested{Name: "John", Age: 30}),
			//}),
		)

		// When & Then
		th.RunAndWait()
		// Give time for async processing
		time.Sleep(100 * time.Millisecond)
		assert.Len(t, svc.List(), 1)
	})

}
