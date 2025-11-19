# Future Work - Deferred Items

This document tracks future work items that were identified but deferred during the implementation of issue #19. These items require significant additional infrastructure or deeper integration with external systems.

## Registry Support

**Status**: Deferred  
**Complexity**: High  
**Dependencies**: External registry service infrastructure

### Description
Add support for discovering and using templates from a centralized registry.

### Affected Packages
- `pkg/init` - Template discovery and loading
- `pkg/cli` - CLI integration for registry commands

### Requirements
- Design and implement registry service backend
- Define registry API and data formats
- Add authentication and authorization
- Implement template versioning
- Add caching mechanism for registry lookups

### Related Future Work
- Template test execution (see below)

---

## Template Test Execution

**Status**: Deferred  
**Complexity**: High  
**Dependencies**: Testing infrastructure, potentially CI integration

### Description
Automatically execute tests defined within templates during scaffolding or as a validation step.

### Affected Packages
- `pkg/init` - Template processing and execution

### Requirements
- Define template test specification format
- Implement test runner within template context
- Add test result reporting
- Handle test failures gracefully
- Support multiple test frameworks (shell, Nix, etc.)

### Considerations
- Security implications of running arbitrary code from templates
- Sandboxing requirements
- Resource limits and timeouts

---

## Loading Templates from Flakes

**Status**: Deferred  
**Complexity**: High  
**Dependencies**: Deep Nix integration, flake schema understanding

### Description
Support using Nix flakes directly as template sources, allowing templates to be defined and distributed via flakes.

### Affected Packages
- `pkg/init` - Template loading mechanism
- `pkg/nix` - Flake parsing and interaction

### Requirements
- Parse flake template schemas
- Handle flake inputs and dependencies
- Support flake-based template parameters
- Implement proper flake evaluation
- Handle template updates via flake lock

### Technical Challenges
- Nix evaluation complexity
- Flake schema variability
- Performance of Nix evaluation
- Error handling for malformed flakes

---

## Integration Tests with Real Nix

**Status**: Deferred  
**Complexity**: Medium-High  
**Dependencies**: CI environment with Nix installed, test infrastructure

### Description
Add comprehensive integration tests that interact with actual Nix installations rather than mocked data.

### Affected Packages
- All packages, particularly:
  - `pkg/health` - Health check validation
  - `pkg/nix` - Nix command execution
  - `pkg/ci` - CI workflow validation
  - `pkg/develop` - Development environment setup

### Requirements
- Set up CI runners with Nix installed
- Create test fixtures (flakes, configurations)
- Implement cleanup mechanisms
- Add test parallelization with proper isolation
- Handle different Nix versions
- Test on multiple platforms (Linux, macOS)

### Implementation Notes
- Tests should be marked with build tags (e.g., `//go:build integration`)
- Create separate test workflow in GitHub Actions
- Document required Nix configuration for tests
- Consider using containers for isolation

---

## Shell Hook Customization

**Status**: Deferred  
**Complexity**: Medium  
**Dependencies**: Deep Nix devShell integration

### Description
Allow users to customize shell hooks that run when entering a development environment.

### Affected Packages
- `pkg/develop` - Development environment configuration

### Requirements
- Define hook configuration schema in `om.yaml`
- Support multiple hook types (pre-shell, post-shell, on-exit)
- Integrate with Nix devShell hook mechanism
- Provide templating for hook scripts
- Handle hook failures gracefully

### Configuration Example
```yaml
develop:
  hooks:
    pre-shell:
      - echo "Setting up environment..."
      - direnv allow
    post-shell:
      - cat README.md
    on-exit:
      - echo "Cleanup complete"
```

---

## Actually Invoke Nix devShell

**Status**: Deferred  
**Complexity**: High  
**Dependencies**: Process management, Nix integration, shell compatibility

### Description
Implement actual Nix devShell invocation instead of just showing a warning message to use direnv.

### Affected Packages
- `pkg/develop` - Shell invocation logic

