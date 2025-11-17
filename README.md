[![project chat](https://img.shields.io/github/discussions/juspay/omnix)](https://github.com/juspay/omnix/discussions)
[![Naiveté Compass of Mood](https://img.shields.io/badge/naïve-FF10F0)](https://compass.naivete.me/ "This project follows the 'Naiveté Compass of Mood'")

# omnix

<img width="10%" src="./doc/favicon.svg">

*Pronounced [`/ɒmˈnɪks/`](https://ipa-reader.com/?text=%C9%92m%CB%88n%C9%AAks&voice=Geraint)*

Omnix aims to supplement the [Nix](https://nixos.asia/en/nix) CLI to improve developer experience.

## Usage

See <https://omnix.page/>

## Developing

**Note:** This project is currently being migrated from Rust to Go (see `DESIGN_DOCUMENT.md`). Phase 1 is complete with both Rust and Go code coexisting during the transition.

### Rust Development (Production)

1. [Install Nix](https://nixos.asia/en/install)
1. [Setup `direnv`](https://nixos.asia/en/direnv)
1. Clone this repo, `cd` to it, and run `direnv allow`.

This will automatically activate the nix develop shell. Open VSCode and install recommended extensions, ensuring that direnv activates in VSCode as well.

### Go Development (Migration in Progress)

For working on the Go implementation:

1. Install Go 1.22 or later (or use Nix devShell)
2. Run `go mod download` to fetch dependencies
3. See `GO_QUICKSTART.md` for detailed Go development guide

Quick Go commands:
```sh
just go-build   # Build Go version
just go-test    # Run Go tests
just go-ci      # Full Go CI (format, lint, test, build)
```

### Running locally

**Rust version:**
```sh
just watch # Or `just w`; you can also pass args, e.g.: `just w show`
```

**Go version (in development):**
```sh
just go-run [args]  # Run Go version
```

### Nix workflows

Inside the nix develop shell (activated by direnv) you can use any of the `cargo` or `rustc` commands, as well as [`just`](https://just.systems/) workflows. Nix specific commands can also be used to work with the project:

```sh
# Full nix build of CLI
nix build

# Build and run the CLI
nix run
```

### Contributing

>[!TIP]
> Run `just pca` to autoformat the source tree.

- Run `just ci` to **run CI locally**.
- Add **documentation** wherever useful.
    - Run `just doc run` to preview website docs; edit, and run `just doc check`
    - To preview Rust API docs, run `just doc cargo`.
- Changes must accompany a corresponding `history.md` entry.[^cc]

[^cc]: We don't use any automatic changelog generator for this repo.

### Release HOWTO

Begin with a release PR:

- Pick a version
- Update `history.md` to make sure new release header is present
- Run [`cargo workspace publish`](https://github.com/pksunkara/cargo-workspaces?tab=readme-ov-file#publish) in devShell, using the picked version.
