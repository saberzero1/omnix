{ inputs, ... }:
# Nix module for the Go part of the project
#
# This module provides the Go build for omnix v2.0.0
{
  perSystem = { config, self', pkgs, lib, system, ... }:
    let
      # Import environment variables from nix/envs
      envVars = import "${inputs.self}/nix/envs" {
        inherit (config.rust-project) src;
        inherit (pkgs) cachix fetchFromGitHub lib;
      };
    in
    {
      packages = {
        # Go version of omnix (v2.0.0)
        omnix-go = pkgs.buildGo123Module rec {
          pname = "omnix";
          version = "2.0.0-beta";
          src = lib.cleanSource inputs.self;

          # vendorHash computed by Nix (set to lib.fakeHash, build, then use reported hash)
          vendorHash = "sha256-fw5op35m+fp0PGR60tqXuU6t0f4KMKw19ip3RTCiibc=";

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
