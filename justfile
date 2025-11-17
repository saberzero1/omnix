# Documentation targets
mod doc

default:
    @just --list

# Run all pre-commit hooks on all files
pca:
    pre-commit run -a

# Run omnix-cli locally (Rust)
watch *ARGS:
    bacon --job run -- -- {{ ARGS }}

run *ARGS:
    cargo run -p omnix-cli {{ ARGS }}

alias w := watch

# Run CI locally
[group('ci')]
ci:
    nix --accept-flake-config run . ci

# Run CI locally in devShell (using cargo)
[group('ci')]
ci-cargo:
    cargo run -p omnix-cli -- ci run

# Run CI locally in devShell (using cargo) on a simple flake with subflakes
[group('ci')]
ci-cargo-ext:
    cargo run -p omnix-cli -- ci run github:srid/nixos-unified

# Do clippy checks for all crates
[group('ci-steps')]
clippy:
    cargo clippy --release --locked --all-targets --all-features --workspace -- --deny warnings

# Build cargo doc for all crates
[group('ci-steps')]
cargo-doc:
    cargo doc --release --all-features --workspace

# Run cargo test for all crates
[group('ci-steps')]
cargo-test:
    cargo test --release --all-features --workspace

# Go targets

# Run om binary locally (Go)
[group('go')]
go-run *ARGS:
    go run ./cmd/om {{ ARGS }}

# Build Go binary
[group('go')]
go-build:
    go build -o bin/om ./cmd/om

# Run Go tests
[group('go')]
go-test:
    go test -v -race ./...

# Run Go tests with coverage
[group('go')]
go-test-coverage:
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run Go linter
[group('go')]
go-lint:
    golangci-lint run

# Format Go code
[group('go')]
go-fmt:
    go fmt ./...
    goimports -w .

# Run full Go CI (lint, test, build)
[group('go')]
go-ci: go-fmt go-lint go-test-coverage go-build
