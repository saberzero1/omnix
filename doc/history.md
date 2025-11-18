---
order: 100
---

# Release history

## Unreleased

### Phase 3 Migration: Health & Init Packages (2025-11-18)

Added Go implementations of health checks, project initialization, and CLI framework:

- **pkg/health**: Comprehensive Nix environment health checks
  - Nix version validation (minimum version requirement)
  - Binary cache configuration (required, trusted, optional)
  - Direnv installation and shell integration
  - Experimental flakes feature detection
  - Homebrew interference detection (macOS)
  - Max-jobs configuration validation
  - Rosetta 2 check for x86_64 emulation (Apple Silicon)
  - Shell compatibility (bash, zsh, fish)
  - Trusted users configuration for caches
  - 79.4% test coverage with 26 test functions

- **pkg/init**: Project initialization with template support
  - Template action processing (copy, replace, retain)
  - Recursive directory copying with symlink preservation
  - Pattern-based string replacement in files and paths
  - File permission preservation
  - 19.8% test coverage with 19 test functions

- **pkg/cli**: Command line interface framework
  - Cobra-based CLI structure
  - `om health` command implementation
  - `om init` command implementation
  - Global flags and error handling
  - 32.7% test coverage with 9 test functions

**Code Metrics:**
- Total implementation: 1,707 LOC (27.8% reduction from Rust)
- Total tests: 1,508 LOC
- Documentation: ~500 LOC (comprehensive README files)
- All tests passing ✅

See [PHASE3_SUMMARY.md](../PHASE3_SUMMARY.md) for detailed implementation notes.

### Phase 4 Migration: CI & Develop Packages (2025-11-18)

Added Go implementations of CI/CD and development shell management:

- **pkg/ci**: Comprehensive CI/CD automation for Nix projects
  - Configuration parsing from om.yaml
  - Build, lockfile, and flake check steps
  - Custom command execution
  - GitHub Actions matrix generation
  - CI orchestration and results tracking
  - 52.8% test coverage with 47 test functions
  
- **pkg/develop**: Development shell management
  - Project configuration and setup
  - Pre-shell health checks (Nix version, Rosetta, max-jobs)
  - Post-shell README rendering
  - Cachix integration support
  - 50.0% test coverage with 37 test functions

**Code Metrics:**
- Total implementation: 1,017 LOC (68% reduction from Rust for CI)
- Total tests: 774 LOC
- Documentation: 351 LOC (comprehensive README files)
- All tests passing ✅

**Migration Status:**
- Phase 1 (Foundation): ✅ Complete
- Phase 2 (Nix Integration): ✅ Complete  
- Phase 3 (Health & Init): ✅ Complete
- Phase 4 (CI & Develop): ✅ Core Complete (CLI integration pending)
- Phase 5 (CLI Integration): 🔄 Next
- Phase 6 (GUI & Testing): Planned
- Phase 7 (Release): Planned

See [PHASE4_SUMMARY.md](../PHASE4_SUMMARY.md) for detailed implementation notes.

## 1.3.0 (2025-07-15) {#1.3.0}

- `om ci`: Allow impure builds through `impure = true;` setting in `om.yaml` (#445)
- `om health`
  - Fix DetSys installer hijacking its own version into `nix --version` causing false Nix version detection. (#458)
  - Add homebrew check (disabled by default) (#459)

## 1.0.3 (2025-03-17) {#1.0.3}

### Fixes

- `om ci`
  - Extra nix handling
      - Allow `--override-input` to work again (#439)
      - Support `--rebuild` by disallowing it in irrelevant subcommands (`eval`, `develop`, `run`, `flake {lock,check}`) (#441)
- `om init`
  - Handle symlinks *as is* (we expect relative symlink targets) without resolution (#443)

## 1.0.2 (2025-03-11) {#1.0.2}

### Fixes

- `om ci`
  - Prevent bad UTF-8 in build logs from crashing `om ci run` (#437)

## 1.0.1 (2025-03-10) {#1.0.1}

### Fixes

- `om init`
  - now copies over permissions as is (e.g.: respects executable bits on files) (#434)
  - applies replace in proper order so that directory rename doesn't skip content replace in its children  (#435)

### Chores

- Allow building on stable version of Rust (#427)
- Define ENVs in a single place and import them as default for all crates (#430)

## 1.0.0 (2025-02-17) {#1.0.0}

### Enhancements

- `om develop`: New command
- `om init`
  - Initial working version of `om init` command
- `om health`
  - Display Nix installer used (supports DetSys installer)
  - Display information in Markdown
  - Remove RAM/disk space checks, moving them to "information" section
  - Add shell check, to ensure its dotfiles are managed by Nix.
  - Add `--json` that returns the health check results as JSON
  - Switch from `nix-version.min-required` to more flexible `nix-version.supported`.
- `om ci`
  - Support for remote builds over SSH (via `--on` option)
  - Support for CI steps
    - Run `nix flake check` on all subflakes (#200)
    - Ability to add a custom CI step. For example, to run arbitrary commands.
  - Add `--accept-flake-config`
  - Add `--results=FILE` to store CI results as JSON in a file
  - Misc
    - Avoid running `nix-store` command multiple times (#224)
    - Locally cache `github:nix-systems` (to avoid Github API rate limit)

### Fixes

- `om ci run`: The `--override-input` option mandated `flake/` prefix (nixci legacy) which is no longer necessary in this release.
- `om health`: Use `whoami` to determine username which is more reliable than relying on `USER` environment variable

### Backward-incompatible changes

- `nix-health` and `nixci` flake output configurations are no longer supported.
- `om ci build` has been renamed to `om ci run`.

## 0.1.0 (2024-08-08) {#0.1.0}

Initial release of omnix.
