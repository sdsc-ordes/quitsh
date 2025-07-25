{
  description = "quitsh";

  nixConfig = {
    extra-trusted-substituters = [
      # Nix community's cache server
      "https://nix-community.cachix.org"
      "https://devenv.cachix.org"
    ];
    extra-trusted-public-keys = [
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
      "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw="
    ];

    allow-import-from-derivation = "true";
  };

  inputs = {
    # Nixpkgs
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

    # The devenv module to create good development shells.
    devenv = {
      url = "github:cachix/devenv/main";
      inputs.nixpkgs.follows = "nixpkgsDevenv";
    };
    # We have to lock somehow the pkgs in `mkShell` here:
    # https://github.com/cachix/devenv/issues/1797
    # `nixpkgs` is used in the devShell modules.
    nixpkgsDevenv.url = "github:cachix/devenv-nixpkgs/rolling";
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
