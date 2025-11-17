# Copilot Instructions for omnix

## Overview

**⚠️ PROJECT IN TRANSITION: Rust → Go Migration (Phase 1 Complete)**

Omnix is a Nix companion CLI tool currently being migrated from Rust to Go according to `DESIGN_DOCUMENT.md`. The repository contains **both Rust and Go code** during this transition period.

**Current Status:**
- ✅ **Phase 1 Complete**: Go foundation established, `pkg/common` migrated (80.6% test coverage)
- 🔄 **Active**: Both Rust (8 crates, 117 .rs files) and Go (11 .go files) implementations coexist
- 📋 **Next**: Phase 2 - Core Nix Integration (`nix_rs` → `pkg/nix`)

**Commands:** `om ci` (CI builds), `om health` (environment checks), `om develop` (dev envs), `om show` (flake inspection), `om init` (project scaffolding). 

**Build Systems:** 
- Rust: Nix flakes + Cargo workspace (production, being phased out)
- Go: Go modules + Nix buildGoModule (in development, future production)

**Platforms:** x86_64/aarch64 Linux/Darwin

## Critical Setup Requirements

### For Rust Code: Nix is Mandatory

**DO NOT use plain `cargo build/test`** for Rust code - it will fail. Rust compilation requires Nix with flakes enabled. Build environment variables (`TRUE_FLAKE`, `FALSE_FLAKE`, `FLAKE_METADATA`, `DEFAULT_FLAKE_SCHEMAS`, `INSPECT_FLAKE`, `NIX_SYSTEMS`, etc.) are defined in `nix/envs/default.nix` and injected by Nix during compilation.

### For Go Code: Standard Go Toolchain

Go code can be built with standard Go tools (`go build`, `go test`) but Nix development shell is recommended for consistency and access to all development tools.

**Setup**: Install Nix with flakes, setup direnv, run `direnv allow` in repo root. The `.envrc` activates Nix devShell automatically, providing both Rust and Go toolchains.

## Build Commands

**Verify devShell active**: `echo $OMNIX_SOURCE` (should show Nix store path). If not: `nix develop --accept-flake-config`

### Rust Build Commands (Production - Requires Nix DevShell)

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

### Go Build Commands (Development - Phase 1 Complete)

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

## Project Structure

### Rust Crates (in `crates/`, Cargo workspace - being phased out)
- `omnix-cli`: Main CLI entry, command dispatching. Entry: `src/main.rs`, commands: `src/command/*.rs`
- `omnix-ci`: CI command, builds all flake outputs
- `omnix-health`: Health checks for Nix environment
- `omnix-develop`: Development environment setup
- `omnix-init`: Project scaffolding with templates (registry: `crates/omnix-init/registry/`)
- `omnix-common`: Shared utilities, logging (**✅ Migrated to Go: `pkg/common`**)
- `nix_rs`: Core Nix interaction library. Key: `src/flake/` (URL parsing, metadata, schema), `src/flake/functions/` (Rust+Nix FFI) (**🔄 Next migration target: `pkg/nix`**)
- `omnix-gui`: Dioxus-based GUI (experimental, migration TBD)

### Go Packages (in `pkg/`, Go modules - Phase 1 complete)

**Current Structure:**
```
omnix/
├── cmd/om/              # Go CLI entry point (placeholder)
│   └── main.go
├── pkg/common/          # ✅ MIGRATED from omnix-common (551 LOC, 80.6% coverage)
│   ├── logging.go       # Structured logging (zap)
│   ├── check.go         # Binary existence checks
│   ├── fs.go            # Filesystem utilities
│   ├── config.go        # Config parsing (JSON/YAML)
│   ├── markdown.go      # Markdown rendering (glamour)
│   └── *_test.go        # Comprehensive test suite
└── internal/            # Private packages (future)
```

