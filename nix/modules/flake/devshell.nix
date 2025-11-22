let
  root = ../../..;
in
{
  imports = [
    (root + /crates/omnix-health/module/flake-module.nix)
  ];

  perSystem = { config, self', pkgs, ... }:
    let
      # Import environment variables for consistency
      envVars = import "${root}/nix/envs" {
        inherit (config.rust-project) src;
        inherit (pkgs) cachix fetchFromGitHub lib;
      };
    in
    {
      devShells.default = pkgs.mkShell {
        name = "omnix-devshell";
        meta.description = "Omnix development environment";
        inputsFrom = [
          config.pre-commit.devShell
          self'.devShells.rust
        ];
        inherit (config.rust-project.crates."omnix-cli".crane.args)
          DEVOUR_FLAKE
          NIX_SYSTEMS
          DEFAULT_FLAKE_SCHEMAS
          FLAKE_METADATA
          FLAKE_ADDSTRINGCONTEXT
          INSPECT_FLAKE
          TRUE_FLAKE
          FALSE_FLAKE
          OMNIX_SOURCE
          OM_INIT_REGISTRY
          CACHIX_BIN
          ;

        packages = with pkgs; [
          just
          nixd
          bacon
          cargo-expand
          cargo-nextest
          cargo-audit
          cargo-workspaces
          trunk
        ];

        # Set up environment for Go development
        # These variables make the same paths available in the dev shell
        shellHook = ''
          # Environment variables for Go development (matches Nix build)
          export DEFAULT_FLAKE_SCHEMAS="${envVars.DEFAULT_FLAKE_SCHEMAS}"
          export INSPECT_FLAKE="${envVars.INSPECT_FLAKE}"
        
          echo "Omnix development environment"
          echo "- DEFAULT_FLAKE_SCHEMAS: $DEFAULT_FLAKE_SCHEMAS"
          echo "- INSPECT_FLAKE: $INSPECT_FLAKE"
        '';
      };
    };
}
