{
  description = "squire";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    nix-pgquarrel.url = "github:mitchellh/nix-pgquarrel";
    nix-pgquarrel.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    flake-utils.lib.eachDefaultSystem (system:
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
      in {
        devShell = repo.shell;
      }
    );
}
