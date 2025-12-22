# Auto MelonLoader Installer (Go)

[![Nix Flake](https://img.shields.io/badge/Nix-Flake-5277C3?logo=nixos&logoColor=white)](https://nixos.wiki/wiki/Flakes)
[![Go Version](https://img.shields.io/github/go-mod/go-version/specCon18/automelonloaderinstallergo)](https://go.dev/)
![Arch amd64](https://img.shields.io/badge/arch-amd64-success)

[![License](https://img.shields.io/github/license/specCon18/automelonloaderinstallergo)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/specCon18/automelonloaderinstallergo)](https://github.com/specCon18/automelonloaderinstallergo/releases)



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

## Design Philosophy

This project is intentionally **TUI-first**, with explicit CLI subcommands provided as a secondary, non-interactive interface.

The interactive terminal UI is the primary user experience and is optimized for guided, keyboard-driven workflows. The CLI exists to support automation, scripting, and CI use cases, and is designed to be predictable, flag-driven, and free of prompts.

Architectural decisions prioritize clear separation of concerns, deterministic behavior, and long-term maintainability over feature density.

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

---

### Using as a Nix flake input

This project can be consumed directly as a **Nix flake input**.

#### Multi-arch flake input example

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
        packages.default =
          automelonloaderinstallergo.packages.${system}.default;

        apps.default =
          automelonloaderinstallergo.apps.${system}.default;
      });
}
```

Build for your current system:

```sh
nix build .#default
```

Run the TUI directly:

```sh
nix run github:specCon18/automelonloaderinstallergo
```

Run a CLI subcommand via `nix run`:

```sh
nix run github:specCon18/automelonloaderinstallergo -- \
  getTags --owner LavaGang --repo MelonLoader
```

