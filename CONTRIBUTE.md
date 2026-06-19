# Contributing to it-happened

Thank you for your interest in contributing to `it-happened`! This document will help you get started with local development and explain our standards.

## 🛠️ Technical Requirements

To work on this project, you need:

- **Go**: 1.25 or higher
- **Mockgen**: For generating mocks (`go install go.uber.org/mock/mockgen@latest`)
- **UV**: For running the documentation locally (optional, if you want to preview docs)

## 💻 Local Setup

### 1. Clone the repository

```bash
git clone https://github.com/thomas-marquis/it-happened.git
cd it-happened
```

### 2. Install dependencies

```bash
go get .
```

### 3. Generate Mocks

We use `gomock` for testing. Mocks are generated based on the interfaces defined in the project. To generate them, run:

```bash
go generate ./...
```

This will update the files in the `mocks/` directory.

### 4. Run Tests

To ensure everything is working correctly, run the test suite:

```bash
go test ./...
```

## 📁 Project Structure

The project follows a standard Go structure:

- `TODO/`: Core package containing the client implementation and public API.
- `internal/`: Private packages used by the library.
- `mocks/`: Generated mocks for unit testing.
- `examples/`: Runnable examples demonstrating various features.
- `docs/`: Markdown files for the project documentation.
- `tools/`: Development utilities and scripts. This folder contains tools used for development, such as the linting utility.
- `specs/`: Documentation and specifications for coding agent
- `Makefile`: Shortcut commands for documentation management.
- `gen.go`: Configuration for mock generation.

## 🧪 Testing Strategy

This project follows a comprehensive testing approach to ensure reliability and maintainability.

### Testing Framework

- **Assertions**: Use [testify/assert](https://github.com/stretchr/testify) and [testify/require](https://github.com/stretchr/testify) for all assertions. Prefer `require` for conditions that must be true to continue the test.
- **Mocking**: Use `go.uber.org/mock/mockgen` (gomock) for generating mock implementations of interfaces. Mocks are generated using `go generate ./...` and stored in the `internal/mocks/` directory.

### Test Structure

All test files should follow these conventions:

- **Naming**: Test functions should be named using the pattern `Test<StructName>_<MethodName>` for methods or `Test<FunctionName>` for package-level functions.
- **Organization**: Each test function should contain a single subtest using `t.Run()` with a descriptive name.
- **Structure**: Use Given/When/Then comments to clearly separate test phases:
  ```go
  t.Run("Given <setup>, When <action>, Then <expected result>", func(t *testing.T) {
      // Given: Setup test conditions
      
      // When: Execute the operation under test
      
      // Then: Verify the results
  })
  ```
- **Subtests**: For complex scenarios, nest additional `t.Run()` calls within the main test function to test multiple cases.

### Mocking Strategy

**When to mock**:
- Mock external dependencies (databases, APIs, etc.)
- Mock interfaces to isolate the code under test
- Mock when the real implementation is slow or non-deterministic

**When NOT to mock**:
- Avoid mocking simple data structures
- Avoid mocking code within the same package (test through public APIs)
- Avoid over-mocking which makes tests brittle

**Mock generation**:
```bash
# Generate mocks for all interfaces
go generate ./...

# Generate mocks for a specific package
go generate ./event/...
```

Mocks are automatically generated from interface definitions. Do not manually edit files in `internal/mocks/`.

### Testing Concurrent Operations

For testing concurrent code, follow these best practices:

1. **Use the race detector**: Always run tests with `-race` flag to detect data races:
   ```bash
   go test -race ./...
   ```

2. **Synchronization primitives**: Use `sync.WaitGroup` to wait for goroutines to complete:
   ```go
   var wg sync.WaitGroup
   wg.Add(1)
   go func() {
       defer wg.Done()
       // concurrent operation
   }()
   wg.Wait()
   ```

3. **Channels**: Use buffered channels to prevent goroutine leaks:
   ```go
   ch := make(chan Event, 100) // Buffered to avoid blocking
   ```

4. **Timeouts**: Always set timeouts for async operations to prevent deadlocks:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
   defer cancel()
   ```

### Testing Event Buses

For testing inmemory bus implementations:

- Test **publish/subscribe**: Verify events are delivered to registered subscribers
- Test **multiple subscribers**: Verify all subscribers receive published events
- Test **event matching**: Verify only subscribers with matching criteria receive events
- Test **concurrent operations**: Verify thread-safety with simultaneous publishes and subscribes
- Test **order preservation**: For sequential carriers, verify event order is maintained

Example test structure:
```go
t.Run("Given bus with registered subscriber, When event is published, Then subscriber receives the event", func(t *testing.T) {
    // Given
    bus := inmemory.NewBus()
    received := make(chan event.Event, 1)
    
    sub := bus.Subscribe().
        On(event.Is("test.payload"), func(evt event.Event) {
            received <- evt
        })
    sub.ListenWithWorkers(1)
    defer sub.Detach()
    
    // When
    bus.Publish(event.New(testPayload("test")))
    
    // Then
    select {
    case evt := <-received:
        assert.Equal(t, "test.payload", evt.Type())
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for event")
    }
})
```

### Testing Event Carriers

For testing All and Sequence carriers:

**All carrier** (parallel dispatch):
- Test that events are dispatched concurrently
- Test that order is NOT preserved
- Test timeout scenarios with `WithTimeout` option
- Test followup event completion with `CompletionCondition`

**Sequence carrier** (sequential dispatch):
- Test that events are dispatched in order
- Test that next event waits for previous to complete
- Test timeout scenarios
- Test followup event completion

Example carrier test:
```go
t.Run("Given All carrier with multiple events, When Dispatch is called, Then all events are published", func(t *testing.T) {
    // Given
    events := []event.ChainableEvent{
        event.New(testPayload("event1")),
        event.New(testPayload("event2")),
    }
    
    carrierEvent := carrier.NewAll(
        events,
        func(received []event.Event) event.Event { 
            return event.New(testPayload("done")) 
        },
        event.New(testPayload("timeout")),
        carrier.WithMaxConcurrency(2),
    )
    
    // When/Then - verify construction
    assert.NotNil(t, carrierEvent)
    _, ok := carrierEvent.Payload().(carrier.Carrier)
    assert.True(t, ok)
})
```

### Computing Coverage Locally

To compute test coverage for the project:

```bash
# Run all tests with coverage
 go test ./... -cover

# Generate coverage profile
 go test ./... -coverprofile=coverage.out

# View coverage by function
 go tool cover -func=coverage.out

# View coverage by file
 go tool cover -func=coverage.out | grep "event/"

# Generate HTML report and open in browser
 go tool cover -html=coverage.out -o coverage.html
```

The project aims for a minimum of 80% code coverage for all core packages (event, carrier, inmemory).

### Linting

Before submitting a PR, make sure your code passes the linting check:

```bash
./tools/lint.sh
```

## ✅ Definition of Done

A contribution is considered complete when:

1.  **Implementation**: The code is clean, documented, and follows the existing style.
2.  **Tests**: Unit tests are added for new features or bug fixes. All tests must pass.
3.  **Lint**: The `./tools/lint.sh` script passes without errors.
4.  **Documentation**: Relevant documentation in the `docs/` folder is updated or added.
5.  **CI**: All GitHub Action workflows pass.
6.  **Example**: If you added a new feature, a corresponding example should be added in the `examples/` folder.
7.  **README**: If the change is significant, update the `README.md` to reflect it.
8.  **Guidelines**: If needed, update the `specs/constitution.md` (and any other AI guideline files) file to reflect the new changes.
