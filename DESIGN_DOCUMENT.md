# Omnix Rust to Go Rewrite - Design Document

## Executive Summary

This design document outlines the comprehensive strategy for rewriting the Omnix project from Rust to Go. Omnix is a developer-friendly companion tool for Nix, consisting of approximately 9,300 lines of Rust code organized into 8 crates with extensive Nix integration through flakes and modules.

### Goals

1. **Maintain Feature Parity**: Ensure all existing functionality is preserved during the rewrite
2. **Improve Developer Experience**: Leverage Go's simplicity and faster compilation times
3. **Preserve Nix Integration**: Maintain seamless integration with Nix ecosystem
4. **Ensure Quality**: Achieve equivalent or better test coverage and reliability
5. **Minimize Disruption**: Provide a smooth transition path for users and contributors

### Key Metrics

- **Current Codebase**: ~9,300 lines of Rust across 8 crates
- **Estimated Go Codebase**: ~10,000-12,000 lines (accounting for Go's verbosity)
- **Timeline**: 3-6 months phased migration
- **Test Coverage Target**: ≥ current coverage (unit + integration tests)

---

## 1. Current Architecture Analysis

### 1.1 Crate Structure

The current Rust implementation consists of 8 distinct crates:

| Crate | Files | Purpose | Key Dependencies |
|-------|-------|---------|------------------|
| `omnix-cli` | 11 | Main CLI entry point, command routing | clap, tokio, human-panic |
| `omnix-ci` | 24 | CI/CD automation for Nix projects | GitHub Actions integration |
| `omnix-health` | 14 | System health checks for Nix environment | nix_rs, serde |
| `omnix-init` | 8 | Project initialization and templating | - |
| `omnix-develop` | 4 | Development shell management | direnv integration |
| `omnix-common` | 6 | Shared utilities (logging, markdown, config) | tracing, pulldown-cmark |
| `nix_rs` | 29 | Core Nix interaction library | Low-level Nix command execution |
| `omnix-gui` | 12 | GUI components (WASM-based) | fermi, dioxus (planned) |

### 1.2 Key Technical Characteristics

**Rust-Specific Features:**
- Async/await using Tokio runtime (primarily in omnix-health, omnix-ci)
- Strong type system with algebraic data types (enums with associated data)
- Cargo workspace for multi-crate management
- Zero-cost abstractions and ownership model
- WASM compilation support for GUI components

**External Integrations:**
- **Nix**: Command-line execution via `nix` CLI
- **GitHub**: API integration for CI/CD workflows
- **Direnv**: Shell environment management
- **DetSys Installer**: Nix installation detection

**Build System:**
- Cargo for Rust compilation
- Crane (via rust-flake) for Nix builds
- Custom `crate.nix` per crate for Nix integration
- Flake-parts modular Nix configuration

### 1.3 Command Structure

Current commands implemented:
- `om show` - Display flake information
- `om ci run` - Run CI for Nix projects
- `om ci gh-matrix` - Generate GitHub Actions matrix
- `om health` - Check Nix environment health
- `om init` - Initialize new Nix projects
- `om develop` - Manage development shells
- `om completion` - Generate shell completions

### 1.4 Dependencies Analysis

**Critical Rust Dependencies:**
- `clap` (v4.3) - CLI parsing → Go equivalent: `cobra` or `urfave/cli`
- `tokio` (v1.33) - Async runtime → Go: native goroutines
- `serde` (v1.0) - Serialization → Go: standard `encoding/json`, `gopkg.in/yaml.v3`
- `anyhow` (v1.0) - Error handling → Go: standard `error` + custom types
- `tracing` (v0.1) - Structured logging → Go: `zap` or `logrus`
- `reqwest` (v0.11) - HTTP client → Go: `net/http` or `resty`

---

## 2. Go Architecture Design

### 2.1 Package Structure

Proposed Go module organization mirroring Rust crate structure:

```
omnix/
├── go.mod                          # Root module
├── go.sum
├── cmd/
│   └── om/
│       └── main.go                 # CLI entry point
├── pkg/                            # Public packages
│   ├── cli/                        # CLI framework (replaces omnix-cli)
│   │   ├── cmd/
│   │   │   ├── root.go
│   │   │   ├── show.go
│   │   │   ├── ci.go
│   │   │   ├── health.go
│   │   │   ├── init.go
│   │   │   ├── develop.go
│   │   │   └── completion.go
│   │   └── version.go
│   ├── ci/                         # CI functionality (replaces omnix-ci)
│   │   ├── runner.go
│   │   ├── config.go
│   │   ├── steps/
│   │   │   ├── build.go
│   │   │   ├── lockfile.go
│   │   │   ├── check.go
│   │   │   └── custom.go
│   │   ├── github/
│   │   │   ├── matrix.go
│   │   │   ├── actions.go
│   │   │   └── pullrequest.go
│   │   └── flake.go
│   ├── health/                     # Health checks (replaces omnix-health)
│   │   ├── checks/
│   │   │   ├── nix.go
│   │   │   ├── cache.go
│   │   │   ├── direnv.go
│   │   │   ├── homebrew.go
│   │   │   └── shell.go
│   │   ├── report.go
│   │   └── output.go
│   ├── init/                       # Project initialization (replaces omnix-init)
│   │   ├── template.go
│   │   ├── registry.go
│   │   └── replace.go
│   ├── develop/                    # Dev shell (replaces omnix-develop)
│   │   ├── shell.go
│   │   └── direnv.go
│   ├── nix/                        # Nix interaction (replaces nix_rs)
│   │   ├── command.go
│   │   ├── flake.go
│   │   ├── store.go
│   │   ├── config.go
│   │   ├── copy.go
│   │   ├── env.go
│   │   ├── version.go
│   │   ├── info.go
│   │   └── installer.go
│   └── common/                     # Shared utilities (replaces omnix-common)
│       ├── logging.go
│       ├── markdown.go
│       ├── config.go
│       ├── fs.go
│       └── check.go
├── internal/                       # Private packages
│   ├── testutil/                   # Test utilities
│   └── testdata/                   # Test fixtures
└── web/                            # GUI components (replaces omnix-gui)
    └── wasm/
        └── main.go                 # WASM entry point
```

### 2.2 Go-Specific Design Patterns

**Error Handling:**
```go
// Custom error types for domain-specific errors
type NixError struct {
    Command string
    ExitCode int
    Stderr string
    Err error
}

func (e *NixError) Error() string {
    return fmt.Sprintf("nix command failed: %s (exit %d): %v", e.Command, e.ExitCode, e.Err)
}

func (e *NixError) Unwrap() error {
    return e.Err
}
```

**Concurrency:**
```go
// Use channels and goroutines instead of async/await
func RunCISteps(ctx context.Context, steps []Step) error {
    errChan := make(chan error, len(steps))
    
    for _, step := range steps {
        go func(s Step) {
            errChan <- s.Execute(ctx)
        }(step)
    }
    
    // Collect results
    for i := 0; i < len(steps); i++ {
        if err := <-errChan; err != nil {
            return err
        }
    }
    return nil
}
```

**Configuration:**
```go
// Use struct tags for YAML/JSON parsing
type CIConfig struct {
    CI struct {
        Default map[string]SubflakeConfig `yaml:"default" json:"default"`
    } `yaml:"ci" json:"ci"`
    Health HealthConfig `yaml:"health" json:"health"`
    Develop DevelopConfig `yaml:"develop" json:"develop"`
}
```

### 2.3 Dependency Selection

| Rust Crate | Purpose | Go Package | Rationale |
|------------|---------|------------|-----------|
| clap | CLI parsing | cobra | Industry standard, feature-rich, well-maintained |
| tokio | Async runtime | (native) | Go's goroutines and channels are built-in |
| serde + serde_json | Serialization | encoding/json | Standard library is sufficient |
| serde_yaml | YAML parsing | gopkg.in/yaml.v3 | Most popular Go YAML library |
| anyhow + thiserror | Error handling | (custom types) | Go's error interface + wrapping |
| tracing | Logging | go.uber.org/zap | High-performance structured logging |
| reqwest | HTTP client | net/http | Standard library sufficient, or resty for convenience |
| colored | Terminal colors | fatih/color | Popular, simple API |
| tabled | Table formatting | olekukonko/tablewriter | Well-established |
| glob | Pattern matching | github.com/gobwas/glob | Fast implementation |
| pulldown-cmark | Markdown | github.com/gomarkdown/markdown | Active, feature-complete |
| tempfile | Temp files | os.CreateTemp | Standard library |
| which | Find executables | (custom impl) | Simple to implement in Go |

**Additional Go Dependencies:**
- `github.com/spf13/viper` - Configuration management
- `github.com/stretchr/testify` - Testing assertions
- `github.com/google/go-cmp` - Deep equality testing

### 2.4 Nix Integration Strategy

**Current Rust Approach:**
- Uses Crane via rust-flake for Nix builds
- Custom `crate.nix` per crate
- flake-parts modular configuration
- Compiled as static binary

**Go Nix Integration:**
```nix
# flake.nix excerpt
{
  outputs = { self, nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = self.packages.${system}.omnix;
          omnix = pkgs.buildGoModule {
            pname = "omnix";
            version = "2.0.0";
            src = ./.;
            
            vendorHash = "sha256-..."; # Will be computed
            
            # CGO disabled for static linking
            CGO_ENABLED = 0;
            
            ldflags = [
              "-s" "-w"
              "-X main.Version=${version}"
              "-X main.Commit=${self.rev or "dev"}"
            ];
            
            # Shell completions
            postInstall = ''
              installShellCompletion --cmd om \
                --bash <($out/bin/om completion bash) \
                --zsh <($out/bin/om completion zsh) \
                --fish <($out/bin/om completion fish)
            '';
            
            meta = with pkgs.lib; {
              description = "Developer-friendly companion for Nix";
              homepage = "https://omnix.page";
              license = licenses.agpl3Only;
              maintainers = [ ];
            };
          };
        }
      );
      
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              go-tools
              golangci-lint
              just
              nix
            ];
            
            shellHook = ''
              echo "🍾 Welcome to the omnix development shell"
              echo "Run 'just' to see available commands"
            '';
          };
        }
      );
    };
}
```

**Build Process:**
1. `go mod download` - Download dependencies
2. `go mod vendor` - Vendor dependencies for Nix
3. `nix hash path vendor` - Compute vendorHash
4. `nix build` - Build static binary with buildGoModule

---

## 3. Migration Strategy

### 3.1 Migration Principles

1. **Incremental Migration**: Rewrite one package at a time, maintaining compatibility
2. **Test-First**: Write Go tests before implementation to ensure behavior preservation
3. **Parallel Development**: Keep Rust version functional during Go development
4. **Feature Freeze**: No new Rust features during migration period
5. **Documentation**: Update docs progressively as packages migrate

### 3.2 Migration Phases

#### Phase 1: Foundation (Weeks 1-4)

**Objectives:**
- Set up Go module structure
- Establish build and CI infrastructure
- Migrate core utilities

**Tasks:**
1. Initialize Go module and directory structure
2. Set up Go development environment in Nix
3. Configure golangci-lint, go-fmt, go-vet
4. Update CI workflows for Go (GitHub Actions)
5. **Migrate `pkg/common`** (replaces omnix-common)
   - logging.go - Port tracing to zap
   - markdown.go - Port pulldown-cmark to gomarkdown
   - config.go - Port serde YAML to yaml.v3
   - fs.go - Port filesystem utilities
   - check.go - Port check utilities
6. Write comprehensive unit tests for common package
7. Document migration patterns and conventions

**Success Criteria:**
- [ ] Go module builds successfully
- [ ] All common package tests passing
- [ ] CI pipeline validates Go code
- [ ] Developer guide for Go migration created

#### Phase 2: Core Nix Integration (Weeks 5-8)

**Objectives:**
- Establish Nix command interaction layer
- Ensure feature parity with nix_rs crate

**Tasks:**
1. **Migrate `pkg/nix`** (replaces nix_rs)
   - command.go - Nix command execution
   - flake.go - Flake operations (show, check, lock)
   - store.go - Store path operations
   - config.go - Nix configuration parsing
   - copy.go - Nix copy operations
   - env.go - Environment detection
   - version.go - Version parsing
   - info.go - System info
   - installer.go - Installer detection (DetSys, etc.)
2. Create integration tests with actual Nix commands
3. Test on all supported platforms (Linux, macOS, x86_64, aarch64)
4. Benchmark performance vs Rust implementation
5. Document Nix integration API

**Success Criteria:**
- [ ] All nix package functions implemented
- [ ] Integration tests passing on all platforms
- [ ] Performance within 10% of Rust version
- [ ] No regressions in Nix interaction

#### Phase 3: Health & Init (Weeks 9-11)

**Objectives:**
- Migrate user-facing features with clear boundaries
- Validate CLI framework choice

**Tasks:**
1. **Migrate `pkg/health`** (replaces omnix-health)
   - Implement all health check types
   - Port async check execution to goroutines
   - Implement JSON and Markdown output
   - Test check accuracy and reliability
2. **Migrate `pkg/init`** (replaces omnix-init)
   - Template management
   - Registry handling
   - String replacement logic
   - Symlink preservation
   - Permission handling
3. Set up Cobra CLI framework in `pkg/cli`
4. Implement `om health` command in Go
5. Implement `om init` command in Go
6. End-to-end testing of both commands
7. User acceptance testing with real projects

**Success Criteria:**
- [ ] `om health` produces identical output
- [ ] `om init` creates identical project structures
- [ ] All edge cases handled (symlinks, permissions)
- [ ] CLI experience matches Rust version

#### Phase 4: CI & Develop (Weeks 12-15)

**Objectives:**
- Migrate complex business logic
- Ensure GitHub integration works correctly

**Tasks:**
1. **Migrate `pkg/ci`** (replaces omnix-ci)
   - CI runner logic
   - Configuration parsing
   - Build, lockfile, check steps
   - Custom step execution
   - GitHub matrix generation
   - GitHub Actions integration
   - Pull request handling
   - Remote build support (SSH)
   - Results JSON output
2. **Migrate `pkg/develop`** (replaces omnix-develop)
   - Development shell management
   - Direnv integration
   - Shell detection
3. Implement `om ci run` command
4. Implement `om ci gh-matrix` command
5. Implement `om develop` command
6. Test with multiple real-world Nix projects
7. Validate GitHub Actions integration

**Success Criteria:**
- [ ] CI runs produce identical results
- [ ] GitHub matrix generation matches
- [ ] Remote builds work over SSH
- [ ] Develop shell activation works
- [ ] All CI steps execute correctly

#### Phase 5: CLI Integration (Weeks 16-18)

**Objectives:**
- Complete CLI migration
- Ensure all commands work together

**Tasks:**
1. **Complete `cmd/om` and `pkg/cli`**
   - Wire all commands
   - Implement completion generation
   - Add version/help information
   - Port human-panic behavior
   - Configure logging verbosity
2. Implement `om show` command
3. Implement remaining commands
4. Create shell completions (bash, zsh, fish)
5. End-to-end integration testing
6. Performance benchmarking
7. Memory profiling
8. Update all documentation

**Success Criteria:**
- [ ] All commands functional in Go
- [ ] Shell completions generated
- [ ] Help text matches Rust version
- [ ] Performance acceptable
- [ ] Memory usage reasonable

#### Phase 6: GUI & Testing (Weeks 19-21)

**Objectives:**
- Address WASM/GUI components
- Comprehensive testing

**Tasks:**
1. Evaluate GUI migration options:
   - Option A: Keep Rust WASM for GUI (hybrid approach)
   - Option B: Migrate to Go WASM (experimental)
   - Option C: Remove GUI temporarily, re-implement later
2. **If migrating GUI:**
   - Research Go WASM support
   - Port omnix-gui to Go WASM
   - Test in browsers
3. Comprehensive test suite:
   - Increase unit test coverage to 80%+
   - Add integration tests for all commands
   - Add regression tests for known bugs
   - Property-based testing where applicable
4. Cross-platform testing
5. Create test fixtures and test data
6. Update CI to run full test suite

**Success Criteria:**
- [ ] GUI decision made and documented
- [ ] Test coverage ≥ 80%
- [ ] All integration tests passing
- [ ] Tests run in CI successfully

#### Phase 7: Release & Migration (Weeks 22-24)

**Objectives:**
- Prepare for production release
- Smooth user migration

**Tasks:**
1. Update Nix flake.nix with Go build
2. Remove Rust build infrastructure (keep as v1 branch)
3. Update all documentation:
   - README.md
   - Contributing guide
   - Website docs
   - API documentation (godoc)
4. Create migration guide for users
5. Update history.md with 2.0.0 release notes
6. Beta release for community testing
7. Address beta feedback
8. Final release v2.0.0
9. Update package managers (nixpkgs, homebrew, etc.)
10. Announcement and communication

**Success Criteria:**
- [ ] Go version packaged in Nix
- [ ] All documentation updated
- [ ] Migration guide published
- [ ] Beta testing completed
- [ ] v2.0.0 released successfully
- [ ] Community informed

### 3.3 Parallel Rust Maintenance

During migration (Phases 1-6):
- **Bug Fixes**: Critical bugs in Rust version get backported
- **Security**: Security issues are fixed in both versions
- **Features**: Feature freeze on Rust, new features go to Go
- **Releases**: Continue Rust releases as 1.x.y versions

After migration (Phase 7):
- Rust version remains on v1 branch for reference
- No new Rust releases unless critical security issues
- All development focuses on Go version (2.x.y)

### 3.4 Rollback Strategy

If critical issues arise during migration:
1. **Immediate Rollback**: Revert to latest Rust release
2. **Issue Analysis**: Identify root cause in Go implementation
3. **Fix Forward**: Patch Go version if possible
4. **Pause Migration**: Pause next phase if needed
5. **Communication**: Inform users of status

---

## 4. Testing Strategy

### 4.1 Test Categories

**Unit Tests:**
- Test individual functions and methods
- Mock external dependencies (Nix commands, file system)
- Target: 80%+ code coverage
- Use `testify` for assertions

**Integration Tests:**
- Test interactions between packages
- Use real Nix commands in isolated environments
- Test actual file operations in temp directories
- Validate JSON/YAML parsing with real configs

**End-to-End Tests:**
- Test complete command workflows
- Run `om` CLI with various arguments
- Validate output format and exit codes
- Use `assert_cmd` equivalent in Go

**Regression Tests:**
- Test for previously fixed bugs
- Maintain test suite from Rust version
- Add tests for any new bugs found during migration

**Performance Tests:**
- Benchmark critical operations
- Compare against Rust baseline
- Ensure no significant regressions
- Profile memory usage

**Platform Tests:**
- Test on all supported platforms:
  - x86_64-linux (Ubuntu, NixOS)
  - aarch64-linux (ARM servers)
  - x86_64-darwin (Intel Mac)
  - aarch64-darwin (Apple Silicon)
- Use CI matrix for cross-platform validation

### 4.2 Test Organization

```
pkg/
├── common/
│   ├── logging.go
│   ├── logging_test.go
│   ├── markdown.go
│   └── markdown_test.go
├── nix/
│   ├── command.go
│   ├── command_test.go
│   ├── integration_test.go      # Integration tests
│   └── testdata/
│       ├── flake.nix
│       └── expected_output.json
└── ci/
    ├── runner.go
    ├── runner_test.go
    ├── e2e_test.go              # End-to-end tests
    └── testdata/
        ├── om.yaml
        └── test_flake/
```

### 4.3 Testing Tools

- `testing` - Standard Go testing framework
- `github.com/stretchr/testify` - Rich assertion library
- `github.com/google/go-cmp` - Deep comparison
- `github.com/golang/mock` - Mock generation (if needed)
- `github.com/google/pprof` - Performance profiling
- `golangci-lint` - Linting and static analysis

### 4.4 Test Automation

**Pre-commit Hooks:**
- Run `go fmt` (format code)
- Run `go vet` (static analysis)
- Run `golangci-lint` (comprehensive linting)
- Run unit tests for changed packages

**CI Pipeline:**
```yaml
# .github/workflows/go-ci.yaml
name: Go CI
on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        go-version: ['1.22']
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - uses: cachix/install-nix-action@v24
      - run: go mod download
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go build -v ./cmd/om
      - uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
  
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: golangci/golangci-lint-action@v3
        with:
          version: latest
```

### 4.5 Test Data Management

- Use `testdata/` directories for fixtures
- Embed test files with `//go:embed` directive
- Create minimal Nix flakes for testing
- Use golden files for expected outputs
- Mock GitHub API responses

---

## 5. Risk Assessment and Mitigation

### 5.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| **Performance Regression** | Medium | High | Benchmark continuously, optimize hot paths, use profiling |
| **Async Complexity** | Low | Medium | Go's goroutines are simpler than Rust async; thorough testing |
| **Type Safety Loss** | Medium | Medium | Use strict linting, code review, comprehensive tests |
| **Nix Integration Bugs** | Medium | High | Extensive integration testing, gradual rollout |
| **WASM Support Limited** | High | Low | Hybrid approach: keep Rust WASM or defer GUI |
| **Dependency Issues** | Low | Medium | Careful dependency selection, vendor dependencies |
| **Platform-Specific Bugs** | Medium | Medium | CI testing on all platforms, use build tags |
| **Memory Safety Issues** | Low | High | Go's GC prevents most issues; careful with CGO (avoided) |

### 5.2 Project Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| **Timeline Overrun** | High | Medium | Phased approach allows extending individual phases |
| **Resource Constraints** | Medium | High | Clear documentation enables community contribution |
| **Feature Creep** | Medium | Medium | Strict feature freeze during migration |
| **Incomplete Migration** | Low | High | Well-defined success criteria per phase |
| **User Adoption Issues** | Medium | Medium | Beta testing, migration guide, backward compatibility |
| **Team Expertise Gap** | Medium | Medium | Training, documentation, Go best practices guide |
| **Breaking Changes** | Medium | High | Semantic versioning (v2), deprecation notices |

### 5.3 Security Considerations

**Rust Security Features Lost:**
- Memory safety (Go has GC, different trade-offs)
- Ownership system (must rely on conventions)
- Borrow checker (manual memory management awareness)

**Go Security Measures:**
- Disable CGO for static linking (no C dependencies)
- Use `go mod verify` to check dependencies
- Regular dependency updates via Dependabot
- Security scanning with `govulncheck`
- Input validation and sanitization
- Avoid reflection where possible
- No `unsafe` package usage unless absolutely necessary

**Supply Chain Security:**
- Pin dependency versions in go.mod
- Vendor dependencies for Nix builds
- Verify checksums (vendorHash in Nix)
- Regular security audits
- SBOM generation for releases

### 5.4 Compatibility Risks

**Behavioral Changes:**
- Subtle differences in JSON/YAML parsing
- Error message wording differences
- Output formatting variations
- Timing and ordering in concurrent operations

**Mitigation:**
- Extensive regression testing
- Compare outputs between Rust and Go versions
- Document any intentional changes
- Provide compatibility flags if needed

---

## 6. Documentation Plan

### 6.1 Documentation Updates

**User Documentation:**
- [ ] Update omnix.page website
- [ ] Update README.md
- [ ] Update command documentation (om.md, etc.)
- [ ] Create v1 to v2 migration guide
- [ ] Update FAQ with Go-specific information
- [ ] Update installation instructions

**Developer Documentation:**
- [ ] Create CONTRIBUTING.md for Go
- [ ] Document Go code style and conventions
- [ ] Create Go development setup guide
- [ ] Update build and release process
- [ ] Document testing practices
- [ ] Create API documentation with godoc

**Internal Documentation:**
- [ ] Architecture decision records (ADRs)
- [ ] Migration decision log
- [ ] Package dependency map
- [ ] Performance benchmarking results
- [ ] Security review documentation

### 6.2 Documentation Tools

- `godoc` - API documentation generation
- Markdown for guides and READMEs
- Code comments following Go conventions
- Examples in documentation (`Example*` functions)

### 6.3 Documentation Maintenance

- Keep docs in sync with code
- Review docs in every PR
- Update examples when APIs change
- Maintain changelog (history.md)
- Version documentation with releases

---

## 7. Timeline and Milestones

### 7.1 Detailed Timeline

| Week | Phase | Milestone | Deliverable |
|------|-------|-----------|-------------|
| 1-4 | Foundation | Go setup complete | Working Go build, common pkg migrated |
| 5-8 | Nix Integration | Core Nix layer done | pkg/nix functional with tests |
| 9-11 | Health & Init | User features working | `om health` and `om init` commands |
| 12-15 | CI & Develop | Business logic migrated | `om ci` and `om develop` commands |
| 16-18 | CLI Integration | All commands working | Complete `om` CLI in Go |
| 19-21 | GUI & Testing | Quality assurance | 80%+ test coverage, GUI decision |
| 22-24 | Release | v2.0.0 launched | Production-ready Go version |

### 7.2 Key Milestones

**M1 - Foundation Complete (Week 4)**
- Go module structure established
- CI pipeline for Go in place
- Common package migrated and tested
- Development environment documented

**M2 - Nix Integration (Week 8)**
- All Nix operations working in Go
- Integration tests passing
- Performance benchmarks completed
- No regressions from Rust version

**M3 - Feature Parity 50% (Week 11)**
- Health and Init commands functional
- CLI framework established
- User acceptance testing initiated

**M4 - Feature Parity 100% (Week 18)**
- All commands migrated
- End-to-end tests passing
- Documentation updated
- Beta release candidate ready

**M5 - Production Ready (Week 21)**
- Test coverage goals met
- GUI strategy decided and implemented
- Performance optimizations complete
- Security review complete

**M6 - General Availability (Week 24)**
- v2.0.0 released
- Nix package updated
- Community informed
- Migration guide published

### 7.3 Dependencies and Blockers

**Critical Path:**
1. Foundation → Nix Integration → All other features
2. Cannot release without test coverage goals
3. Beta testing must complete before GA

**External Dependencies:**
- Nix stable releases (compatibility testing)
- Go language updates (currently 1.22+)
- GitHub Actions API stability
- Community feedback on beta

---

## 8. Go-Specific Considerations

### 8.1 Language Feature Mapping

| Rust Feature | Go Equivalent | Notes |
|--------------|---------------|-------|
| `Result<T, E>` | `(T, error)` | Go's idiomatic error handling |
| `Option<T>` | `*T` or sentinel | Pointers can be nil, or use zero values |
| Enums with data | Interfaces + types | Use interface + concrete types |
| Pattern matching | Type switch | `switch v := x.(type)` |
| Traits | Interfaces | Similar concept |
| Generics | Generics (Go 1.18+) | Available and usable |
| Macros | Code generation | Use `go generate` |
| Async/await | Goroutines + channels | Different model but powerful |
| Ownership | Convention | Use pointers carefully, document ownership |

### 8.2 Idiomatic Go Patterns

**Error Handling:**
```go
// Idiomatic Go error handling
func RunNixCommand(cmd string, args ...string) (string, error) {
    output, err := exec.Command(cmd, args...).CombinedOutput()
    if err != nil {
        return "", &NixError{
            Command: cmd,
            ExitCode: err.(*exec.ExitError).ExitCode(),
            Stderr: string(output),
            Err: err,
        }
    }
    return string(output), nil
}

// Usage
output, err := RunNixCommand("nix", "flake", "show")
if err != nil {
    return fmt.Errorf("failed to show flake: %w", err)
}
```

**Interfaces:**
```go
// Define small, focused interfaces
type NixCommand interface {
    Execute(ctx context.Context) (string, error)
}

type FlakeOperation interface {
    Show() (*FlakeMetadata, error)
    Check() error
    Lock() error
}
```

**Struct Embedding:**
```go
// Embed common fields
type BaseCheck struct {
    Name string
    Description string
}

func (b *BaseCheck) GetName() string {
    return b.Name
}

type NixVersionCheck struct {
    BaseCheck
    MinVersion string
}
```

**Table-Driven Tests:**
```go
func TestParseVersion(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Version
        wantErr bool
    }{
        {name: "valid", input: "2.18.0", want: Version{2, 18, 0}, wantErr: false},
        {name: "invalid", input: "invalid", want: Version{}, wantErr: true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseVersion(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 8.3 Performance Optimization

**Memory Management:**
- Use sync.Pool for frequently allocated objects
- Avoid unnecessary string allocations (use []byte where appropriate)
- Use buffer pools for I/O operations
- Profile with pprof to find hotspots

**Concurrency:**
- Use worker pools for bounded parallelism
- Use context for cancellation and timeouts
- Avoid goroutine leaks (ensure they exit)
- Use sync.WaitGroup for coordination

**Build Optimization:**
```bash
# Production build with optimizations
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath ./cmd/om

# Flags explanation:
# -s -w: Strip debug symbols and DWARF info
# -trimpath: Remove absolute file paths from binary
# CGO_ENABLED=0: Static linking, no C dependencies
```

### 8.4 Tooling and Quality

**Development Tools:**
- `gofmt` - Code formatting (run automatically)
- `goimports` - Import management
- `golangci-lint` - Comprehensive linting
- `go vet` - Static analysis
- `staticcheck` - Advanced static analysis
- `gopls` - Language server for IDE support

**Quality Gates:**
- Code coverage: minimum 80%
- Linting: zero warnings
- Code formatting: enforced by CI
- Documentation: godoc for all public APIs
- Benchmarks: for critical paths

---

## 9. Community and Ecosystem

### 9.1 Community Impact

**Benefits to Community:**
- **Easier Contributions**: Go's simplicity lowers barrier to entry
- **Faster Builds**: Go compiles faster than Rust
- **Broader Ecosystem**: More Go developers than Rust developers
- **Better IDE Support**: Excellent Go tooling across editors

**Migration Support:**
- Community beta testing program
- Migration office hours (Discord/GitHub Discussions)
- Detailed migration guide
- FAQ for common issues
- Video walkthroughs

### 9.2 Ecosystem Integration

**Nix Ecosystem:**
- Maintain compatibility with nixpkgs
- Support for Nix flakes
- Integration with nix-darwin, home-manager
- Work with Determinate Systems installer

**Go Ecosystem:**
- Publish to pkg.go.dev
- Follow Go module best practices
- Support go install
- Potential for Go-based Nix tools collaboration

### 9.3 Breaking Changes Communication

**Semantic Versioning:**
- v1.x.y - Rust version (maintenance mode)
- v2.0.0 - Go rewrite (new major version)
- Clear versioning communicates breaking changes

**Deprecation Notice:**
- Add deprecation notice to v1 releases
- Point to migration guide
- Announce on all channels:
  - GitHub Discussions
  - Project website
  - Release notes
  - Social media

---

## 10. Success Criteria

### 10.1 Technical Success Criteria

- [ ] All commands migrated with feature parity
- [ ] Test coverage ≥ 80%
- [ ] Performance within 10% of Rust version
- [ ] Binary size within 20% of Rust version
- [ ] Zero critical bugs in first month post-release
- [ ] CI/CD pipeline fully functional
- [ ] Cross-platform support maintained
- [ ] Security audit passed

### 10.2 User Success Criteria

- [ ] Migration guide published and clear
- [ ] Beta testing feedback incorporated
- [ ] No major user-reported regressions
- [ ] Documentation comprehensive and accurate
- [ ] Community adoption > 80% within 3 months
- [ ] Positive community feedback
- [ ] Support issues resolved promptly

### 10.3 Project Success Criteria

- [ ] Timeline met (or acceptable delay)
- [ ] Budget maintained (primarily time-based)
- [ ] Team satisfaction with Go codebase
- [ ] Easier onboarding for new contributors
- [ ] Reduced build times for developers
- [ ] Maintainable and well-structured code
- [ ] Clear path for future development

---

## 11. Future Considerations

### 11.1 Post-Migration Enhancements

Once Go migration is complete, consider:

**Performance Improvements:**
- Profile and optimize hot paths
- Consider parallel execution where safe
- Optimize memory allocations
- Cache frequently accessed data

**New Features:**
- Features deferred during migration
- Community-requested enhancements
- Go-specific optimizations
- Better error messages and diagnostics

**Tooling:**
- VSCode extension for omnix
- Language server for om.yaml
- Nix flake template for Go projects
- Integration with other Nix tools

### 11.2 Long-Term Maintenance

**Regular Updates:**
- Keep dependencies up to date
- Follow Go security advisories
- Update for new Nix versions
- Maintain compatibility with nixpkgs

**Community Building:**
- Encourage contributions
- Mentor new contributors
- Recognize contributors
- Build a maintainer team

---

## 12. Conclusion

The migration from Rust to Go represents a significant but achievable undertaking. The phased approach ensures that:

1. **Quality is maintained** through extensive testing at each phase
2. **Users experience minimal disruption** with parallel Rust maintenance
3. **The team can adapt** with flexibility to extend phases if needed
4. **Success is measurable** with clear criteria at each milestone

The Go implementation will provide:
- **Simpler codebase** that's easier for contributors
- **Faster builds** during development
- **Strong Nix integration** maintained through careful design
- **Production-ready quality** through comprehensive testing

By following this design document, the omnix project will successfully transition to Go while maintaining its core mission: being a developer-friendly companion for Nix.

---

## Appendix A: Rust to Go Code Examples

### Example 1: Error Handling

**Rust:**
```rust
use anyhow::{Context, Result};

fn parse_config(path: &Path) -> Result<Config> {
    let content = fs::read_to_string(path)
        .context("Failed to read config file")?;
    let config: Config = serde_yaml::from_str(&content)
        .context("Failed to parse config YAML")?;
    Ok(config)
}
```

**Go:**
```go
import (
    "fmt"
    "os"
    "gopkg.in/yaml.v3"
)

func ParseConfig(path string) (*Config, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config Config
    if err := yaml.Unmarshal(content, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config YAML: %w", err)
    }
    
    return &config, nil
}
```

### Example 2: Async Operations

**Rust:**
```rust
use tokio::process::Command;

async fn run_nix_command(args: &[&str]) -> Result<String> {
    let output = Command::new("nix")
        .args(args)
        .output()
        .await?;
    
    if !output.status.success() {
        anyhow::bail!("nix command failed");
    }
    
    Ok(String::from_utf8(output.stdout)?)
}
```

**Go:**
```go
import (
    "context"
    "fmt"
    "os/exec"
)

func RunNixCommand(ctx context.Context, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, "nix", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("nix command failed: %w", err)
    }
    
    return string(output), nil
}

