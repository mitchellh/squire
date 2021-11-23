{
  description = "squire";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    nix-pgquarrel.url = "github:mitchellh/nix-pgquarrel";
    nix-pgquarrel.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    # Nix flake is Linux-only because my pg-quarrel package is linux only
    # for now. I'd be happy to add more systems if we can find a way to test it.
    flake-utils.lib.eachSystem ["aarch64-linux" "x86_64-linux"] (system:
      let
        # Our in-repo overlay of packages
        overlay = (import ./nix/overlay.nix) nixpkgs;

        # Initialize our package repository, adding overlays from inputs
        pkgs = import nixpkgs {
          inherit system;

          overlays = [
            inputs.nix-pgquarrel.overlay.${system}
            overlay
          ];
        };

        repo = pkgs.callPackage ./nix/repo.nix {
          inherit pkgs;
        };
      in rec {
        devShell = repo.shell;
        packages.squire = repo.package;
        defaultPackage = packages.squire;
        checks.fmt = repo.fmtcheck;
      }
    );
}
