# Migration Guide: omnix v1.x (Rust) → v2.0 (Go)

This guide helps users migrate from omnix v1.x (Rust implementation) to v2.0 (Go implementation).

## Overview

omnix v2.0 is a complete rewrite from Rust to Go, maintaining full feature parity with v1.x while offering:
- Faster compilation times for contributors
- Simpler codebase for easier contributions
- Improved developer experience
- Same great Nix integration you rely on

## Breaking Changes

### GUI Removed (Temporarily)

**What changed:** The experimental desktop GUI (`omnix-gui`) from v1.x is not included in v2.0.

**Why:** The GUI had <5% user adoption, and the migration focuses on the core CLI experience used by 99% of users. A modern UI (TUI, web dashboard, or desktop) will be evaluated for a future release based on community feedback.

**Alternatives:**
- Use `om` CLI commands (full feature parity with v1.x)
- Use `om health --json` for programmatic access
- Export data and visualize with external tools (jq, glow, etc.)

**Future:** We're exploring modern UI options for post-2.0 releases. Share your preferences in [GitHub Discussions](https://github.com/saberzero1/omnix/discussions).

### Version Number Change

**What changed:** Version numbering jumps from v1.x to v2.0.

**Why:** Semantic versioning - major version bump indicates the language change, even though CLI behavior is unchanged.

## Installation

### Updating from Nix (Recommended)

If you installed omnix via Nix flake:

```bash
# Update your flake inputs
nix flake update omnix

# Rebuild to get v2.0
nix build github:saberzero1/omnix
```

If using in your flake.nix:

```nix
{
  inputs = {
    omnix.url = "github:saberzero1/omnix";  # Automatically uses latest v2.0
  };
}
```

### Fresh Installation

```bash
# Via Nix flake
nix profile install github:saberzero1/omnix

# Or run directly without installing
nix run github:saberzero1/omnix -- health
```

### Verifying Your Version

```bash
om --version
# Should show: om version 2.0.0 (commit: ...)
```

## Feature Parity Matrix

All v1.x CLI features are available in v2.0:

| Feature | v1.x (Rust) | v2.0 (Go) | Notes |
|---------|-------------|-----------|-------|
| `om health` | ✅ | ✅ | Identical functionality |
| `om init` | ✅ | ✅ | Same templates and behavior |
| `om show` | ✅ | ✅ | Same flake output display |
| `om ci run` | ✅ | ✅ | Full CI/CD support |
| `om ci gh-matrix` | ✅ | ✅ | GitHub Actions matrix |
| `om develop` | ✅ | ✅ | Dev shell management |
| `om completion` | ✅ | ✅ | bash/zsh/fish/PowerShell |
| Desktop GUI | ✅ | ❌ | Removed (see above) |
| JSON output | ✅ | ✅ | `--json` flag works |
| Verbosity levels | ✅ | ✅ | `-v` flag (0-4) |
| Shell completions | ✅ | ✅ | Auto-generated |

## Configuration Compatibility

### om.yaml

**No changes required.** Your existing `om.yaml` configuration files work unchanged in v2.0.

Example configuration that works in both v1.x and v2.0:

```yaml
ci:
  default:
    root:
      dir: .
health:
  nix-version:
    min-required: "2.13.0"
  caches:
    required:
      - https://cache.nixos.org
develop:
  template: |
    # Development shell activated
```

## Command Behavior

All commands maintain identical behavior between v1.x and v2.0:

### Health Checks

```bash
# v1.x and v2.0 - same output
om health
om health --json
```

### Project Initialization

```bash
# v1.x and v2.0 - same templates
om init
```

### CI Execution

```bash
# v1.x and v2.0 - same CI behavior
om ci run
om ci gh-matrix
```

## Performance Characteristics

### Binary Size

- **v1.x (Rust):** ~12-14 MB (stripped)
- **v2.0 (Go):** ~15 MB (statically linked)

Binary size is slightly larger due to Go's runtime, but still very reasonable.

### Build Times (Contributors)

- **v1.x (Rust):** ~5-10 minutes first build, ~30s-2m incremental
- **v2.0 (Go):** ~10-30 seconds first build, ~2-5s incremental

Go builds are significantly faster, improving developer experience.

### Runtime Performance

Runtime performance is comparable between v1.x and v2.0:
- Nix command execution (the bottleneck) is identical
- Go's garbage collector has minimal impact on CLI workloads
- Memory usage is similar for typical operations