// Usage with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := RunNixCommand(ctx, "flake", "show")
```

### Example 3: Struct with Methods

**Rust:**
```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FlakeUrl {
    pub url: String,
}

impl FlakeUrl {
    pub fn parse(s: &str) -> Result<Self> {
        Ok(Self { url: s.to_string() })
    }
    
    pub async fn show(&self) -> Result<FlakeMetadata> {
        let output = run_nix_command(&["flake", "show", &self.url, "--json"]).await?;
        let metadata = serde_json::from_str(&output)?;
        Ok(metadata)
    }
}
```

**Go:**
```go
type FlakeUrl struct {
    URL string `json:"url" yaml:"url"`
}

func ParseFlakeUrl(s string) (*FlakeUrl, error) {
    return &FlakeUrl{URL: s}, nil
}

func (f *FlakeUrl) Show(ctx context.Context) (*FlakeMetadata, error) {
    output, err := RunNixCommand(ctx, "flake", "show", f.URL, "--json")
    if err != nil {
        return nil, err
    }
    
    var metadata FlakeMetadata
    if err := json.Unmarshal([]byte(output), &metadata); err != nil {
        return nil, fmt.Errorf("failed to parse metadata: %w", err)
    }
    
    return &metadata, nil
}
```

---

## Appendix B: Testing Examples

### Unit Test Example

**Go:**
```go
package nix

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParseFlakeUrl(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "github URL",
            input:   "github:juspay/omnix",
            want:    "github:juspay/omnix",
            wantErr: false,
        },
        {
            name:    "local path",
            input:   ".",
            want:    ".",
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseFlakeUrl(tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want, got.URL)
        })
    }
}
```

### Integration Test Example

**Go:**
```go
// +build integration

