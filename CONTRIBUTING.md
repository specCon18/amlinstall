# Contributing

Thank you for your interest in contributing to Auto MelonLoader Installer.

This document describes the architectural intent, code layout, and development conventions used in the project. It is intended for contributors and maintainers rather than end users.

---

## Design Goals

### TUI-first, CLI-secondary

This project is deliberately **TUI-first**.

- The default execution path always launches the terminal UI.
- Interactive behavior, validation, and user feedback belong exclusively in the TUI.
- The TUI is optimized for human-driven workflows.

The CLI exists as a **secondary interface** and must remain:

- Non-interactive
- Fully flag-driven
- Suitable for scripting and CI environments
- Free of prompts, confirmations, or UI state

CLI commands should never attempt to replicate TUI behavior.

---

### Separation of concerns

The codebase is structured to keep responsibilities clearly isolated:

- GitHub and git-related logic is UI-agnostic and reusable
- Cobra commands act only as orchestration layers
- The TUI is responsible solely for presentation and interaction
- Side effects (network, filesystem) are explicit and bounded

This separation is enforced to keep the codebase maintainable and testable.

---

## Project Structure

```text
.
├── cmd/                     # Cobra command definitions and CLI entrypoints
│   ├── doc.go               # Package documentation for CLI commands
│   ├── root.go              # Root command; launches the TUI by default
│   ├── getAsset.go          # CLI subcommand to download a release asset by tag
│   ├── getTags.go           # CLI subcommand to list repository tags
│   └── version.go           # Version subcommand
│
├── config/                  # Application configuration initialization (Viper)
│   ├── doc.go               # Package documentation
│   └── config.go            # Loads config files, env vars, and defaults
│
├── internal/
│   ├── ghrel/               # GitHub release and git-tag domain logic
│   │   ├── doc.go           # Package documentation
│   │   └── ghrel.go         # Tag discovery and asset download implementation
│   │
│   └── logger/              # Centralized structured logging
│       ├── doc.go           # Package documentation
│       └── logger.go        # Logger initialization and helpers
│
├── tui/                     # Bubble Tea terminal user interface
│   ├── doc.go               # Package documentation
│   ├── model.go             # UI state, inputs, validation, and helpers
│   ├── update.go            # Event handling, keybindings, and async commands
│   ├── view.go              # UI rendering and layout
│   └── run.go               # Program entrypoint and Bubble Tea setup
│
├── main.go                  # Application entrypoint; calls cmd.Execute()
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── README.md                # User-facing documentation
└── CONTRIBUTING.md          # Contributor documentation
```

---

## Generating Additional Subcommands

New CLI subcommands should be generated using `cobra-cli`:

```sh
cobra-cli add <command-name>
```

### Rules for CLI commands

All CLI subcommands must:

- Be non-interactive
- Use explicit flags for all inputs
- Print results to stdout
- Return errors via exit codes
- Avoid any TUI or prompt-based behavior

CLI commands should act as thin orchestration layers and delegate all logic to internal packages.

---

## Development notes

- Go formatting and idiomatic structure are expected
- Package-level documentation (`doc.go`) is required
- Static analysis warnings from `gopls` should be addressed, not suppressed
- Nix is used for reproducible builds but is not required for development

Contributions should preserve the TUI-first design and avoid introducing cross-layer coupling.

