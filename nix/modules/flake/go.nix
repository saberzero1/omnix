{ inputs, ... }:
# Nix module for the Go part of the project
#
# This module provides the Go build for omnix v2.0.0
{
  perSystem = { config, self', pkgs, lib, system, ... }:
    let
      # Cleaned omnix repository source - includes all necessary files including flake functions
      # This is used for both the Go build and embedded paths to ensure consistency
      omnixSrc = lib.cleanSource inputs.self;

      # Import environment variables from nix/envs
      # Use omnixSrc so the embedded paths reference the same Nix store path
      # that contains the flake function directories
      envVars = import "${inputs.self}/nix/envs" {
        src = omnixSrc;
        inherit (pkgs) cachix fetchFromGitHub lib;
      };
    in
    {
      packages = {
        # Go version of omnix (v2.0.0)
        omnix-go = pkgs.buildGo123Module rec {
          pname = "omnix";
          version = "2.0.0-beta";
          src = omnixSrc;

          # vendorHash computed by Nix (set to lib.fakeHash, build, then use reported hash)
          vendorHash = "sha256-3MnvLADYS4vDYMDNAOzg/vfJtPKBgYgVl1H2uiq9lkU=";

          # Disable CGO for static linking
          CGO_ENABLED = 0;

          # Build flags - inject environment variables as compile-time constants
          ldflags = [
            "-s"
            "-w" # Strip debug symbols
            "-X main.Version=${version}"
            "-X main.Commit=${inputs.self.rev or inputs.self.dirtyRev or "dev"}"
            # Inject flake-related environment variables
            "-X github.com/saberzero1/omnix/pkg/nix/flake.defaultFlakeSchemas=${envVars.DEFAULT_FLAKE_SCHEMAS}"
            "-X github.com/saberzero1/omnix/pkg/nix/flake.inspectFlake=${envVars.INSPECT_FLAKE}"
            # Inject flake function paths (for addstringcontext and metadata)
            # These are embedded at build time so the binary works from any directory
            "-X github.com/saberzero1/omnix/pkg/nix/flake/functions.flakeMetadata=${envVars.FLAKE_METADATA}"
            "-X github.com/saberzero1/omnix/pkg/nix/flake/functions.flakeAddStringContext=${envVars.FLAKE_ADDSTRINGCONTEXT}"
            # Inject omnix source path for remote CI
            "-X github.com/saberzero1/omnix/pkg/ci.omnixSourcePath=${envVars.OMNIX_SOURCE}"
          ];

          # Only build the main binary
          subPackages = [ "cmd/om" ];

          # Install shell completions
          nativeBuildInputs = [ pkgs.installShellFiles ];
          postInstall = ''
            # Generate shell completions
            # Note: PowerShell is supported by the CLI but installShellCompletion doesn't support it
            # PowerShell users can generate completions with: om completion powershell
            installShellCompletion --cmd om \
              --bash <($out/bin/om completion bash) \
              --zsh <($out/bin/om completion zsh) \
              --fish <($out/bin/om completion fish)
          '';

          meta = with lib; {
            description = "Developer-friendly companion for Nix";
            homepage = "https://omnix.page";
            license = licenses.agpl3Only;
            maintainers = [ ];
            mainProgram = "om";
            platforms = platforms.unix;
          };
        };
      };

      # packages.default is set in rust.nix to point to omnix-go
    };
}