package nix_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/juspay/omnix/pkg/nix"
)

func TestFlakeShow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    ctx := context.Background()
    flake, err := nix.ParseFlakeUrl("github:juspay/omnix")
    require.NoError(t, err)
    
    metadata, err := flake.Show(ctx)
    require.NoError(t, err)
    require.NotNil(t, metadata)
    
    // Validate metadata structure
    require.NotEmpty(t, metadata.Description)
}
```

---

## Appendix C: Build Configuration

### justfile for Go

```makefile
# justfile for Go development

# Run tests
test:
    go test -v -race ./...

# Run tests with coverage
test-coverage:
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run linting
lint:
    golangci-lint run

# Format code
fmt:
    go fmt ./...
    goimports -w .

# Build binary
build:
    CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/om ./cmd/om

# Run locally
run *ARGS:
    go run ./cmd/om {{ARGS}}

# Watch and rebuild on changes
watch *ARGS:
    watchexec -e go -r -- go run ./cmd/om {{ARGS}}

# Build for all platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -o bin/om-linux-amd64 ./cmd/om
    GOOS=linux GOARCH=arm64 go build -o bin/om-linux-arm64 ./cmd/om
    GOOS=darwin GOARCH=amd64 go build -o bin/om-darwin-amd64 ./cmd/om
    GOOS=darwin GOARCH=arm64 go build -o bin/om-darwin-arm64 ./cmd/om

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

# Run CI locally
ci: fmt lint test-coverage build

# Update dependencies
update-deps:
    go get -u ./...
    go mod tidy
```

---

## Appendix D: References and Resources

### Go Learning Resources
- Official Go Documentation: https://go.dev/doc/
- Effective Go: https://go.dev/doc/effective_go
- Go by Example: https://gobyexample.com/
- Go Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments

### Go Testing
- Go Testing Best Practices: https://go.dev/doc/tutorial/add-a-test
- Testify Documentation: https://github.com/stretchr/testify
- Table-Driven Tests: https://github.com/golang/go/wiki/TableDrivenTests

### Nix with Go
- buildGoModule Documentation: https://nixos.org/manual/nixpkgs/stable/#sec-language-go
- Nix Flakes: https://nixos.wiki/wiki/Flakes

### Tools
- golangci-lint: https://golangci-lint.run/
- Cobra CLI: https://cobra.dev/
- Zap Logger: https://github.com/uber-go/zap

### Migration Examples
- kubernetes/kubernetes (C++ to Go historical reference)
- Various open-source projects documented migrations

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-17  
**Authors:** Omnix Development Team  
**Status:** Draft for Review
