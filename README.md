[![project chat](https://img.shields.io/github/discussions/saberzero1/omnix)](https://github.com/saberzero1/omnix/discussions)
[![Naiveté Compass of Mood](https://img.shields.io/badge/naïve-FF10F0)](https://compass.naivete.me/ "This project follows the 'Naiveté Compass of Mood'")

# omnix

<img width="10%" src="./doc/favicon.svg">

*Pronounced [`/ɒmˈnɪks/`](https://ipa-reader.com/?text=%C9%92m%CB%88n%C9%AAks&voice=Geraint)*

Omnix aims to supplement the [Nix](https://nixos.asia/en/nix) CLI to improve developer experience.

## Usage

See <https://omnix.page/>

## Developing

**Note:** omnix v2.0 is now written in Go. The Rust v1.x codebase is maintained in the v1 branch for reference.

### Go Development (Production)

1. [Install Nix](https://nixos.asia/en/install)
2. [Setup `direnv`](https://nixos.asia/en/direnv)
3. Clone this repo, `cd` to it, and run `direnv allow`.

This will automatically activate the nix develop shell with Go 1.23+ and all development tools. Open VSCode and install recommended extensions, ensuring that direnv activates in VSCode as well.

#### Quick Start

```sh
just go-build   # Build Go binary
just go-test    # Run tests
just go-ci      # Full CI (format, lint, test, build)
just go-run [args]  # Run locally (e.g., just go-run health)
```

See [`GO_QUICKSTART.md`](./GO_QUICKSTART.md) for detailed Go development guide.

### Nix workflows

Inside the nix develop shell (activated by direnv):

```sh
# Build Go version via Nix (recommended for production builds)
nix build

# Build and run the CLI
nix run

# Or run directly without building
nix run . -- health
```

### Rust v1.x (Legacy)

The Rust version (v1.x) is maintained in the v1 branch:

```sh
# Build Rust version
nix build .#omnix-cli

# Work on Rust code (legacy)
git checkout v1
just watch  # Development with live reload
```

See [`MIGRATION_GUIDE.md`](./MIGRATION_GUIDE.md) for migrating from v1.x to v2.0.

### Contributing

>[!TIP]
> Run `just pca` to autoformat the source tree (runs gofmt, nixpkgs-fmt).

- Run `just go-ci` to **run CI locally** (format, lint, test, build).
- Add **documentation** wherever useful.
    - Run `just doc run` to preview website docs; edit, and run `just doc check`
    - For Go API docs, see package documentation with `go doc`
- Changes must accompany a corresponding `history.md` entry.[^cc]

[^cc]: We don't use any automatic changelog generator for this repo.

### Release HOWTO

See [PHASE7_SUMMARY.md](./PHASE7_SUMMARY.md) for release process.

---

## Credits

This project is a fork of the original [Omnix](https://github.com/juspay/omnix) created and maintained by [Juspay Technologies](https://github.com/juspay). We are grateful for their foundational work and contributions to the Nix ecosystem.