**Planned Packages (per DESIGN_DOCUMENT.md):**
- `pkg/nix/` - Nix interaction (replaces `nix_rs`) - **Phase 2**
- `pkg/health/` - Health checks (replaces `omnix-health`) - **Phase 3**
- `pkg/init/` - Project initialization (replaces `omnix-init`) - **Phase 3**
- `pkg/ci/` - CI functionality (replaces `omnix-ci`) - **Phase 4**
- `pkg/develop/` - Dev shells (replaces `omnix-develop`) - **Phase 4**
- `pkg/cli/` - CLI framework - **Phase 5**

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

### Working on Rust Code (Production)

1. Activate: `direnv allow` (or `nix develop --accept-flake-config`)
2. Edit Rust sources in `crates/`
3. Test locally: `just watch` (live reload) or `just cargo-test`
4. Format: `just pca` (pre-commit hooks)
5. Lint: `just clippy` (must pass, no warnings)
6. Build: `nix build --accept-flake-config`
7. CI: `just ci` (optional, recommended before push)
8. **Update `doc/history.md`** (required for all changes)

**Code requirements**: Clippy with `--deny warnings`, nixpkgs-fmt for Nix. Add tests in `crates/*/tests/` or `src/`. CLI integration tests in `crates/omnix-cli/tests/command/`. Use `#[tokio::test]` for async.

### Working on Go Code (Migration in Progress)

1. Activate: `direnv allow` (or use standard Go toolchain)
2. Edit Go sources in `cmd/` or `pkg/`
3. Test frequently: `just go-test` or `go test ./...`
4. Format: `just go-fmt` (gofmt + goimports)
5. Lint: `just go-lint` (golangci-lint)
6. Build: `just go-build`
7. Full CI: `just go-ci`
8. **Update migration docs**: `GO_MIGRATION.md`, `PHASE*_SUMMARY.md` as appropriate
9. **Update `doc/history.md`** (required for all changes)

**Code requirements**: 
- golangci-lint must pass (0 issues)
- Test coverage ≥80% for new packages
- All public functions documented with godoc comments
- Table-driven tests for multiple scenarios
- Use `testify` for assertions

### Migration Guidelines (Rust → Go)

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

### Rust (Production)
**Essential commands**: `just watch`, `just pca`, `just clippy`, `just cargo-test`, `nix build`, `just ci`

**Key files**: 
- Main entry: `crates/omnix-cli/src/main.rs`
- Commands: `crates/omnix-cli/src/command/*.rs`
- Nix FFI: `crates/nix_rs/src/`
- Env vars: `nix/envs/default.nix`

### Go (Development - Phase 1)
**Essential commands**: `just go-test`, `just go-lint`, `just go-fmt`, `just go-build`, `just go-ci`

**Key files**:
- Main entry: `cmd/om/main.go` (placeholder)
- Common utilities: `pkg/common/*.go`
- Tests: `pkg/common/*_test.go`
- Config: `go.mod`, `.golangci.yml`

### Both
**Config**: `om.yaml` (omnix config)
**Docs**: `doc/*.md`, `DESIGN_DOCUMENT.md`, `GO_MIGRATION.md`
**Debug**: 
- Rust devShell: `echo $OMNIX_SOURCE` (verify devShell)
- Go version: `go version`
- Nix: `nix --version` (>=2.16.0)
- Just: `just --list` (all recipes)

### Migration Status Tracking
- **Completed**: `pkg/common` (✅ Phase 1)
- **In Progress**: None (awaiting Phase 2 start)
- **Next**: `pkg/nix` (replaces `nix_rs` crate)
- **Future**: `pkg/health`, `pkg/init`, `pkg/ci`, `pkg/develop`, `pkg/cli`

See `DESIGN_DOCUMENT.md` Section 3.2 for detailed phase breakdown.

**Trust these instructions**: Only search if info incomplete/incorrect. For Rust: Nix is mandatory - no workarounds. For Go: standard toolchain works but Nix devShell recommended. License: AGPL-3.0.
