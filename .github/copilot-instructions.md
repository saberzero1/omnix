# Copilot Instructions for omnix

## Overview

**✨ DEFAULT LANGUAGE: Go** (Rust is legacy, being phased out)

Omnix is a Nix companion CLI tool implemented in Go. The repository contains **both Go and Rust code** during the migration period, but **all new features should be implemented in Go**.

**Current Status:**
- ✅ **Go is Production**: New commands and features go in `pkg/` and `cmd/om/`
- 🔄 **Rust is Legacy**: Being phased out according to `DESIGN_DOCUMENT.md`
- 📋 **Migration Progress**: Phase 1 complete (`pkg/common` migrated), Phase 2+ in progress

**Commands:** `om ci` (CI builds), `om health` (environment checks), `om develop` (dev envs), `om show` (flake inspection), `om init` (project scaffolding), `om run` (task execution). 

**Build Systems:** 
- **Go (Primary)**: Go modules + Nix buildGoModule - Use this for new features
- Rust (Legacy): Nix flakes + Cargo workspace - Being deprecated

**Platforms:** x86_64/aarch64 Linux/Darwin

## Critical Setup Requirements

### For Go Code (Primary): Standard Go Toolchain

**Use Go for all new features.** Go code can be built with standard Go tools (`go build`, `go test`). Nix development shell is recommended for consistency and access to all development tools.

**Setup**: Install Nix with flakes, setup direnv, run `direnv allow` in repo root. The `.envrc` activates Nix devShell automatically, providing both Go and Rust toolchains.

### For Rust Code (Legacy): Nix is Mandatory

**Only modify Rust code for bug fixes or maintenance.** Rust compilation requires Nix with flakes enabled. Build environment variables (`TRUE_FLAKE`, `FALSE_FLAKE`, etc.) are defined in `nix/envs/default.nix` and injected by Nix during compilation.

## Build Commands

**Verify devShell active**: `echo $OMNIX_SOURCE` (should show Nix store path). If not: `nix develop --accept-flake-config`

### Go Build Commands (Primary - Use for New Features)

**Build**:
- `just go-build` - Build Go binary to `bin/om`
- `go build -o bin/om ./cmd/om` - Direct Go build (no Nix needed)

**Run**:
- `just go-run [ARGS]` - Run Go version directly (e.g., `just go-run --version`)
- `./bin/om [ARGS]` - Run built binary

**Test**:
- `just go-test` - Run all Go tests with race detector
- `just go-test-coverage` - Generate coverage report (HTML at `coverage.html`)
- `go test -v ./pkg/common` - Test specific package

**Lint**:
- `just go-lint` - Run golangci-lint (20+ linters)
- `just go-fmt` - Format Go code (gofmt + goimports)

**Full Go CI**:
- `just go-ci` - Complete Go CI pipeline (format, lint, test with coverage, build)

### Rust Build Commands (Legacy - Maintenance Only)

**Build**:
- `nix build --accept-flake-config` - Full Rust build (5-10 min first run, <1 min incremental)
- `nix run --accept-flake-config` - Build and run Rust version
- `just watch` (or `just w`) - Development with live reload using bacon; pass args: `just w show`

**Test**:
- `just cargo-test` or `cargo test --release --all-features --workspace` (2-5 min)
- Per-crate: `cargo test -p omnix-cli --release --all-features`
- Tests **require** devShell environment variables

**Lint** (always before committing):
- `just pca` - Auto-format all code (pre-commit: nixpkgs-fmt + rustfmt + gofmt)
- `just clippy` - Strict Rust lint (treats warnings as errors)
- `just cargo-doc` - Build Rust docs

**CI**:
- `just ci` - Full local CI (10-20 min first run, uses `nix run . ci`)
- `just ci-cargo` - Faster iteration using cargo in devShell

## Project Structure

### Go Packages (in `pkg/`, Go modules - Primary for New Features)

