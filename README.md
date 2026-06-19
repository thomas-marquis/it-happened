# it-happened

[![Go Reference](https://pkg.go.dev/badge/github.com/thomas-marquis/it-happened.svg)](https://pkg.go.dev/github.com/thomas-marquis/it-happened)
[![CI](https://github.com/thomas-marquis/it-happened/actions/workflows/ci.yaml/badge.svg)](https://github.com/thomas-marquis/it-happened/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/badge/Coverage-66%25-yellow)](https://github.com/thomas-marquis/it-happened/actions/workflows/coverage-badge.yml)
[![License](https://img.shields.io/github/license/thomas-marquis/it-happened)](LICENSE)

Event management library written in Go simplifying event driven application development

<p align="center">
  <img src="docs/assets/images/logo-tr.png" width="200" alt="it-happened logo">
</p>

## ✨ Features

- **Asynchronous Event Bus**: Decouple your components with a robust pub-sub system
- **Event Chaining**: Track related events across workflows using ChainRef and ChainPosition
- **Powerful Matchers**: Subscribe to events using precise criteria (by type, followup relationship, etc.)
- **Event Carriers**: Orchestrate complex workflows with All (parallel) and Sequence (sequential) carriers
- **Automated Lifecycle**: Carriers handle timeouts, concurrency, and completion tracking

## 📦 Installation

### Requirements

- Go 1.25 or higher

### Installation process

```bash
go get github.com/thomas-marquis/it-happened
```

## 📚 Documentation

- [Project Documentation](https://thomas-marquis.github.io/it-happened/) - Complete documentation with:
  - [Quick Start](https://thomas-marquis.github.io/it-happened/) - Get started in 10 minutes
  - [Concepts](https://thomas-marquis.github.io/it-happened/concepts/) - Core library abstractions
  - [Tutorials](https://thomas-marquis.github.io/it-happened/tutorials/) - Practical examples
  - [API References](https://thomas-marquis.github.io/it-happened/references/) - Package documentation links
- [Go Package Documentation](https://pkg.go.dev/github.com/thomas-marquis/it-happened)

## 💻 Usage

See the [Quick Start](https://thomas-marquis.github.io/it-happened/) guide to get started, or explore the [examples](examples/) directory for runnable code samples.

## 🤝 Contribute

All contributions are welcome! Feel free to open an issue or submit a PR. ✨

Check out [CONTRIBUTE.md](CONTRIBUTE.md) for more details.
