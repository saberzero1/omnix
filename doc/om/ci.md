# `om ci`

`om ci` runs continuous integration (CI)-friendly builds for your project. It builds all outputs in the flake, or optionally its [sub-flakes](https://github.com/hercules-ci/flake-parts/issues/119). You can run `om ci` locally or in an actual CI envirnoment, like GitHub Actions. Using [devour-flake] it will automatically build the following outputs:

| Type                   | Output Key                                      |
| ---------------------- | ----------------------------------------------- |
| Standard flake outputs | `packages`, `apps`, `checks`, `devShells`       |
| NixOS                  | `nixosConfigurations.*`                         |
| nix-darwin             | `darwinConfigurations.*`                        |
| home-manager           | `legacyPackages.${system}.homeConfigurations.*` |

A `result` symlink is also produced, containing a JSON of all built paths. See [here](#out-link).

## Basic Usage {#usage}

`om ci run` accepts any valid [flake URL](https://nixos.asia/en/flake-url) or a Github PR URL.

By default, `om ci run` displays an interactive terminal UI with real-time progress tracking, spinner animations, and color-coded status indicators. The UI is automatically disabled for non-interactive environments (pipes, CI) or can be manually disabled with `--no-ui`.

```sh
# Run CI on current directory flake (with interactive UI)
$ om ci # Or `om ci run` or `om ci run .`

# Disable the interactive UI (useful for scripts or CI)
$ om ci run --no-ui

# Run CI on a local flake (default is $PWD)
$ om ci run ~/code/myproject

# Pass custom arguments to `nix` after '--'
$ om ci run ~/code/myproject -- --accept-flake-config

# Run CI on a github repo
$ om ci run github:hercules-ci/hercules-ci-agent

# Run CI on a github PR
$ om ci run https://github.com/srid/emanote/pull/451

# Run CI only the selected sub-flake
$ git clone https://github.com/srid/haskell-flake && cd haskell-flake
$ om ci run .#default.dev

# Run CI remotely over SSH
$ om ci run --on ssh://myname@myserver ~/code/myproject
```

### Interactive UI Features {#ui}

When running in an interactive terminal, `om ci run` displays a live progress view with:

- **Real-time progress tracking**: See each step as it runs with spinner animations
- **Status indicators**: Color-coded symbols (○ pending, ◐ running, ✓ success, ✗ failed, ⊘ skipped)
- **Duration tracking**: See how long each step and subflake takes
- **Interactive controls**: Press 'o' to toggle output details, 'q' to quit
- **Clean formatting**: No JSON strings, proper newlines and structured layout

The UI automatically detects terminal support and falls back to JSON logging when:
- Output is piped to another command
- Running in CI environments (with `--github-output`)
- Manually disabled with `--no-ui`


## Results JSON and closure {#out-link}

Just like `nix build`, `om ci` will produce a `result` symlink that contains a JSON of all store paths built. Use options `--out-link <PATH>` and `--no-link` to control this behaviour.

As long as this symlink exists, your built paths will survive garbage collection, because the closure of this symlink contains the entire build closure.

Note that in order to include all build dependencies, you should pass `--include-all-dependencies`, viz.:

```
om ci run --include-all-dependencies | xargs cachix push mycache
```

The above command will push the *entire* build closure (runtime and build dependencies) to the given cache.

## Using in Github Actions {#gh}

In addition to serving the purpose of being a "local CI", `om ci` can be used in Github Actions to enable CI for your GitHub repositories.

### Standard Runners {#gh-simple}

Add this to your workflow file (`.github/workflows/ci.yml`) to build all flake outputs using GitHub provided runners:

```yaml
      - uses: actions/checkout@v4
      - uses: DeterminateSystems/nix-installer-action@main
      - name: Install omnix
        run: nix profile install nixpkgs#omnix
      - run: om ci
```

### Self-hosted Runners with Job Matrix {#gh-matrix}

Here's a more advanced example that configures a job matrix. This is useful when you want to run the CI on multiple systems (e.g. `aarch64-linux`, `aarch64-darwin`), each captured as a separate job by GitHub, as shown in the screenshot below. It also, incidentally, demonstrates how to use self-hosted runners.

![](../ci-github-matrix.png)

The `om ci gh-matrix` command outputs the matrix JSON for creating [a matrix of job variations](https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/running-variations-of-jobs-in-a-workflow). An example configuration, using self-hosted runners, is shown below.

> [!NOTE]
> This currently requires an explicit [CI configuration](../config.md) in your flake, setting `om.ci.default.root.dir` to `.`.

```yaml
# Run on aarch64-linux and aarch64-darwin
jobs:
  configure:
    runs-on: x86_64-linux
    outputs:
      matrix: ${{ steps.set-matrix.outputs.matrix }}
    steps:
     - uses: actions/checkout@v4
     - id: set-matrix
       run: |
         set -euxo pipefail
         MATRIX="$(om ci gh-matrix --systems=x86_64-linux,aarch64-darwin | jq -c .)"
         echo "matrix=$MATRIX" >> $GITHUB_OUTPUT
  nix:
    runs-on: ${{ matrix.system }}
    needs: configure
    strategy:
      matrix: ${{ fromJson(needs.configure.outputs.matrix) }}
      fail-fast: false
    steps:
      - uses: actions/checkout@v4
      - run: om ci run --systems "${{ matrix.system }}" ".#default.${{ matrix.subflake }}"
```

> [!TIP]
> If your builds fail due to GitHub's rate limiting, consider passing `--extra-access-tokens` (see [an example PR](https://github.com/srid/nixos-flake/pull/55)).

## Configuring {#config}

By default, `om ci` will build the top-level flake, but you can tell it to build sub-flakes (here, `./dir1` and `./dir2`) by adding the following to your [Om configuration](../config.md):

```nix
# myproject/flake.nix
{
  om.ci.default = {
    dir1 = {
      dir = "dir1";
    };
    dir2 = {
      dir = "dir2";
      overrideInputs.myproject = ./.;
    };
  }
}
```

You can have more than one CI configuration. For eg., `om ci run .#foo` will run the configuration from `om.ci.foo` flake output.

### Custom CI actions {#custom}

You can define custom CI actions in your flake, which will be run as part of `om ci run`. For example, to run tests in the nix develop shell:

```nix
{
  om.ci.default = {
    root = {
      dir = ".";
      steps = {
        # The build step is enabled by default. It builds all flake outputs.
        build.enable = true;
        # Other steps include: lockfile & flake-check

        # Users can define custom steps to run any arbitrary flake app or devShell command.
        custom = {
          # Here, we run cargo tests in the nix shell
          # This equivalent to `nix develop .#default -c cargo test`
          cargo-test = {
            type = "devshell";
            # name = "default"
            command = [ "cargo" "test" ];
          };

          # We can also flake apps
          # This is equivalent to `nix run .#check-closure-size`
          closure-size = {
            type = "app";
            name = "check-closure-size";
          };
        };
      };
    };
  };
}
```

For a real-world example of custom steps, checkout [Omnix's configuration](https://github.com/saberzero1/omnix/blob/5322235ce4069e72fd5eb477353ee5d1f5100243/nix/modules/om.nix#L16-L33).

## Remote CI {#remote}

Omnix provides two modes for running CI remotely over SSH:

### Metadata-based Remote CI (Recommended) {#remote-metadata}

The `--on` flag enables efficient remote CI with flake caching:

```sh
om ci run --on ssh://myname@myserver ~/code/myproject
```

What this does:

1. **Cache flake locally**: Uses flake metadata to cache the flake (and optionally its inputs) in your local Nix store
2. **Copy once**: Copies the cached flake closure and omnix source to the remote server (one-time initial operation)
3. **Build remotely**: Runs `om ci` on the remote server using the cached flake
4. **Copy results back**: If `--out-link` is specified, copies build results back to local store and creates a GC root

**Key Benefits:**
- Significantly reduces network traffic by caching flakes before copying
- Only copies each flake/input once, even for repeated builds
- Leverages Nix's content-addressed storage for deduplication
- Matches the Rust implementation's efficiency

**Options:**
- `--copy-inputs`: Transitively copy all flake inputs to the remote store. Useful for private Git inputs that the remote can't access
- `--out-link <PATH>`: Creates a local symlink to the results JSON and copies all build outputs back to local store
- `--no-link`: Don't copy results back (faster for CI scenarios where you only need remote builds)

**Examples:**

```sh
# Basic remote CI with caching
om ci run --on ssh://builder@myserver ~/code/myproject

