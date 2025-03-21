{
  description = "quitsh";

  nixConfig = {
    extra-trusted-substituters = [
      # Nix community's cache server
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];

    allow-import-from-derivation = "true";
  };

  inputs = {
    # Nixpkgs
    nixpkgs.url = "github:nixos/nixpkgs/nixos-24.11";

    # The devenv module to create good development shells.
    devenv = {
      url = "github:cachix/devenv?ref=v1.4.1";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    devenv-root = {
      url = "file+file:///dev/null";
      flake = false;
    };

    # Format the repo with nix-treefmt.
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # Snowfall provides a structured way of creating a flake output.
    # Documentation: https://snowfall.org/guides/lib/quickstart/
    snowfall-lib = {
      url = "github:snowfallorg/lib";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs:
    let
      root-dir = ../..;
    in
    inputs.snowfall-lib.mkFlake {
      inherit inputs;

      # The `src` must be the root of the flake.
      src = "${root-dir}";

      snowfall = {
        root = "${root-dir}" + "/tools/nix";
        namespace = "quitsh";
        meta = {
          name = "quitsh";
          title = "quitsh";
        };
      };
    };
}
