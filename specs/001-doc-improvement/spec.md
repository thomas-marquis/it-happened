# Feature Specification: Documentation Improvement

**Feature Branch**: `feat/newApi`

**Created**: 2026-06-19

**Status**: Draft

**Input**: User description: "I'd like to improve the library's documentation for its users: clear and relevant doc comments for all exported object + MKdoc pages. The MKdoc must follow the structure quick-start/concepts/references/tutorial. The tutorials must link to the examples folder (that contains one subfolder per tutorial). The tutorial part must show off the most importants use case only, not everything. The part abount concept must braink down the main library's ... concepts. Such as: what an event is, a followup, a chain, etc. No complicated or technical explaination here. Keep it hort and neat. The references part must simply link to the go pkg documentation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developers can understand library concepts quickly (Priority: P1)

Developers new to the it-happened library need to understand the core concepts (Event, Followup, Chain, ChainRef, ChainPosition, Payload, Bus, Subscriber, Matcher) without diving into implementation details. The concepts documentation should provide clear, concise explanations that enable developers to grasp the event-driven patterns the library implements.

**Why this priority**: Without understanding the core concepts, developers cannot use the library effectively. This is the foundation for all other documentation.

**Independent Test**: Can be fully tested by having a developer read the concepts page and correctly explain what each concept means and how they relate to each other.

**Acceptance Scenarios**:

1. **Given** a developer reads the concepts documentation, **When** they are asked about Events, **Then** they can explain that Events are the fundamental building blocks with unique IDs and payloads
2. **Given** a developer reads the concepts documentation, **When** they are asked about Chains, **Then** they can explain that Chains are sequences of related events tracked via ChainRef and ChainPosition
3. **Given** a developer reads the concepts documentation, **When** they are asked about Followups, **Then** they can explain that Followups are new events created from parent events within a chain

---

### User Story 2 - Developers can get started quickly with the library (Priority: P1)

Developers need a quick-start guide that gets them up and running with the library in minimal time. This should cover the most common use case (basic event publishing and subscription) without overwhelming them with all features.

**Why this priority**: Quick-start is the first thing developers see and determines their initial experience with the library. A good quick-start reduces friction and increases adoption.

**Independent Test**: Can be fully tested by having a developer follow the quick-start guide and successfully implement a basic event-driven workflow.

**Acceptance Scenarios**:

1. **Given** a developer follows the quick-start guide, **When** they reach the end, **Then** they have a working example of event publishing and subscription
2. **Given** the quick-start guide, **When** a developer reads it, **Then** they understand the minimal setup required
3. **Given** the quick-start guide, **When** a developer completes it, **Then** they can modify it for their own use case

---

### User Story 3 - Developers can explore practical examples through tutorials (Priority: P2)

Developers need practical tutorials that demonstrate important use cases with links to runnable examples. Each tutorial should focus on one key use case (e.g., event chaining, filtering with matchers, using carriers) and link to a corresponding example in the examples/ folder.

**Why this priority**: Tutorials bridge the gap between understanding concepts and applying them to real-world scenarios. They are essential for intermediate users.

**Independent Test**: Can be fully tested by having a developer follow a tutorial and successfully run the linked example.

**Acceptance Scenarios**:

1. **Given** a developer reads the event chaining tutorial, **When** they follow the link to the example, **Then** they can run the example and see event chaining in action
2. **Given** a developer reads a tutorial, **When** they want to see the code, **Then** they find a clear link to the corresponding example folder
3. **Given** the tutorials section, **When** a developer browses it, **Then** they see only the most important use cases covered

---

### User Story 4 - Developers can access API reference documentation (Priority: P3)

Developers need easy access to the Go package documentation for all exported types, interfaces, and functions. The references section should provide direct links to the Go pkg documentation.

**Why this priority**: API documentation is essential for developers who need detailed information about specific functions or types. However, it's less critical than conceptual understanding.

**Independent Test**: Can be fully tested by verifying that all links in the references section point to valid Go pkg documentation pages.

**Acceptance Scenarios**:

1. **Given** a developer visits the references section, **When** they click on a link, **Then** they are taken to the corresponding Go pkg documentation
2. **Given** the references section, **When** a developer looks for a specific type, **Then** they find a link to its documentation

---

### User Story 5 - All exported objects have clear doc comments (Priority: P1)

All exported types, interfaces, functions, and methods in the library must have clear, relevant Go doc comments that explain their purpose and usage. This ensures that `go doc` and IDE tooltips provide useful information.

**Why this priority**: Doc comments are the primary source of documentation for Go developers. Without them, the library is much harder to use.