# Copy all flake inputs (useful for private inputs)
om ci run --on "ssh://builder@myserver?copy-inputs=true" ~/code/myproject
# Or use the flag:
om ci run --on ssh://builder@myserver --copy-inputs ~/code/myproject

# Don't copy results back
om ci run --on ssh://builder@myserver --no-link ~/code/myproject

# Copy results to custom location
om ci run --on ssh://builder@myserver --out-link ./build-results.json ~/code/myproject
```

### Direct SSH Remote CI (Legacy) {#remote-direct}

The `--remote` flag provides simple direct SSH execution without caching:

```sh
om ci run --remote myname@myserver ~/code/myproject
```

This mode directly executes build commands via SSH without the flake caching optimization. It's simpler but less efficient for repeated builds.

**When to use `--on` vs `--remote`:**
- Use `--on` (metadata-based): For most remote CI scenarios, especially with repeated builds or when network efficiency matters
- Use `--remote` (direct SSH): For quick one-off builds or when you don't need caching overhead

## Examples

Some real-world examples of how `om ci` is used with specific configurations:

- [omnix](https://github.com/saberzero1/omnix/blob/5322235ce4069e72fd5eb477353ee5d1f5100243/nix/modules/om.nix#L16-L33)
- [services-flake](https://github.com/juspay/services-flake/blob/197fc1c4d07d09f4e01dd935450608c35393b102/flake.nix#L10-L24)
- [nixos-flake](https://github.com/srid/nixos-flake/blob/4af32875e7cc6df440c5f5cf93c67af41902768b/flake.nix#L29-L45)
- [haskell-flake](https://github.com/srid/haskell-flake/blob/d128c7329bfc73c3eeef90f6d215d0ccd7baf78c/flake.nix#L15-L67)
    - Here's [a blog post](https://twitter.com/sridca/status/1763528379188265314) that talks about how it is used in haskell-flake
- [superposition](https://github.com/juspay/superposition/blob/5eeb498cb351d958c923874afafe8a21127ac8ce/nix/om.nix)
- [haskell-rust-ffi-template](https://github.com/shivaraj-bh/haskell-rust-ffi-template/blob/f38b383b2a12afcb069ef38142faa99bb4b726f4/nix/om.nix)

## What it does {#mech}

- Check that the Nix version is not tool old (using [`om health`](health.md))
- Determine the list of flakes in the repo to build
  - By default, this is the root flake.
  - The user can also explicitly specify multiple sub-flakes in `om.ci.default` output of their root flake.
- For each (sub)flake identified, `om ci run` will run the following steps:
    - Check that `flake.lock` is up to date, if applicable.
    - Build all flake outputs, using [devour-flake](https://github.com/srid/devour-flake)[^schema]
      - Then, print the built store paths to stdout
    - If the `flake-check` step is enabled ([example](https://github.com/saberzero1/omnix/pull/376/files)), run `nix flake check`
    - Run user defined [custom steps](#custom)

[^schema]: Support for [flake-schemas](https://github.com/srid/devour-flake/pull/11) is planned

[devour-flake]: https://github.com/srid/devour-flake

## See also

- [github-nix-ci](https://github.com/juspay/github-nix-ci) - A simple NixOS & nix-darwin module for self-hosting GitHub runners
- [jenkins-nix-ci](https://github.com/juspay/jenkins-nix-ci) - Jenkins NixOS module that supports `nixci` (predecessor of `om ci`) as a Groovy function
- [cachix-push](https://github.com/juspay/cachix-push) - A flake-parts module that provides an app to enable whitelisted pushing and pinning of store paths to cachix.
