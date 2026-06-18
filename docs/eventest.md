# eventest

## Concepts

Testing an event-oriented code is not as easy as a regular code.
Usually, in a unit test, we mock external dependencies, making assertions on them and/or stubing their behaviors by defined ones.

For example, with `gomock`, we can write:

```go


func TestUserService(t *testing.T) {
    // Given
    ctrl := gomock.NewController(t)
    mockRepo := userRepo.NewUserRepositoryMock()

    user := user.User{ID: "fakeUserID", Email: "toto@lolo.com"}

    mockRepo.EXPECT().
        GetByID(gomock.Eq("fakeUserID")).
        Return(user, nil).
        Times(1)

    mockRepo.EXPECT().
        Save(gomock.Cond(func (u any) bool {
            usr := u.(user.User)
            return assert.Equal(t, "newemail@lolo.com", usr.Email)
        }).
        Return(nil).
        Times(1)

    svc := service.New(mockRepo)

    // When
    err := svc.UpdateEmail("fakeUserID", "newemail@lolo.com")

    // Then
    assert.NoError(t, err)
}
```


## Introducing marble testing

**Timeline**

The timeline is the sequence that drives the test.
The timeline duration IS the test duration.

The simplest possible timeline is: `"-"`

A dash means a waiting tick.

```go
timeline := "---" // the test will end after 3 ticks of time
```

The timeline represents the `Return` method of `gomock`. We can use it to simulate things that happens from the outside world.

For example:
```go
timeline := "--a" // a is an event
```

Here, we wait for 2 time ticks, them an event named `a` is emitted.

An event can be written as a simple letter or with a complex label starting with a slash.
For example:
- `c`
- `/emailUpdated`

Both are valid label.

By default, an event occupies exactly one time tick. So, in the bellow example, the timeline lasts 4 ticks: 2 waiting and 2 emitting events.
```go
timeline := "--ab"

// equivalent to
timeline = "--/updatedAsked/emailUpdated"

// also equivalent to
timeline = "--a/emailUpdated"

// or
timeline = "--/updatedAsked b"
```

Please note all the white characters are ignored. Thus, they can be used to separate 2 events or to make the sequence easier to read.

### Making asserting on events

A timeline set the test duration and allows us to simulate events.
Now, we need a way to make assertions.

Our code will publish event to the event bus, in reaction to the initial event and/or to the mocked event from the timeline.
We need to write an assertion sequence to specify what the events are expected to be published from our code.

Here's an example:

```go
// Given
bus := inmemory.New(nil)

timeline := "----"
expected := "^--/emailUpdated"
th := eventest.NewHarness(t, bus,
    timeline, expected,
)

svc := service.New(bus)

initEvt := event.New(user.UpdateAsked{Field: "email", Value: "newemail@lolo.com", UserID: "fakeUserID"})

// When & Then
th.RunAndWait(initEvt)
```

Please note:
- the epxected equence MUST be sorter or equal to the timeline. Remember the test is driven by the timeline.
- the expected seq MUST start with the initial event `^`. This symbol asserts the bus has received the initial event. It must be placed on the begining of first time tick