**Independent Test**: Can be fully tested by running `go doc` on each exported symbol and verifying that the output is clear and helpful.

**Acceptance Scenarios**:

1. **Given** an exported type in the library, **When** a developer runs `go doc <Type>`, **Then** they see a clear description of its purpose
2. **Given** an exported function in the library, **When** a developer hovers over it in their IDE, **Then** they see a helpful tooltip
3. **Given** an exported interface in the library, **When** a developer reads its documentation, **Then** they understand what it represents and how to implement it

---

### Edge Cases

- What happens when a new exported type is added without doc comments? (Should be caught in code review)
- How does the documentation handle breaking changes in future versions? (Document in a changelog or migration guide)
- What if a concept is too complex to explain simply? (Break it down into simpler parts or provide progressive disclosure)
- How are deprecated features documented? (Mark clearly as deprecated with migration path)

## Clarifications

### Session 2026-06-19

- Q: Should the documentation include ALL implemented library concepts, or only the core ones? → A: The documentation must contain all the global concepts. For example, the concept of carrier must be described in the "concept" section of the documentation, but not necessarily all the concrete carrier implementation (all, sequence, etc.).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Documentation MUST be organized into four main sections: Quick Start, Concepts, Tutorials, and References
- **FR-002**: Quick Start section MUST provide a minimal working example that demonstrates the core value proposition
- **FR-003**: Concepts section MUST explain all global library concepts including: Event, Type, Payload, Chainable, ChainableEvent, Chain, ChainRef, ChainPosition, Followup, Bus, Subscriber, Matcher, Option, Notifier, Carrier, and CompletionCondition in simple, non-technical language
- **FR-004**: Each concept explanation MUST be short (maximum 3-4 sentences) and focus on what it is and why it matters
- **FR-005**: Tutorials section MUST cover the most important use cases only (event publishing/subscription, event chaining, using matchers, using carriers)
- **FR-006**: Each tutorial MUST link to a corresponding runnable example in the examples/ folder
- **FR-007**: Each example folder MUST contain a main.go file that demonstrates the tutorial concept
- **FR-008**: References section MUST provide direct links to Go pkg documentation for the main packages (event, carrier, inmemory)
- **FR-009**: All exported types, interfaces, functions, and methods MUST have Go doc comments
- **FR-010**: Doc comments MUST follow Go conventions (start with the name, describe purpose, explain parameters and return values for functions/methods)
- **FR-011**: Documentation MUST be written in clear, concise English without unnecessary jargon
- **FR-012**: Tutorial examples MUST be self-contained and runnable with `go run .`
- **FR-013**: All documentation pages MUST be properly formatted in Markdown
- **FR-014**: Documentation MUST use consistent terminology across all pages

### Key Entities *(include if feature involves data)*

- **Documentation Section**: A top-level category in the documentation (Quick Start, Concepts, Tutorials, References)
- **Concept**: An explanation of a core library abstraction (Event, Chain, Followup, etc.)
- **Tutorial**: A step-by-step guide demonstrating a specific use case
- **Example**: A runnable code sample in the examples/ folder that accompanies a tutorial
- **Doc Comment**: Documentation for a Go symbol that appears in `go doc` output
- **Reference Link**: A URL pointing to Go pkg documentation

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of exported types, interfaces, functions, and methods have Go doc comments
- **SC-002**: All four documentation sections (Quick Start, Concepts, Tutorials, References) are present and accessible
- **SC-003**: At least 4 tutorials are available, each with a corresponding runnable example
- **SC-004**: All links in the documentation (to examples, to Go pkg docs) are valid and working
- **SC-005**: Documentation builds successfully with MkDocs Material
- **SC-006**: All example code in tutorials runs without errors
- **SC-007**: Concepts are explained in 3-4 sentences or less each
- **SC-008**: Quick Start guide can be completed in under 10 minutes by a developer familiar with Go

## Assumptions

- Documentation will be hosted on GitHub Pages via MkDocs Material (already configured in mkdocs.yml)
- Target audience is Go developers with at least basic familiarity with Go syntax and concepts
- The library's current codebase structure and exported API will remain stable during documentation development
- Examples will be kept in the examples/ directory with one subfolder per tutorial
- Each tutorial will have exactly one corresponding example
- Doc comments will follow standard Go documentation conventions
- The quick-start guide will cover the minimal setup: creating an event, publishing it, and subscribing to it
- Technical explanations in the concepts section will be kept minimal, focusing on "what" and "why" rather than "how"
- The references section will simply list links to the Go pkg documentation without additional content