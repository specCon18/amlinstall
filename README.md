# Auto MelonLoader Installer Go

## TLDR;
Automated installer for melon loader written in go and built as an app image using nix for portability.

---

## Getting started
Running the binary by itself will execute the tui and request you to select a version and game to install to.
Running the cli using `amlinstall getTags` to get a list of available versions then `amlinstall getAsset --tag --output` to download the desired version.

### Project structure
```sh
. <- Project Root
├── cmd <- All CLI related code is in here
│   ├── doc.go <- please the linter
│   ├── getAsset.go <- all code related to executing `amlinstall getAsset` sub-command
│   ├── getTags.go <- all code related to executing `amlinstall getTags` sub-command
│   ├── root.go <- command router logic
│   └── version.go <- define `amlinstall version` sub-command
├── config <- All viper configuation related code is in here
│   ├── config.go <-
│   └── doc.go <- please the linter
├── flake.lock <-
├── flake.nix <-
├── go.mod <-
├── go.sum <-
├── internal <- all internal packages that are not publicly exposed exist in here per go std reqs
│   ├── ghrel <- all code related to retrieving github release assets
│   │   ├── doc.go <- please the linter
│   │   └── ghrel.go <- all code related to retrieving github release assets
│   └── logger <- all code related to logging using https://github.com/charmbracelet/log
│       ├── doc.go <- please the linter
│       └── logger.go <- all code related to logging using https://github.com/charmbracelet/log
├── justfile <- MAKEFILE tooling just rusty read it its self explanitory.
├── main.go <- entry point
├── README.md <- YOU ARE HERE
└── tui <- all code related to the TUI built on top of https://charm.land 's libs
    ├── doc.go <- please the linter
    ├── model.go <- all code defining the Model of the application this is where the data is structured
    ├── run.go <- TUI entrypoint
    ├── update.go <- all code defining how to update the model based on the users manipulation of the view
    └── view.go <- all code related to defining the user interface.
```

### Generating sub-commands
use `cobra-cli` to generate sub-commands