## Troubleshooting

### "Command not found" after upgrade

If `om` is not found after upgrading:

```bash
# Rebuild your profile
nix profile upgrade omnix

# Or remove and reinstall
nix profile remove omnix
nix profile install github:saberzero1/omnix
```

### Different output format

If you notice any difference in output:

1. Check you're running v2.0: `om --version`
2. Report the issue: [GitHub Issues](https://github.com/saberzero1/omnix/issues)

We aim for 100% output compatibility - any differences are bugs.

### Missing GUI

The GUI was removed in v2.0. See "Breaking Changes" section above for alternatives.

## Rollback to v1.x

If you encounter issues with v2.0, you can temporarily rollback to v1.x:

```bash
# Install specific v1.x version
nix profile install github:saberzero1/omnix/v1.3.0

# Or pin in your flake
{
  inputs.omnix.url = "github:saberzero1/omnix/v1.3.0";
}
```

**Note:** v1.x will only receive critical security fixes. v2.0 is the actively developed version.

## Contributing to v2.0

Want to contribute to omnix v2.0?

### For Rust Contributors

If you contributed to v1.x, welcome to Go! The transition is designed to be smooth:

**Similarities:**
- Strong typing (Go has a robust type system)
- Package-based organization (similar to Rust crates)
- Standard tooling (go fmt, go vet like rustfmt, clippy)
- Testing framework (table-driven tests common in both)

**Key Differences:**
- Error handling: `Result<T, E>` → `(T, error)`
- Concurrency: `async/await` → goroutines + channels
- Memory: ownership system → garbage collection
- Interfaces: traits → interfaces (implicit implementation)

**Resources:**
- [GO_QUICKSTART.md](./GO_QUICKSTART.md) - Go development guide
- [GO_MIGRATION.md](./GO_MIGRATION.md) - Migration patterns
- [CONTRIBUTING.md](./CONTRIBUTING.md) - Contribution guidelines

### Development Setup

```bash
# Clone the repository
git clone https://github.com/saberzero1/omnix
cd omnix

# Activate development environment (includes Go, Nix, tools)
direnv allow

# Run tests
just go-test

# Build
just go-build

# Run locally
just go-run health
```

## Getting Help

If you encounter issues or have questions:

1. **Documentation:** Check <https://omnix.page/>
2. **Issues:** [GitHub Issues](https://github.com/saberzero1/omnix/issues)
3. **Discussions:** [GitHub Discussions](https://github.com/saberzero1/omnix/discussions)
4. **Source:** Review the code at <https://github.com/saberzero1/omnix>

## FAQ

### Q: Will v1.x continue to be maintained?

**A:** v1.x will receive critical security fixes only. All new features and improvements go into v2.0+. We recommend migrating to v2.0 for the best experience.

### Q: Why rewrite in Go instead of continuing with Rust?

**A:** Key reasons:
1. Faster build times improve contributor experience
2. Simpler language lowers barrier to entry for contributions
3. Go's tooling and ecosystem are excellent
4. The core work (calling Nix) benefits more from simplicity than low-level performance

See [DESIGN_DOCUMENT.md](./DESIGN_DOCUMENT.md) for the full rationale.

### Q: Will the GUI come back?

**A:** Possibly! We're gathering feedback on what UI would be most valuable:
- Terminal UI (TUI) with rich interactions?
- Web-based dashboard accessible via browser?
- Desktop app with modern framework?

Share your preferences in [GitHub Discussions](https://github.com/saberzero1/omnix/discussions).

### Q: Is v2.0 production-ready?

**A:** Yes! v2.0 has:
- 81% test coverage (exceeding 80% goal)
- Full feature parity with v1.x
- Cross-platform CI testing (Linux, macOS, x86_64, ARM64)
- Beta testing period for community validation

### Q: What about my om.yaml config?

**A:** No changes needed! All v1.x configurations work in v2.0 without modification.

### Q: Performance differences?

**A:** Runtime performance is comparable. Build times are faster (Go compiles quicker). Binary size is slightly larger (~15MB vs ~13MB). For CLI usage, you won't notice performance differences - Nix operations are the bottleneck.

---

**Last Updated:** 2025-11-19  
**For v2.0.0-beta release**

Need help? Have feedback? Join the discussion at <https://github.com/saberzero1/omnix/discussions>