### Current Behavior
```go
logger.Warn("🚧 !!!!")
logger.Warn("🚧 Not invoking Nix devShell (not supported yet). Please use `direnv`!")
logger.Warn("🚧 !!!!")
```

### Requirements
- Detect user's current shell (bash, zsh, fish, etc.)
- Execute `nix develop` with appropriate shell
- Preserve environment variables
- Handle shell-specific configuration
- Support shell customization via configuration
- Properly exit and return to original shell
- Handle signals (Ctrl+C, etc.)

### Technical Challenges
- Cross-shell compatibility
- Process lifecycle management
- Environment variable isolation
- Signal handling
- Terminal control transfer

---

## Multiple Development Shells Support

**Status**: Deferred  
**Complexity**: High  
**Dependencies**: Nix multi-shell architecture, configuration design

### Description
Support projects with multiple development shells (e.g., separate shells for frontend, backend, documentation).

### Affected Packages
- `pkg/develop` - Multi-shell configuration and selection
- `pkg/cli/cmd` - CLI for shell selection

### Requirements
- Extend configuration to support multiple shell definitions
- Add shell selection mechanism (`om develop --shell backend`)
- Support default shell configuration
- List available shells (`om develop --list`)
- Document shell purposes and dependencies

### Configuration Example
```yaml
develop:
  shells:
    default:
      health-checks:
        nix-version: true
      readme:
        file: README.md
    frontend:
      health-checks:
        nix-version: true
        node: true
      readme:
        file: frontend/README.md
    backend:
      health-checks:
        nix-version: true
        rust: true
      readme:
        file: backend/README.md
```

---

## Development Environment Templates

**Status**: Deferred  
**Complexity**: Medium  
**Dependencies**: Template infrastructure, Nix devShell configuration

### Description
Provide pre-configured development environment templates for common technology stacks.

### Affected Packages
- `pkg/init` - Template integration
- `pkg/develop` - Environment validation

### Examples
- `rust-workspace` - Rust development with cargo, clippy, rustfmt
- `node-typescript` - Node.js with TypeScript, ESLint, Prettier
- `python-poetry` - Python with Poetry, pytest, mypy
- `go-modules` - Go with modules, golangci-lint, delve

### Requirements
- Define template format for dev environments
- Include health checks specific to each stack
- Provide sensible defaults
- Allow customization
- Document each template

---

## Summary

### Deferred Items by Complexity

**High Complexity** (5 items):
1. Registry Support
2. Template Test Execution
3. Loading Templates from Flakes
4. Actually Invoke Nix devShell
5. Multiple Development Shells Support

**Medium-High Complexity** (1 item):
1. Integration Tests with Real Nix

**Medium Complexity** (2 items):
1. Shell Hook Customization
2. Development Environment Templates

### Recommended Implementation Order

For future PRs, consider implementing in this order:

1. **Integration Tests with Real Nix** - Provides foundation for testing other features
2. **Shell Hook Customization** - Relatively self-contained, high value
3. **Development Environment Templates** - Builds on existing template infrastructure
4. **Multiple Development Shells Support** - Prerequisite for some other features
5. **Actually Invoke Nix devShell** - Requires multi-shell support
6. **Loading Templates from Flakes** - Deep Nix integration
7. **Registry Support** - Requires external infrastructure
8. **Template Test Execution** - Depends on registry and template loading

### Infrastructure Requirements

Before implementing these features, consider:

- **CI/CD**: Nix-enabled runners, increased build times
- **Security**: Template execution sandboxing, code review process
- **Documentation**: User guides, API documentation
- **Testing**: Comprehensive test suites, multiple platform testing
- **Maintenance**: Long-term support commitment, deprecation strategy

---

## Contributing

If you're interested in implementing any of these features:

1. Open an issue to discuss the approach
2. Reference this document in the issue
3. Break down the work into smaller, manageable PRs
4. Ensure proper testing and documentation
5. Consider backward compatibility

## Related Issues

- #19 - Implement future work (original issue for completed items)

---

*Last updated: 2025-11-19*
*Based on analysis during PR for issue #19*
