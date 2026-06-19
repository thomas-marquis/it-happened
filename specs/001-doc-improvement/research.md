# Research: Documentation Improvement

**Date**: 2026-06-19
**Feature**: Documentation Improvement
**Spec**: [specs/001-doc-improvement/spec.md](../specs/001-doc-improvement/spec.md)

## Technical Decisions

### Documentation Toolchain

**Decision**: Use MkDocs Material for documentation site generation

**Rationale**: 
- Already configured in the project (mkdocs.yml exists with Material theme)
- Provides professional, searchable documentation with minimal setup
- Supports Markdown natively
- GitHub Pages integration is straightforward
- Consistent with project's existing tooling

**Alternatives considered**: 
- Hugo: More complex, requires additional setup
- Docusaurus: React-based, heavier for a Go library
- Sphinx: Python-based, not ideal for Go projects
- Plain GitHub Pages: Lacks features like search, theme customization

---

### Documentation Structure

**Decision**: Organize documentation into four main sections: Quick Start, Concepts, Tutorials, References

**Rationale**:
- **Quick Start**: Provides immediate value for new users
- **Concepts**: Explains the "what" and "why" of library abstractions
- **Tutorials**: Shows practical "how" with runnable examples
- **References**: Provides detailed API documentation links
- This structure follows common patterns for library documentation (e.g., Go standard library, popular Go frameworks)

**Alternatives considered**:
- Single-page documentation: Too overwhelming for a library with multiple concepts
- Concepts-first approach: Less beginner-friendly
- Tutorials-first approach: Users need conceptual understanding first

---

### Doc Comments Standard

**Decision**: Follow standard Go documentation conventions

**Rationale**:
- Consistent with Go ecosystem expectations
- Works seamlessly with `go doc` command
- Provides good IDE integration (VS Code, GoLand, etc.)
- Well-documented and widely understood by Go developers

**Conventions to follow**:
- Start with the name of the symbol being documented
- Describe the purpose, not the implementation
- For functions/methods: explain parameters and return values
- Use complete sentences
- Keep it concise but informative

**Example**:
```go
// Event represents a domain event with a unique identifier and payload.
// Events are the fundamental building blocks of event-driven applications.
type Event interface {
    ID() string
    Type() Type
    Payload() Payload
    Context() context.Context
}
```

---

### Concept Documentation Style

**Decision**: Keep concept explanations short (3-4 sentences max), non-technical, and focused on value

**Rationale**:
- Developers want to understand concepts quickly
- Long explanations are often skipped
- Technical details belong in tutorials or API documentation
- Focus on "what it is" and "why it matters" rather than "how it works"

**Example format**:
```markdown
### Event

An Event is a representation of something that happened in your application. 
Each event has a unique identifier, a type that categorizes it, and a payload 
containing the data. Events are immutable once created.
```

---

### Tutorial and Example Relationship

**Decision**: Each tutorial MUST link to exactly one runnable example in the examples/ folder

**Rationale**:
- Tutorials are more valuable with accompanying code
- Examples serve as both documentation and testable code
- One-to-one mapping keeps the structure simple and maintainable
- Examples can be run independently to verify behavior

**Structure**:
```
docs/tutorials/
├── basic-pubsub.md          # Tutorial documentation
└── ...

examples/
├── basic-pubsub/
│   └── main.go              # Runnable example
└── ...
```

---

### References Section Approach

**Decision**: Provide direct links to Go pkg documentation without additional content

**Rationale**:
- Go pkg documentation is comprehensive and up-to-date
- Avoids duplication of effort
- Single source of truth for API details
- Links are easy to maintain (point to pkg.go.dev)

**Format**:
```markdown
## References

- [event package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/event)
- [carrier package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/carrier)
- [inmemory package](https://pkg.go.dev/github.com/thomas-marquis/it-happened/inmemory)
```

---

## Best Practices Identified

### For Documentation

1. **Progressive Disclosure**: Start with simple concepts, link to detailed information
2. **Code Examples**: Every concept should have at least one code example
3. **Consistency**: Use the same terminology throughout all documentation
4. **Accessibility**: Use clear language, avoid jargon, explain acronyms on first use
5. **Maintainability**: Keep documentation close to the code it describes

### For Go Doc Comments

1. **Complete Coverage**: Every exported symbol (type, interface, function, method) needs documentation
2. **First Sentence**: Should be a complete sentence that can stand alone
3. **Parameter Documentation**: Use `@param name description` format for functions
4. **Return Value Documentation**: Use `@return description` or describe in text
5. **Examples**: Include usage examples where helpful

### For MkDocs

1. **Navigation**: Keep navigation simple and intuitive
2. **Depth**: Limit nesting to 2-3 levels maximum
3. **Cross-references**: Link between related pages
4. **Search**: Ensure all important terms are searchable
5. **Mobile-friendly**: Test on mobile devices

---

## Tools and Commands

### Documentation Generation
- `mkdocs serve` - Local development server with live reload
- `mkdocs build` - Build static site
- `mkdocs deploy` - Deploy to GitHub Pages

### Code Documentation
- `go doc <symbol>` - View documentation for a symbol
- `go doc -all <package>` - View all documentation for a package
- `godoc -http=:6060` - Local documentation server

### Validation
- `go vet ./...` - Check for common Go issues
- `gofmt -d .` - Check formatting
- `./tools/lint.sh` - Run project linter

---

## Open Questions (None)

All technical decisions have been resolved. No blocking questions remain.