

## Marble syntax

**General syntax**

* `-`: a single time tick
* `_`: one or multiple consecutive underscores count as a single time tick
* `[...]`: a group of ordered events emitted on the same time tick
* `(...)`: a group of unordered events emitted on the same time tick
* `a`: a single event as a single-character label (can include all letters: a-zA-Z)
* `/myEvent`: a single event as a multiple-character label (must start with a letter: a-zA-Z, can contain numbers but nothing else)
* `^`: represent the first published event on the bus. Count as one time tick (can be grouped with other events). Each timeline MUST start with it
* ` `: all white/blank characters are ignored (spaces, tabs, new line), even in the middle of the sequence
* `a<-b` or `/myevent<-/previousEvent`: expect `a` to be a `b`'s followup event


```go
type fakePayload string

func (p fakePayload) Type() event.Type {
    return "fake"
}


// carrier all
se := "  ^       - d<-a - (f<-ce<-b)   "
exp := "[^(abc)] - d<-a - [(e<-b f<-c)x]"

// carrier sequence
se := "   ^   /aRes<-a   -- /bRes<-b    /cRes<-c"
exp := " [^a][/aRes<-a b]--[/bRes<-b c][/cRes<-cx]" // with strict match
exp2 := " ^   b          -- c           x"          // with non-strict match

bus := inmemory.NewBus()
th := eventest.NewHarness(bus
    exp,
    WithSideEffect(),
    WithEvents(map[string]event.Payload{ // define a payload implementation for labels. A fake payload impl is used by default for labels that don't apear in this map
        "a": ...,
        "b": ...,
    }),
    WithMatchers(map[string]event.Matcher{ //
        "": ...,
        "": ...,
    }),
)

in := event.New(fakePayload("input value"))

th.Run(func(bus event.Bus) {
    bus.Publish(in)
})

```
