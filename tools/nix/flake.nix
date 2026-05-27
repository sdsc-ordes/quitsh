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
    # The `nixpkgs-devenv` must aligned with the pinned version.
    devenv = {
      url = "github:cachix/devenv?ref=v2.1.2";
      inputs.nixpkgs.follows = "nixpkgs-devenv";
    };
    # This is the rolling nixpkgs with what devenv was tested.
    nixpkgs-devenv = {
      url = "github:cachix/devenv-nixpkgs?ref=ec3063523dcd911aeadb50faa589f237cdab5853";
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