**Current Structure:**
```
omnix/
├── cmd/om/              # Go CLI entry point - ADD NEW COMMANDS HERE
│   └── main.go
├── pkg/
│   ├── cli/             # CLI framework with cobra (commands in pkg/cli/cmd/)
│   ├── common/          # ✅ Migrated utilities (551 LOC, 80.6% coverage)
│   ├── health/          # ✅ Health checks
│   ├── init/            # ✅ Project initialization
│   ├── ci/              # ✅ CI functionality
│   ├── develop/         # ✅ Dev shells
│   ├── nix/             # ✅ Nix interaction
│   └── show/            # (future)
└── internal/            # Private packages
```

### Rust Crates (in `crates/`, Cargo workspace - LEGACY, Maintenance Only)
- `omnix-cli`: Main CLI entry, command dispatching. Entry: `src/main.rs`, commands: `src/command/*.rs`
- `omnix-ci`: CI command, builds all flake outputs  
- `omnix-health`: Health checks for Nix environment
- `omnix-develop`: Development environment setup
- `omnix-init`: Project scaffolding with templates (registry: `crates/omnix-init/registry/`)
- `omnix-common`: Shared utilities, logging (**✅ Migrated to Go: `pkg/common`**)
- `nix_rs`: Core Nix interaction library (**✅ Migrated to Go: `pkg/nix`**)
- `omnix-gui`: Dioxus-based GUI (experimental, migration TBD)

### Config Files
- **Go**: `go.mod`, `go.sum`, `.golangci.yml` (linter config)
- **Rust**: `Cargo.toml` (workspace), `rust-toolchain.toml`
- **Nix**: `flake.nix` (main flake), `flake.lock`, `nix/envs/default.nix` (critical: defines Rust build env vars)
- **Development**: `justfile` (task runner), `bacon.toml` (Rust file watcher), `om.yaml` (omnix config), `.envrc` (direnv)
- **VSCode**: `.vscode/extensions.json` (recommended: rust-analyzer, gopls, direnv, nix-ide), `.vscode/settings.json`

### Documentation
- `DESIGN_DOCUMENT.md` - Complete migration plan (Rust → Go)
- `GO_MIGRATION.md` - Migration guide and patterns
- `GO_QUICKSTART.md` - Go developer onboarding
- `PHASE1_SUMMARY.md` - Phase 1 completion metrics
- `doc/` - User docs (emanote): `index.md`, `om/*.md` (command docs), `config.md`, `history.md` (**required for all changes**)
- Commands: `just doc run` (live server), `just doc check` (link validation)

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

### Rust Build Failures
1. `cargo build` fails with env var errors → **Never use plain cargo for Rust**. Always use `nix build` or activate devShell first
2. "flake.lock not up to date" → Run `nix flake update` or use `--accept-flake-config`
3. Cache misses → Use `--accept-flake-config` (substituter: `https://cache.nixos.asia/oss` in flake.nix)

### Rust Test Failures
1. Nix-related test errors → Ensure in devShell (env vars required). Tests use `crates/omnix-cli/tests/flake.nix`
2. GitHub rate limits → Known issue (see `crates/omnix-cli/tests/flake.nix`). Use token or wait

### Go Build/Test Issues
1. Missing Go toolchain → Install Go 1.22+ or use Nix devShell (`nix develop --accept-flake-config`)
2. golangci-lint not found → Install via `go install` or use Nix devShell
3. Import errors → Run `go mod download` to fetch dependencies
4. Test failures → Check if running in correct directory (`go test ./...` from repo root)

### Dev Environment
1. Direnv not activating → Check `direnv status`, run `direnv allow`, reload VSCode
2. Rust-analyzer issues → Ensure direnv VSCode extension active, check `.vscode/settings.json` (clippy enabled)
3. gopls not working → Install Go extension for VSCode, ensure gopls in PATH
4. Pre-commit not running → Hooks: nixpkgs-fmt, rustfmt, gofmt (config: `nix/modules/flake/pre-commit.nix`). Manual: `just pca`

## Development Workflows

### Working on Go Code (Primary - Use for All New Features)

