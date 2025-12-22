# Auto MelonLoader Installer (Go)

[![Go Version](https://img.shields.io/github/go-mod/go-version/specCon18/automelonloaderinstallergo)](https://go.dev/)
[![License](https://img.shields.io/github/license/specCon18/automelonloaderinstallergo)](LICENSE)

[![Latest Release](https://img.shields.io/github/v/release/specCon18/automelonloaderinstallergo)](https://github.com/specCon18/automelonloaderinstallergo/releases)
[![Downloads](https://img.shields.io/github/downloads/specCon18/automelonloaderinstallergo/total)](https://github.com/specCon18/automelonloaderinstallergo/releases)

![OS Linux](https://img.shields.io/badge/OS-Linux-blue)
![OS macOS](https://img.shields.io/badge/OS-macOS-blue)
![Arch amd64](https://img.shields.io/badge/arch-amd64-success)
![Arch arm64](https://img.shields.io/badge/arch-arm64-success)

[![Nix Flake](https://img.shields.io/badge/Nix-Flake-5277C3?logo=nixos&logoColor=white)](https://nixos.wiki/wiki/Flakes)
[![Nix Flake Input](https://img.shields.io/badge/flake-input-automelonloaderinstallergo-5277C3?logo=nixos&logoColor=white)](https://github.com/specCon18/automelonloaderinstallergo)

## Overview

Auto MelonLoader Installer is a **terminal-based tool for discovering and downloading MelonLoader releases from GitHub**, written in Go and packaged for portability.

The application is designed to work well in both **interactive** and **scripted** environments:

- By default, it launches a **keyboard-driven terminal UI (TUI)** for interactive use.
- It also exposes **explicit CLI subcommands** for automation, scripting, and CI workflows.

The binary is built with portability in mind and can be distributed as a self-contained application image using **Nix**.

---

## Getting Started

### Default (TUI mode)

Running the binary with no arguments launches the interactive terminal UI:

```sh
amlinstall
```

The TUI allows you to:

- Enter a GitHub repository owner and name
- Fetch and select available release tags
- Specify a release asset and output location
- Download the selected asset with clear status and error feedback

All interaction is keyboard-driven.

---

### CLI mode (non-interactive)

For scripting or automation, the application provides dedicated CLI subcommands.

#### List available tags

```sh
amlinstall getTags --owner LavaGang --repo MelonLoader
```

This prints one tag per line to stdout.

#### Download a release asset

```sh
amlinstall getAsset \
  --owner LavaGang \
  --repo MelonLoader \
  --tag v0.5.7 \
  --asset MelonLoader.x64.zip \
  --output ./downloads/MelonLoader.x64.zip
```

If `--output` is omitted, the default is:

```text
./downloads/<asset name>
```

#### Authentication

GitHub authentication is optional and resolved in this order:

1. `--token` flag (if provided)
2. `GITHUB_TOKEN` environment variable
3. Unauthenticated access (subject to GitHub rate limits)

---

## Design Goals

### TUI-first, CLI-secondary

This project is intentionally **TUI-first**:

- The primary user experience is interactive and optimized for human use.
- The root command always launches the TUI by default.
- The TUI provides guided input, validation, and immediate feedback.

The CLI exists as a **secondary interface**, designed to be:

- Explicit (required flags, no prompts)
- Predictable (stable stdout output)
- Suitable for automation, scripting, and CI environments

This separation avoids mixing interactive behavior into CLI workflows while keeping both interfaces clean and maintainable.

### Clear separation of concerns

The codebase is structured to keep responsibilities isolated:

- GitHub and git-related logic lives in reusable internal packages
- Cobra commands act only as orchestration layers
- The TUI is responsible solely for presentation and interaction

---

## Why Nix?

This project uses **Nix** for packaging and distribution to ensure builds are:

- **Reproducible** — the same inputs always produce the same binary
- **Portable** — the resulting artifact does not depend on the host system’s Go version or libraries
- **Self-contained** — runtime dependencies are bundled, reducing “works on my machine” issues

Using Nix allows the application to be built as an **AppImage-like artifact** that can be run on a wide range of systems without requiring Go, Cobra, or Bubble Tea to be installed locally.

This makes the installer well-suited for:

- End users who just want a working binary
- CI pipelines that need deterministic builds
- Long-term maintenance without dependency drift

Nix is used strictly as a **build and packaging tool**; it is not required to run the resulting binary.

### Using as a Nix flake input

This project can be consumed directly as a **Nix flake input**.

#### Multi-arch flake input example

This snippet exposes the package from this flake for **multiple systems** (Linux/macOS; amd64/arm64), using the standard `flake-utils` pattern.

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    automelonloaderinstallergo.url =
      "github:specCon18/automelonloaderinstallergo";
  };

  outputs = { self, nixpkgs, flake-utils, automelonloaderinstallergo, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      {
        # Re-export the upstream package for the current system.
        packages.default =
          automelonloaderinstallergo.packages.${system}.default;

        # Optional: provide an app so `nix run` works for this flake, too.
        apps.default =
          automelonloaderinstallergo.apps.${system}.default;
      });
}
```

Build for your current system:

```sh
nix build .#default
```

Build for a specific system (example: x86_64-linux):

```sh
nix build .#packages.x86_64-linux.default
```

#### `nix run` examples

Run the TUI (default behavior):

```sh
nix run github:specCon18/automelonloaderinstallergo
```

Run a subcommand via `--` (example: list tags):

```sh
nix run github:specCon18/automelonloaderinstallergo -- \
  getTags --owner LavaGang --repo MelonLoader
```

Download an asset via `nix run`:

```sh
nix run github:specCon18/automelonloaderinstallergo -- \
  getAsset \
    --owner LavaGang \
    --repo MelonLoader \
    --tag v0.5.7 \
    --asset MelonLoader.x64.zip \
    --output ./downloads/MelonLoader.x64.zip
```

---

## Project Structure

*The project is organized to keep GitHub release logic, CLI commands, and the TUI cleanly separated, with a TUI-first user experience and scriptable CLI subcommands.*

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
└── README.md                # Project overview and usage documentation
```

---

## Generating Additional Subcommands

Additional CLI subcommands can be generated using `cobra-cli`:

```sh
cobra-cli add <command-name>
```

Generated commands should remain non-interactive and follow the existing flag-based pattern.

