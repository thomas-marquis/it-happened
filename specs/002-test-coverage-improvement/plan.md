# Implementation Plan: Test Coverage Improvement

**Branch**: `TBD` | **Date**: 2026-06-19 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-test-coverage-improvement/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

This feature aims to increase the project's test coverage by implementing comprehensive unit tests for currently untested components, with a focus on the event bus implementations (inmemory) and event carriers (All, Sequence). It also establishes a documented testing strategy for asynchronous components and adds a coverage badge to the README. The technical approach uses gomock for mock generation, testify for assertions, and follows the project's test-first development principles. Carriers' tests must be as readable as possible, and mocks should only be generated if needed, via the gen.go file.

## Technical Context

**Language/Version**: Go 1.25+

**Primary Dependencies**: testify/assert, testify/require, go.uber.org/mock/gomock, github.com/stretchr/testify

**Storage**: N/A (library project, no persistent storage)

**Testing**: go test, testify/assert, testify/require, gomock, race detector

**Target Platform**: Any platform supporting Go 1.25+

**Project Type**: library

**Performance Goals**: Tests should execute quickly for local development (aim for sub-second execution for individual test files)

**Constraints**: 
- All tests must pass the race detector (`-race` flag)
- Mocks must be generated via `go generate ./...` and stored in appropriate mocks/ directories
- Tests must follow the test-first approach (Red-Green-Refactor cycle)
- All tests must use t.Run() subtests with Given/When/Then comments

**Scale/Scope**: 
- Unit tests for 3 packages (event, carrier, inmemory)
- ~10-15 untested components to cover
- Integration tests for asynchronous event flow verification
- Documentation updates (CONTRIBUTE.md Testing Strategy section, README.md badge)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle Compliance Evaluation

| Principle | Requirement | Compliance | Justification |
|-----------|-------------|------------|---------------|
| **I. Event-First Design** | All features designed around event-driven patterns | ✅ PASS | Feature focuses on testing event-driven components (bus, carriers) which reinforces event-first design |
| **II. Test-First Development** (NON-NEGOTIABLE) | Tests MUST be written before implementation | ✅ PASS | Feature is explicitly about writing tests first; follows Red-Green-Refactor cycle |
| **III. Clean Interface Design** | Public APIs defined through clear, minimal interfaces | ✅ PASS | Tests will verify interface contracts; mocks will be generated for interfaces |
| **IV. Type Safety and Contracts** | All types strongly typed | ✅ PASS | Go's type system ensures type safety; tests will verify contract adherence |
| **V. Observability and Traceability** | Events support tracing via ChainRef and ChainPosition | ✅ PASS | Tests will verify event flow and correlation; coverage ensures observability code is tested |
| **VI. Simplicity and Composability** | Components small, focused, composable | ✅ PASS | Tests target individual components; carriers tested for composition behavior |
| **VII. Quality Gates** (NON-NEGOTIABLE) | All contributions pass quality gates | ✅ PASS | Feature includes: code clean & documented, tests pass, linting passes, CI passes, documentation updated |

### Development Workflow Compliance

| Workflow Rule | Compliance | Notes |
|---------------|------------|-------|
| Test-first approach | ✅ PASS | Tests will be written before any implementation changes |
| Mocks generated using mockgen | ✅ PASS | User specified gomock/mockgen for mock generation |
| Mocks stored in mocks/ directory | ✅ PASS | Project structure already has mocks/ directories |
| Mocks NOT edited manually | ✅ PASS | Will use `go generate ./...` for regeneration |
| Code review verifies constitution compliance | ✅ PASS | Plan references specific constitution principles |

### Quality Standards Compliance

| Standard | Requirement | Compliance | Notes |
|----------|-------------|------------|-------|
| Testing | testify/assert, testify/require, gomock | ✅ PASS | User explicitly specified these tools |
| Testing | t.Run() subtests with Given/When/Then | ✅ PASS | All new tests will follow this structure |
| Documentation | Go doc comments for all public APIs | ✅ PASS | N/A for this feature (no new public APIs) |
| Code Style | Standard Go conventions, ./tools/lint.sh | ✅ PASS | All code will pass linting |

**GATE STATUS**: ✅ **ALL GATES PASS** - Proceed to Phase 0 research

## Project Structure

### Documentation (this feature)

```text
specs/002-test-coverage-improvement/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
.
├── event/
│   ├── bus.go
│   ├── event.go
│   ├── matcher.go
│   ├── matcher_test.go
│   ├── notifier.go
│   ├── option.go
│   └── subscriber.go
├── carrier/
│   ├── all.go
│   ├── carrier.go
│   └── sequence.go
├── inmemory/
│   ├── bus.go
│   └── options.go
├── internal/
│   └── mocks/
│       └── ... (generated mocks)
├── gen.go
├── go.mod
├── go.sum
├── Makefile
├── CONTRIBUTE.md
├── README.md
└── tools/
    └── lint.sh
```

**Structure Decision**: Single project structure with existing package organization (event, carrier, inmemory). Test files will be added to each package directory following Go conventions (*_test.go). Documentation updates will be made to CONTRIBUTE.md and README.md at the repository root.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. All constitution principles are satisfied by the feature design.
