---
order: 100
---

# Release history

## Unreleased

### Phase 6 Migration: GUI & Testing (2025-11-18) ✅ **COMPLETE**

**80% Coverage Goal EXCEEDED - Achieved 81.0%!**

Comprehensive testing improvements and GUI migration decision:

- **GUI Decision** ✅ **CONFIRMED**:
  - Analyzed three migration options (Hybrid Rust, Migrate to Go, Remove temporarily)
  - **Decision**: Remove experimental GUI from v2.0, re-evaluate post-release
  - **User Confirmation**: Repository owner confirmed removal is preferred (multiple confirmations)
  - **Rationale**: <5% user adoption, tool is CLI-first, GUI goes basically unused
  - **Future**: Re-evaluate post-2.0 with modern alternatives (TUI, web dashboard)
  - Documented in PHASE6_SUMMARY.md with detailed analysis

- **Test Coverage** ✅ **GOAL EXCEEDED**:
  - Overall coverage: 72.6% → **81.0%** (+8.4 pp, **101% of 80% goal**)
  - pkg/cli: 33.3% → 57.1% (+23.8 pp)
  - pkg/cli/cmd: 36.1% → **72.7%** (+36.6 pp)
  - **8/10 packages exceed 80% coverage**
  - Added comprehensive command execution tests
  - All RunE functions now tested with real command execution

- **Major Coverage Improvements**:
  - `runHealth`: 0.0% → 85.7% (+85.7pp)
  - `newCIRunCmd`: 23.1% → 82.7% (+59.6pp)
  - `newCIGHMatrixCmd`: 31.2% → 87.5% (+56.3pp)
  - `NewDevelopCmd`: 15.4% → 84.6% (+69.2pp)

- **New Tests**:
  - `cmd/om/main_test.go`: Version variable tests
  - `pkg/cli/cli_test.go`: Enhanced with help, version, and verbosity tests
  - `pkg/cli/cmd/cmd_test.go`: Structure and flag tests
  - `pkg/cli/cmd/integration_test.go`: Integration tests for commands
  - `pkg/cli/cmd/ci_develop_test.go`: Help execution and comprehensive flag validation tests
  - `pkg/cli/cmd/error_paths_test.go`: Error path and argument validation tests
  - `pkg/cli/cmd/rune_execution_test.go`: **RunE execution tests for all major commands**

- **Testing Infrastructure**:
  - Documented full vs. short mode testing strategy
  - Created table-driven test patterns
  - Improved test organization and removed duplicates
  - Integration tests properly gated behind -short flag
  - Comprehensive flag validation for all commands
  - **RunE execution tests with real Nix environment**
  - Error path testing for robustness
  - **Cross-platform CI matrix** (Linux x86_64, macOS x86_64/ARM64)
  - All tests passing ✅ (100% pass rate)

- **CI/CD Automation** ✅:
  - Created `.github/workflows/go-test.yaml` for cross-platform testing
  - **Test matrix**: 3 platforms × 2 Go versions = 6 test configurations
  - **Platforms**: Ubuntu (x86_64), macOS 13 (x86_64), macOS latest (ARM64)
  - **Go versions**: 1.22, 1.23
  - Parallel execution with fail-fast disabled
  - Coverage artifacts per platform
  - Integration tests with Nix on Linux
  - Automated linting on Linux and macOS

**Code Metrics:**
- New test code: **845+ LOC**
- Coverage improvement: **+8.4 percentage points** (72.6% → 81.0%)
- Packages meeting 80%+ target: **8/10** ✅
- All tests passing (100% pass rate)
- Integration tests: 7/7 packages
- RunE execution tests: 10 comprehensive test cases

**Coverage Achievement:**
- **Goal**: 80%
- **Achieved**: 81.0%
- **Goal completion**: 101% ✅

**Migration Status:**
- Phase 1 (Foundation): ✅ Complete
- Phase 2 (Nix Integration): ✅ Complete  
- Phase 3 (Health & Init): ✅ Complete
- Phase 4 (CI & Develop): ✅ Complete
- Phase 5 (CLI Integration): ✅ Complete
- **Phase 6 (GUI & Testing): ✅ COMPLETE** (GUI confirmed, 81% coverage achieved)
- Phase 7 (Release): 📋 Pending

### Phase 5 Migration: CLI Integration (2025-11-18)

Completed CLI integration with all remaining commands and features:

- **om show**: Flake output inspection
  - Display packages, devShells, apps, checks, and configurations
  - Colorful table output with usage hints
  - System-specific output filtering
  - Support for templates, schemas, overlays, and NixOS/Darwin configurations
  
- **om completion**: Shell completion generation
  - Support for bash, zsh, fish, and PowerShell
  - Detailed installation instructions per shell
  - Native Cobra completion integration

- **Enhanced CLI features**:
  - Version information with git commit display
  - Logging verbosity control (`--verbose` / `-v` flag with 5 levels)
  - Proper error handling and log flushing
  - Global logging setup via PersistentPreRunE hook

- **pkg/nix enhancements**:
  - FlakeOutputs type for representing flake outputs
  - FlakeMetadata type for flake information
  - FlakeShow method for retrieving flake data
  - Custom JSON unmarshaling for flake structures
  - Path-based output lookup (GetByPath)
  - Terminal value extraction (GetAttrsetOfVal)

**Code Metrics:**
- New implementation: ~300 LOC (show + completion commands)
- Nix package additions: ~160 LOC (FlakeOutputs support)
- Tests: ~170 LOC
- All tests passing ✅
- Zero regressions from previous phases
- Phase 6 (GUI & Testing): 🔄 Next
- Phase 7 (Release): Planned

See [PHASE5_SUMMARY.md](../PHASE5_SUMMARY.md) for detailed implementation notes.

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
  - 84.6% test coverage with 47 test functions
  
- **pkg/develop**: Development shell management
  - Project configuration and setup
  - Pre-shell health checks (Nix version, Rosetta, max-jobs)
  - Post-shell README rendering
  - Cachix integration support
  - 82.4% test coverage with 37 test functions

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
