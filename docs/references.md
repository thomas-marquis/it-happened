# References

This section provides links to the API documentation for the it-happened library packages.

## Package Documentation

- [event package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/event) - Core event types and interfaces
- [carrier package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/carrier) - Event carrier implementations
- [inmemory package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/inmemory) - In-memory bus implementation

## Additional Resources

- [GitHub Repository](https://github.com/thomas-marquis/it-happened)
- [Contribution Guidelines](../CONTRIBUTE.md)

## Viewing Documentation Locally

You can also view the documentation locally using the `go doc` command:

```bash
# View all documentation for a package
go doc -all github.com/thomas-marquis/it-happened/event

# View documentation for a specific type or function
go doc github.com/thomas-marquis/it-happened/event.Bus
go doc github.com/thomas-marquis/it-happened/event.New
```

Or start a local documentation server:

```bash
godoc -http=:6060
```

Then visit `http://localhost:6060/pkg/github.com/thomas-marquis/it-happened/` in your browser.