1. Activate: `direnv allow` (or use standard Go toolchain)
2. Edit Go sources in `cmd/` or `pkg/`
3. Test frequently: `just go-test` or `go test ./...`
4. Format: `just go-fmt` (gofmt + goimports)
5. Lint: `just go-lint` (golangci-lint)
6. Build: `just go-build`
7. Full CI: `just go-ci`
8. **Update `doc/history.md`** (required for all changes)

**Code requirements**: 
- golangci-lint must pass (0 issues)
- Test coverage ≥80% for new packages
- All public functions documented with godoc comments
- Table-driven tests for multiple scenarios
- Use `testify` for assertions

### Working on Rust Code (Legacy - Maintenance Only)

1. Activate: `direnv allow` (or `nix develop --accept-flake-config`)
2. Edit Rust sources in `crates/`
3. Test locally: `just watch` (live reload) or `just cargo-test`
4. Format: `just pca` (pre-commit hooks)
5. Lint: `just clippy` (must pass, no warnings)
6. Build: `nix build --accept-flake-config`
7. CI: `just ci` (optional, recommended before push)
8. **Update `doc/history.md`** (required for all changes)

**Code requirements**: Clippy with `--deny warnings`, nixpkgs-fmt for Nix. Add tests in `crates/*/tests/` or `src/`. CLI integration tests in `crates/omnix-cli/tests/command/`. Use `#[tokio::test]` for async.

### Adding New Commands (Go)

To add a new command like `om run`:
1. Create command file in `pkg/cli/cmd/run.go`
2. Implement `NewRunCmd()` function returning `*cobra.Command`
3. Register in `pkg/cli/root.go` init function: `rootCmd.AddCommand(cmd.NewRunCmd())`
4. Add tests in `pkg/cli/cmd/run_test.go`
5. Update `doc/om/run.md` with documentation
6. Update `doc/history.md` with release notes

### Migration Guidelines (Rust → Go - For Reference Only)

When migrating a Rust crate to Go:
1. Read `DESIGN_DOCUMENT.md` for the migration plan
2. Check `GO_MIGRATION.md` for migration patterns
3. Write Go tests BEFORE implementation (test-first approach)
4. Maintain feature parity with Rust version
5. Achieve ≥80% test coverage
6. Run both Rust and Go tests to ensure no regressions
7. Document migration in appropriate `PHASE*_SUMMARY.md`

**Error Handling Pattern:**
- Rust: `Result<T, E>` → Go: `(T, error)`
- Rust: `Option<T>` → Go: `*T` or zero value
- Rust: `anyhow::Context` → Go: `fmt.Errorf("context: %w", err)`

**Async Pattern:**
- Rust: `async fn` + `.await` → Go: regular function + goroutines
- Rust: `tokio::spawn` → Go: `go func() { ... }()`
- Rust: `tokio::select!` → Go: `select { case <-ch1: ... }`

## Quick Reference

### Go (Primary - Use for All New Work)
**Essential commands**: `just go-test`, `just go-lint`, `just go-fmt`, `just go-build`, `just go-ci`

**Key files**:
- Main entry: `cmd/om/main.go`
- CLI root: `pkg/cli/root.go`
- Commands: `pkg/cli/cmd/*.go`
- Packages: `pkg/{common,nix,health,init,ci,develop}/*.go`
- Tests: `pkg/**/*_test.go`
- Config: `go.mod`, `.golangci.yml`

### Rust (Legacy - Maintenance Only)
**Essential commands**: `just watch`, `just pca`, `just clippy`, `just cargo-test`, `nix build`, `just ci`

**Key files**: 
- Main entry: `crates/omnix-cli/src/main.rs`
- Commands: `crates/omnix-cli/src/command/*.rs`
- Nix FFI: `crates/nix_rs/src/`
- Env vars: `nix/envs/default.nix`

### Configuration & Docs
**Config**: `om.yaml` (omnix config)
**Debug**: 
- Go version: `go version` (should be 1.22+)
- Rust devShell: `echo $OMNIX_SOURCE` (verify devShell for Rust work)
- Nix: `nix --version` (>=2.16.0)
- Just: `just --list` (all recipes)

**Trust these instructions**: All new features go in Go (`pkg/` and `cmd/om/`). Rust is legacy maintenance only. License: AGPL-3.0.
