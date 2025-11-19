{ inputs, ... }:
# Nix module for the Go part of the project
#
# This module provides the Go build for omnix v2.0.0
{
  perSystem = { config, self', pkgs, lib, system, ... }: {
    packages = {
      # Go version of omnix (v2.0.0)
      omnix-go = pkgs.buildGo123Module rec {
        pname = "omnix";
        version = "2.0.0-beta";
        src = lib.cleanSource inputs.self;

        # Computed vendorHash from: nix hash path vendor
        vendorHash = "sha256-fw5op35m+fp0PGR60tqXuU6t0f4KMKw19ip3RTCiibc=";

        # Disable CGO for static linking
        CGO_ENABLED = 0;

        # Build flags
        ldflags = [
          "-s"
          "-w" # Strip debug symbols
          "-X main.Version=${version}"
          "-X main.Commit=${inputs.self.rev or inputs.self.dirtyRev or "dev"}"
          "-X main.BuildTime=1970-01-01T00:00:00Z"
        ];

        # Only build the main binary
        subPackages = [ "cmd/om" ];

        # Install shell completions
        nativeBuildInputs = [ pkgs.installShellFiles ];
        postInstall = ''
          # Generate shell completions
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

    # Set the Go version as default for v2.0
    # Keep Rust version available as omnix-cli during transition
    # packages.default will be set in rust.nix or here based on preference
  };
}
