<!--
Sync Impact Report
==================
Version change: 1.0.0 → 1.1.0 (MINOR: Added test naming convention rule)
Added sections:
  - Test Naming Convention in Quality Standards > Testing
Modified sections:
  - Quality Standards > Testing (added test naming rule)
Templates requiring updates: 
  - ✅ constitution-template.md (referenced but not modified - this is the working copy)
  - ⚠ spec-template.md (no changes needed - aligns with principles)
  - ⚠ plan-template.md (no changes needed - aligns with principles)
  - ⚠ tasks-template.md (no changes needed - aligns with principles)
Follow-up TODOs: None
-->

# it-happened Constitution

## Core Principles

### I. Event-First Design
Every feature and component MUST be designed around event-driven patterns. The library MUST provide clear, composable abstractions for event creation, publishing, subscription, and handling. Event types MUST be strongly typed and payloads MUST implement the Payload interface with an EventType() method.

**Rationale**: The library's core value proposition is simplifying event-driven application development. All architecture decisions must reinforce this pattern.

### II. Test-First Development (NON-NEGOTIABLE)
Tests MUST be written before implementation. Every feature, bug fix, or enhancement MUST have corresponding test cases that fail before the code is written. The Red-Green-Refactor cycle MUST be strictly followed.

**Rationale**: This ensures high code quality, prevents regressions, and validates design decisions before implementation. This is explicitly stated as mandatory in CONTRIBUTE.md.

### III. Clean Interface Design
All public APIs MUST be defined through clear, minimal interfaces. Interfaces MUST follow Go conventions: single responsibility, focused on behavior not implementation. Interface methods MUST have precise, unambiguous contracts.

**Rationale**: Clean interfaces enable easy testing (via mocks), multiple implementations, and clear usage patterns for library consumers.

### IV. Type Safety and Contracts
All types MUST be strongly typed. Event payloads MUST implement the Payload interface. Function signatures MUST clearly express their contracts through types. Type assertions MUST be avoided in favor of interface-based polymorphism.

**Rationale**: Type safety reduces runtime errors and makes the library more predictable and easier to use correctly.

### V. Observability and Traceability
All events MUST support tracing through ChainRef and ChainPosition. Event IDs MUST be unique and traceable. Library users MUST be able to correlate events across distributed systems.

**Rationale**: Event-driven systems require robust observability to debug issues and understand event flows. The existing implementation already supports this via ChainRef and ChainPosition.

### VI. Simplicity and Composability
Components MUST be small, focused, and composable. Complex functionality MUST be achieved through composition of simple parts, not through monolithic components. Each package MUST have a single, clear purpose.

**Rationale**: Simplicity enables maintainability, testability, and flexibility. The existing package structure (event, carrier, inmemory) exemplifies this principle.

### VII. Quality Gates (NON-NEGOTIABLE)
All contributions MUST pass through defined quality gates before being considered complete: code must be clean and documented; all tests must pass; linting must pass without errors; CI workflows must pass; relevant documentation must be updated; examples must be provided for new features.

**Rationale**: Consistent quality ensures reliability and maintainability of the library. This is explicitly defined in CONTRIBUTE.md's "Definition of Done".

## Development Workflow

All development MUST follow the test-first approach. Mocks MUST be generated using mockgen and stored in the mocks/ directory. Mocks MUST NOT be edited manually - they must be regenerated using `go generate ./...`.

Code review MUST verify compliance with all principles in this constitution. Complexity MUST be justified with clear rationale. All PR descriptions MUST reference which constitution principles are addressed or impacted.

## Quality Standards

**Testing**: Use testify/assert for assertions and testify/require for setup preconditions. Structure tests with t.Run() subtests and Given/When/Then comments. Mock dependencies using gomock. Test names in t.Run() MUST follow the pattern: `should <do or return something>... when <some specific case or condition>`. Except the Given/When/Then, avoid writing comments in the tests

**Documentation**: All public APIs MUST be documented with Go doc comments. Examples MUST be provided in the examples/ directory for all significant features.

**Code Style**: Follow standard Go conventions. Run the linting script (./tools/lint.sh) before submitting PRs. Code MUST pass all CI checks.

## Governance

This Constitution supersedes all other practices and guidelines. All PRs and code reviews MUST verify compliance with these principles. Violations MUST be justified with explicit rationale in the PR description or code comments.

Use the existing specs/constitution.md and CONTRIBUTE.md files for runtime development guidance. The Definition of Done in CONTRIBUTE.md MUST be followed for all contributions.

Amendments to this Constitution require:
1. Documentation of the proposed change
2. Approval through PR review process
3. Migration plan for any breaking changes to established practices

**Version**: 1.1.0 | **Ratified**: 2026-06-19 | **Last Amended**: 2026-06-19
