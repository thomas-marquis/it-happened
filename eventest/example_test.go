package eventest_test

import (
	"fmt"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
)

func ExamplePayloadEq() {
	res1 := eventest.PayloadEq(fakePayload("foo")).Matches(event.New(fakePayload("foo")))
	res2 := eventest.PayloadEq(fakePayload("bar")).Matches(event.New(fakePayload("baz")))

	fmt.Println(res1, res2)
	// Output:
	// true false
}

func ExampleIsFollowupOf() {
	a := event.New(fakePayload("test"))
	b := a.NewFollowup(fakePayload("followup"))
	c := event.New(fakePayload("other"))

	res1 := eventest.IsFollowupOf(a).Matches(b)
	res2 := eventest.IsFollowupOf(a).Matches(c)

	fmt.Println(res1, res2)
	// Output:
	// true false
}
