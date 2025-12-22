# Contributing

Thank you for your interest in contributing to Auto MelonLoader Installer.

This document describes the architectural intent, code layout, and development conventions used in the project. It is written for contributors and maintainers rather than end users.

Contributions are expected to preserve the project’s core design principles and maintain a clear separation between interactive and non-interactive behavior.

---

## Design Intent

### TUI-first, CLI-secondary

This project is intentionally TUI-first.

- The default execution path always launches the terminal UI.
- Interactive behavior, validation, confirmations, and user feedback belong exclusively in the TUI.
- The TUI is optimized for human-driven, keyboard-based workflows.

The CLI exists as a secondary interface and must remain:

- Non-interactive
- Fully flag-driven
- Predictable and script-friendly
- Free of prompts, confirmations, or UI state

CLI commands must never attempt to replicate or approximate TUI behavior.

---

### Separation of concerns

The codebase is structured to keep responsibilities strictly isolated:

- GitHub and git-related logic is UI-agnostic and reusable
- Cobra commands act only as orchestration layers
- The TUI is responsible solely for presentation and interaction
- Network and filesystem side effects are explicit and bounded

Cross-layer coupling should be avoided. Logic should live in internal packages, not in UI or command handlers.

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

## Adding or Modifying CLI Commands

New CLI subcommands should be generated using cobra-cli:

```sh
cobra-cli add <command-name>
```

### CLI rules (strict)

All CLI commands must:

- Be non-interactive
- Use explicit flags for all inputs
- Write results to stdout
- Return errors via exit codes
- Avoid confirmations, prompts, or progress UIs

CLI commands should remain thin and delegate all logic to internal packages.

If a feature requires interaction or guidance, it belongs in the TUI.

---

## TUI Development Guidelines

- Keybindings must not interfere with normal text input (avoid single-letter global hotkeys)
- Long-running operations must be asynchronous
- UI state must remain deterministic and recoverable
- Error messages must be explicit and actionable
- The TUI must not embed business logic

Focus management, validation, and feedback are core responsibilities of the TUI layer.

---

## Testing Standard (mandatory)

This project requires integration testing and end-to-end testing for all paths, including failure paths. Testing only the happy path is not acceptable.

### Definition of “done”

A change is considered complete only when:

- Every new feature path has at least one integration or e2e test
- Every failure mode and validation path has at least one integration or e2e test
- Tests assert both behavior and outputs (stdout/stderr, exit codes, and state)
- Tests are deterministic (no dependency on live GitHub, network variance, or machine state)
- All tests pass via `go test ./...`

### Required coverage: happy + unhappy paths

At minimum, tests must cover:

- Invalid input / missing required fields (validation errors)
- Network/API failures (timeouts, non-200 responses, invalid JSON)
- Missing assets / missing tags
- Filesystem failures (permission denied, invalid output path, partial writes)
- External tool failures (git not available, git command fails, malformed output)

### Integration strategy (how to test deterministically)

The project should be tested without hitting real GitHub or relying on a developer’s local git setup.

Recommended approach:

1) GitHub API behavior
- Use `httptest.Server` to emulate the GitHub API endpoints.
- Inject a configurable API base URL (the `internal/ghrel` code already has an internal seam for this).
- Cover:
  - 200 OK with assets
  - 404/500 responses with bodies
  - invalid JSON
  - empty asset list
  - missing asset name

2) Git tag discovery (`git ls-remote`)
- Tests must not require a real remote repository.
- Use one of:
  - A fake `git` executable earlier in PATH during tests, returning controlled stdout/stderr, OR
  - Refactor to allow injection of the git runner (preferred if the project evolves).
- Cover:
  - normal tag output (including annotated tags with `^{}`)
  - empty output
  - non-zero exit code with stderr
  - malformed lines

3) Filesystem and atomic writes
- Tests must verify the atomic-write behavior:
  - successful write produces the expected file contents
  - failures do not leave partial destination files behind
- Cover:
  - create directory success/failure
  - rename failure behavior where possible
  - output path empty handling

### End-to-end expectations

In addition to package-level integration tests, e2e tests should exercise:

- CLI commands (command execution, stdout formatting, exit codes)
- TUI command wiring at the message/cmd layer (at minimum: validation + action triggers)

Full terminal rendering tests are optional, but the action paths and validation flows must be covered.

---

## Code Quality Expectations

- Idiomatic Go formatting and structure are required
- Package-level documentation (`doc.go`) is mandatory
- LSP diagnostics and static analysis warnings should be fixed, not suppressed
- New logic should be testable in isolation
- Avoid introducing hidden global state

---

## Development Notes

- Nix is used for reproducible builds but is not required for development
- Go tooling (`go build`, `go test`, `go vet`) should pass cleanly
- Changes must not break TUI-first behavior or default execution paths

Contributions that weaken separation of concerns or blur the TUI/CLI boundary are unlikely to be accepted.

