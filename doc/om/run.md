---
order: 6
---

# `om run`

Run tasks from the `om/` directory with a simplified configuration format.

`om run` is similar to `om ci run`, but designed for easier task execution with sensible defaults:

- Reads configuration from `om/default.yaml` by default
- Passing a parameter runs that specific task: `om run <name>` loads `om/<name>.yaml`
- Disables `lockfile`, `build`, and `flakeCheck` steps by default for faster execution
- Uses simplified YAML structure without nested `ci` configuration

## Usage

```sh
# Run the default task (om/default.yaml)
om run

# Run a specific task (om/update.yaml)
om run update
```

## Configuration Format

The simplified `om/` configuration format is much cleaner than the full `om.yaml` CI format.

### Example: `om/default.yaml`

```yaml
dir: .
steps:
  activate-configuration:
    type: app
    name: activate
```

This is equivalent to the following in `om.yaml` for `om ci run`:

```yaml
ci:
  default:
    .:
      dir: .
      steps:
        lockfile:
          enable: false
        build:
          enable: false
        flakeCheck:
          enable: false
        custom:
          activate-configuration:
            type: app
            name: activate
```

### Configuration Fields

- **`dir`**: Directory where the flake lives (default: `.`)
- **`steps`**: Custom steps to run (same format as `om ci` custom steps)
  - `type`: Either `app` (run a flake app) or `devshell` (run in dev shell)
  - For apps: `name` and optionally `args`
  - For devshells: `command` (array of command and arguments)

Note: Steps execute in the order they appear in the YAML file.

### Example: DevShell Task

```yaml
dir: .
steps:
  test:
    type: devshell
    command:
      - cargo
      - test
  format:
    type: devshell
    command:
      - cargo
      - fmt
```

### Example: Multiple Apps

```yaml
dir: .
steps:
  build-docs:
    type: app
    name: build-docs
  deploy:
    type: app
    name: deploy
    args:
      - --production
```

## Differences from `om ci run`

| Feature | `om run` | `om ci run` |
|---------|----------|-------------|
| Config location | `om/*.yaml` | `om.yaml` |
| Default config | `om/default.yaml` | `om.yaml#ci.default` |
| Config structure | Simplified (direct steps) | Nested under `ci` key |
| Lockfile check | Disabled by default | Enabled by default |
| Build step | Disabled by default | Enabled by default |
| Use case | Quick tasks, scripts | Full CI pipeline |

## Common Use Cases

### Development Tasks

Create `om/dev.yaml`:

```yaml
dir: .
steps:
  watch:
    type: devshell
    command:
      - just
      - watch
  test:
    type: devshell
    command:
      - just
      - test
```

Run with: `om run dev`

### Deployment Tasks

Create `om/deploy.yaml`:

```yaml
dir: .
steps:
  deploy-staging:
    type: app
    name: deploy
    args:
      - --environment
      - staging
  deploy-production:
    type: app
    name: deploy
    args:
      - --environment
      - production
```

Run with: `om run deploy`

### Update Tasks

Create `om/update.yaml`:

```yaml
dir: .
steps:
  update-flake:
    type: devshell
    command:
      - nix
      - flake
      - update
  update-deps:
    type: devshell
    command:
      - cargo
      - update
```

Run with: `om run update`

## Error Handling

`om run` validates configuration and provides clear error messages:

### Type Validation

All arguments and commands must be strings in YAML:

```yaml
# ❌ WRONG - numeric values not allowed
steps:
  bad-step:
    type: devshell
    command:
      - echo
      - 123  # Error: expected string, got int

# ✅ CORRECT - all values are strings
steps:
  good-step:
    type: devshell
    command:
      - echo
      - "123"
```

### Step Order

Steps execute in the order they appear in the YAML file. This is guaranteed and deterministic:

```yaml
steps:
  first:   # Runs first
    type: devshell
    command: [echo, "1"]
  second:  # Runs second
    type: devshell
    command: [echo, "2"]
  third:   # Runs third
    type: devshell
    command: [echo, "3"]
```

### Working Directory

The `dir` field sets the working directory for all steps:

```yaml
dir: ./my-project
steps:
  build:
    type: devshell
    command:
      - cargo
      - build
```

If `dir` is not specified, it defaults to `.` (current directory).

## See Also

- [[ci]] - Full CI pipeline configuration
- [[show]] - Inspect flake outputs
- [[develop]] - Enter development environment
