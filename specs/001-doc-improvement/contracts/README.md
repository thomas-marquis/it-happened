# Documentation Contracts

This directory contains the interface contracts for the Documentation Improvement feature.
These contracts define the expected structure, format, and quality standards for all documentation artifacts.

## Documentation Structure Contract

The documentation site MUST have the following structure:

```
docs/
├── index.md             # Quick Start / Landing page
├── concepts.md          # Core concepts explanations
├── tutorials/           # Practical tutorials directory
│   ├── basic-pubsub.md
│   ├── event-chaining.md
│   ├── using-matchers.md
│   └── using-carriers.md
└── references.md         # API references
```

**Validation Rules**:
- All files MUST be valid Markdown
- All internal links MUST resolve to existing files
- Navigation in mkdocs.yml MUST include all four sections

## Concept Documentation Contract

Each concept in `concepts.md` MUST follow this format:

```markdown
### [Concept Name]

[Short description: 3-4 sentences maximum]

[Optional: Brief code example]
```

**Required Concepts**:
- Event
- Type
- Payload
- Chainable
- ChainableEvent
- Chain
- ChainRef
- ChainPosition
- Followup
- Bus
- Subscriber
- Matcher
- Option
- Notifier
- Carrier
- CompletionCondition

**Validation Rules**:
- Description MUST be ≤ 4 sentences
- MUST explain "what it is" and "why it matters"
- MUST NOT explain implementation details
- MUST use simple, non-technical language

## Tutorial Contract

Each tutorial MUST follow this structure:

```markdown
# [Tutorial Title]

[One-sentence description of what the tutorial covers]

## Prerequisites

[List of prerequisites, if any]

## What You'll Learn

[Bullet list of learning objectives]

## Step-by-Step Guide

[Numbered steps with code examples]

## Complete Example

[Link to corresponding example in examples/ folder]

```bash
cd examples/[tutorial-name]
go run .
```

## Next Steps

[Suggestions for what to learn next]
```

**Required Tutorials**:
1. Basic Publish/Subscribe
2. Event Chaining
3. Using Matchers
4. Using Carriers

**Validation Rules**:
- MUST have a corresponding example in examples/ folder
- Example MUST be runnable with `go run .`
- MUST demonstrate one of the most important use cases
- Code examples MUST be self-contained

## Example Contract

Each example in `examples/` folder MUST:

1. Be in its own subdirectory named after the tutorial
2. Contain a `main.go` file
3. Be self-contained (no external dependencies beyond the library itself)
4. Be runnable with `go run .`
5. Demonstrate exactly one tutorial concept
6. Include minimal, well-commented code

**Required Examples**:
- examples/basic-pubsub/main.go
- examples/event-chaining/main.go
- examples/using-matchers/main.go
- examples/using-carriers/main.go

## Doc Comment Contract

All Go doc comments MUST follow these standards:

### For Types

```go
// [TypeName] [brief description starting with verb].
//
// [Additional details if needed.]
// [More details.]
type [TypeName] struct {
    ...
}
```

### For Interfaces

```go
// [InterfaceName] [brief description of its role].
//
// [Explanation of what implementers should do.]
type [InterfaceName] interface {
    ...
}
```

### For Functions

```go
// [FunctionName] [brief description of what it does].
//
// Parameters:
//   [param1] - [description]
//   [param2] - [description]
//
// Returns:
//   [return1] - [description]
func [FunctionName]([params]) ([returns]) {
    ...
}
```

### For Methods

```go
// [MethodName] [brief description of what it does].
//
// Parameters:
//   [param1] - [description]
//
// Returns:
//   [return1] - [description]
func ([receiver]) [MethodName]([params]) ([returns]) {
    ...
}
```

**Validation Rules**:
- MUST start with the symbol name
- MUST describe purpose, not implementation
- MUST use complete sentences
- For functions/methods: MUST explain parameters and return values
- MUST be placed immediately before the symbol declaration

## References Contract

The `references.md` file MUST contain:

```markdown
# References

## Package Documentation

- [event package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/event)
- [carrier package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/carrier)
- [inmemory package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/inmemory)

## Additional Resources

- [GitHub Repository](https://github.com/thomas-marquis/it-happened)
- [Contribution Guidelines](../CONTRIBUTE.md)
```

**Validation Rules**:
- All URLs MUST be valid and accessible
- Package URLs MUST point to pkg.go.dev
- MUST include all main packages (event, carrier, inmemory)

## Quality Gates Contract

All documentation MUST pass these quality checks:

1. **Markdown Validation**: All Markdown files MUST be syntactically valid
2. **Link Validation**: All internal links MUST resolve to existing files
3. **Spell Check**: All documentation MUST have correct spelling (English)
4. **Consistency**: All terminology MUST be consistent across all files
5. **Completeness**: All required concepts, tutorials, and examples MUST be present
6. **Build Test**: `mkdocs build` MUST complete without errors
7. **Doc Coverage**: `go doc` on every exported symbol MUST show documentation

**Validation Commands**:
```bash
# Check Markdown syntax
markdownlint docs/**/*.md

# Check links
mkdocs build --strict

# Check Go doc coverage
go doc -all github.com/thomas-marquis/it-happened/event
go doc -all github.com/thomas-marquis/it-happened/carrier
go doc -all github.com/thomas-marquis/it-happened/inmemory
```
