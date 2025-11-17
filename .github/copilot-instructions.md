# Copilot Instructions for omnix

## Overview
Omnix is a Rust CLI tool (8 crates, 117 .rs files) that supplements the Nix CLI. Commands: `om ci` (CI builds), `om health` (environment checks), `om develop` (dev envs), `om show` (flake inspection), `om init` (project scaffolding). Build system: Nix flakes + Cargo workspace. Platforms: x86_64/aarch64 Linux/Darwin.

## Critical: Nix is Mandatory

**DO NOT use plain `cargo build/test`** - it will fail. This project requires Nix with flakes enabled. Build environment variables (`TRUE_FLAKE`, `FALSE_FLAKE`, `FLAKE_METADATA`, `DEFAULT_FLAKE_SCHEMAS`, `INSPECT_FLAKE`, `NIX_SYSTEMS`, etc.) are defined in `nix/envs/default.nix` and injected by Nix during compilation.

**Setup**: Install Nix with flakes, setup direnv, run `direnv allow` in repo root. The `.envrc` activates Nix devShell automatically.

## Build Commands (Always in Nix DevShell)

**Verify devShell active**: `echo $OMNIX_SOURCE` (should show Nix store path). If not: `nix develop --accept-flake-config`

**Build**:
- `nix build --accept-flake-config` - Full build (5-10 min first run, <1 min incremental)
- `nix run --accept-flake-config` - Build and run
- `just watch` (or `just w`) - Development with live reload using bacon; pass args: `just w show`

**Test**:
- `just cargo-test` or `cargo test --release --all-features --workspace` (2-5 min)
- Per-crate: `cargo test -p omnix-cli --release --all-features`
- Tests **require** devShell environment variables

**Lint** (always before committing):
- `just pca` - Auto-format all code (pre-commit: nixpkgs-fmt + rustfmt)
- `just clippy` - Strict lint (treats warnings as errors)
- `just cargo-doc` - Build Rust docs

**CI**:
- `just ci` - Full local CI (10-20 min first run, uses `nix run . ci`)
- `just ci-cargo` - Faster iteration using cargo in devShell

## Project Structure

**Crates** (in `crates/`, Cargo workspace):
- `omnix-cli`: Main CLI entry, command dispatching. Entry: `src/main.rs`, commands: `src/command/*.rs`
- `omnix-ci`: CI command, builds all flake outputs
- `omnix-health`: Health checks for Nix environment
- `omnix-develop`: Development environment setup
- `omnix-init`: Project scaffolding with templates (registry: `crates/omnix-init/registry/`)
- `omnix-common`: Shared utilities, logging
- `nix_rs`: Core Nix interaction library. Key: `src/flake/` (URL parsing, metadata, schema), `src/flake/functions/` (Rust+Nix FFI)
- `omnix-gui`: Dioxus-based GUI (experimental)

**Config Files**:
- Root: `Cargo.toml` (workspace), `flake.nix` (main flake), `flake.lock`, `rust-toolchain.toml` (stable + musl targets), `justfile` (tasks), `bacon.toml` (file watcher), `om.yaml` (omnix config), `.envrc` (direnv)
- Nix: `nix/modules/flake/*.nix` (flake modules), **`nix/envs/default.nix`** (critical: defines build env vars), `crates/*/crate.nix` (per-crate build config)
- VSCode: `.vscode/extensions.json` (recommended: rust-analyzer, direnv, nix-ide), `.vscode/settings.json` (clippy, format on save)

**Documentation** (`doc/`): Built with emanote. `index.md` (homepage), `om/*.md` (command docs), `config.md` (om.yaml guide), `history.md` (changelog - **required for all changes**). Commands: `just doc run` (live server), `just doc check` (link validation).

## CI/CD Pipeline

**GitHub Actions** (`.github/workflows/`):

**ci.yaml** - Main CI (triggers: push to main/PRs, branches `ci/**`):
- Matrix: x86_64-linux, aarch64-linux (main), x86_64-darwin (main), aarch64-darwin
- Steps: 1) `nix build --accept-flake-config`, 2) `nix run . -- ci run --systems "$SYSTEM" --results=$HOME/omci.json`, 3) Upload artifacts, 4) Push to Attic cache (main only), 5) Deploy docs (main only)

**website.yaml** - Documentation deploy to GitHub Pages (called from ci.yaml on main)

**CI Steps** (defined in `om.yaml`):
- `om-show`, `binary-size-is-small` (x86_64-linux), `omnix-source-is-buildable`, `cargo-tests` (x86_64-linux, aarch64-darwin), `cargo-clippy`, `cargo-doc`
- Sub-flakes: `doc/`, `crates/omnix-init/registry/`, `crates/omnix-cli/tests/`

## Common Issues & Workarounds

**Build failures**:
1. `cargo build` fails with env var errors → **Never use plain cargo**. Always use `nix build` or activate devShell first
2. "flake.lock not up to date" → Run `nix flake update` or use `--accept-flake-config`
3. Cache misses → Use `--accept-flake-config` (substituter: `https://cache.nixos.asia/oss` in flake.nix)

**Test failures**:
1. Nix-related test errors → Ensure in devShell (env vars required). Tests use `crates/omnix-cli/tests/flake.nix`
2. GitHub rate limits → Known issue (see `crates/omnix-cli/tests/flake.nix`). Use token or wait

**Dev environment**:
1. Direnv not activating → Check `direnv status`, run `direnv allow`, reload VSCode
2. Rust-analyzer issues → Ensure direnv VSCode extension active, check `.vscode/settings.json` (clippy enabled)
3. Pre-commit not running → Hooks: nixpkgs-fmt, rustfmt (config: `nix/modules/flake/pre-commit.nix`). Manual: `just pca`

## Development Workflow

1. Activate: `direnv allow` (or `nix develop`)
2. Edit Rust sources
3. Test locally: `just watch` (live reload)
4. Format: `just pca` (pre-commit)
5. Lint: `just clippy` (must pass, no warnings)
6. Test: `just cargo-test`
7. Build: `nix build --accept-flake-config`
8. CI: `just ci` (optional, recommended before push)
9. **Update `doc/history.md`** (required for all changes)

**Code requirements**: Clippy with `--deny warnings`, nixpkgs-fmt for Nix. Add tests in `crates/*/tests/` or `src/`. CLI integration tests in `crates/omnix-cli/tests/command/`. Use `#[tokio::test]` for async.

## Quick Reference

**Essential commands**: `just watch`, `just pca`, `just clippy`, `just cargo-test`, `nix build`, `just ci`

**Key files**: Main entry: `crates/omnix-cli/src/main.rs`, Commands: `crates/omnix-cli/src/command/*.rs`, Nix FFI: `crates/nix_rs/src/`, Env vars: `nix/envs/default.nix`, CI config: `om.yaml`, Docs: `doc/*.md`

**Debug**: `echo $OMNIX_SOURCE` (verify devShell), `just --list` (recipes), `nix --version` (>=2.16.0), `nix flake metadata .` (flake info)

**Trust these instructions**: Only search if info incomplete/incorrect. Nix is mandatory - no workarounds. Release: `cargo workspace publish` (see README.md). License: GPL-3.0.
