# Data Model: Documentation Improvement

**Date**: 2026-06-19
**Feature**: Documentation Improvement
**Spec**: [specs/001-doc-improvement/spec.md](../specs/001-doc-improvement/spec.md)

## Documentation Entities

### Documentation Section

A top-level category in the documentation site.

**Attributes**:
- `title` (string, required): Display name of the section
- `path` (string, required): URL path for the section
- `description` (string, optional): Short description for navigation
- `order` (integer, required): Navigation order (lower = earlier)

**Relationships**:
- Contains one or more Documentation Page entities
- Part of the Documentation Site

**Validation Rules**:
- Path must be unique across all sections
- Title must be human-readable
- Order must be a positive integer

**State Transitions**: N/A (static content)

---

### Documentation Page

A single Markdown file in the documentation.

**Attributes**:
- `title` (string, required): Display name of the page
- `path` (string, required): File path relative to docs/ directory
- `section` (string, required): Parent section name
- `content` (string, required): Markdown content
- `order` (integer, optional): Order within section

**Relationships**:
- Belongs to exactly one Documentation Section
- May reference other Documentation Pages via links
- May reference Example entities

**Validation Rules**:
- Path must end with `.md`
- Content must be valid Markdown
- All internal links must be valid

---

### Concept

An explanation of a core library abstraction.

**Attributes**:
- `name` (string, required): Name of the concept (e.g., "Event", "Chain", "Carrier")
- `description` (string, required): Simple explanation in 3-4 sentences
- `category` (string, required): Classification (e.g., "Core", "Advanced", "Utility")
- `relatedConcepts` (string[], optional): Names of related concepts
- `exampleCode` (string, optional): Short code snippet illustrating the concept

**Relationships**:
- Documented in the Concepts section
- May be referenced by Tutorial entities
- May relate to other Concept entities

**Validation Rules**:
- Description must be ≤ 4 sentences
- Must use non-technical language
- Must explain "what it is" and "why it matters"
- Must not explain implementation details

**Identified Concepts** (from FR-003):
1. Event
2. Type
3. Payload
4. Chainable
5. ChainableEvent
6. Chain
7. ChainRef
8. ChainPosition
9. Followup
10. Bus
11. Subscriber
12. Matcher
13. Option
14. Notifier
15. Carrier
16. CompletionCondition

---

### Tutorial

A step-by-step guide demonstrating a specific use case.

**Attributes**:
- `title` (string, required): Display name of the tutorial
- `path` (string, required): File path in docs/tutorials/
- `description` (string, required): One-sentence summary
- `useCase` (string, required): The use case being demonstrated
- `examplePath` (string, required): Path to corresponding Example
- `prerequisites` (string[], optional): Required knowledge or setup
- `estimatedTime` (string, optional): Estimated completion time

**Relationships**:
- Links to exactly one Example entity (FR-006)
- Documented in the Tutorials section
- Demonstrates one or more Concept entities

**Validation Rules**:
- Must have a corresponding Example (FR-006, FR-007)
- Example path must be valid
- Use case must be one of the most important use cases (FR-005)

**Identified Tutorials** (from FR-005):
1. Event Publishing and Subscription (basic-pubsub)
2. Event Chaining (event-chaining)
3. Using Matchers (using-matchers)
4. Using Carriers (using-carriers)

---

### Example

A runnable code sample that demonstrates a tutorial concept.

**Attributes**:
- `name` (string, required): Name of the example
- `path` (string, required): Directory path in examples/ folder
- `mainFile` (string, required): Path to main.go file
- `description` (string, required): One-sentence description
- `tutorialPath` (string, required): Path to corresponding Tutorial

**Relationships**:
- Corresponds to exactly one Tutorial entity (FR-006, FR-007)
- Contains a main.go file that is self-contained and runnable (FR-012)

**Validation Rules**:
- Must contain a main.go file (FR-007)
- Must be runnable with `go run .` (FR-012)
- Path must match the tutorial's examplePath

**Identified Examples**:
1. basic-pubsub (for Event Publishing and Subscription tutorial)
2. event-chaining (for Event Chaining tutorial)
3. using-matchers (for Using Matchers tutorial)
4. using-carriers (for Using Carriers tutorial)

---

### Doc Comment

Documentation for a Go symbol that appears in `go doc` output.

**Attributes**:
- `symbol` (string, required): Fully qualified symbol name (e.g., "event.Event", "event.Bus.Publish")
- `package` (string, required): Package name
- `kind` (string, required): Symbol kind (type, interface, function, method)
- `content` (string, required): The doc comment text
- `file` (string, required): Source file containing the symbol
- `line` (integer, required): Line number of the symbol

**Relationships**:
- Attached to a specific exported symbol in the codebase
- Appears in `go doc` output
- Visible in IDE tooltips

**Validation Rules**:
- Must follow Go documentation conventions (FR-010)
- Must start with the symbol name
- Must describe purpose, not implementation
- For functions/methods: must explain parameters and return values

**Coverage Requirement**: 100% of exported types, interfaces, functions, and methods (SC-001)

---

### Reference Link

A URL pointing to external Go pkg documentation.

**Attributes**:
- `url` (string, required): Full URL to the documentation
- `text` (string, required): Display text for the link
- `package` (string, required): Package being referenced

**Relationships**:
- Documented in the References section
- Points to Go pkg documentation

**Validation Rules**:
- URL must be valid and accessible
- Must point to pkg.go.dev

**Identified Reference Links** (from FR-008):
1. event package: https://pkg.go.dev/github.com/thomas-marquis/it-happened/event
2. carrier package: https://pkg.go.dev/github.com/thomas-marquis/it-happened/carrier
3. inmemory package: https://pkg.go.dev/github.com/thomas-marquis/it-happened/inmemory

---

## Entity Relationships

```
Documentation Site
├── Documentation Section (4)
│   ├── Quick Start
│   │   └── Documentation Page (index.md)
│   ├── Concepts
│   │   └── Documentation Page (concepts.md)
│   │       └── Concept (16)
│   ├── Tutorials
│   │   └── Documentation Page (tutorials/)
│   │       └── Tutorial (4)
│   │           └── Example (4)
│   └── References
│       └── Documentation Page (references.md)
│           └── Reference Link (3)
└── Codebase
    ├── Doc Comment (50+)
    └── Exported Symbol (50+)
```

## Validation Rules Summary

1. All Documentation Pages must have valid Markdown content
2. All internal links must point to existing pages
3. All Example entities must have a corresponding Tutorial entity
4. All Tutorial entities must have a corresponding Example entity
5. All Doc Comment entities must cover exported symbols
6. All Reference Link URLs must be valid
7. All Concept descriptions must be ≤ 4 sentences
8. All documentation must use consistent terminology