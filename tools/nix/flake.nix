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
      url = "github:cachix/devenv";
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
  };

  outputs =
    inputs:
    let
      inherit (inputs.self) outputs;
      inherit (inputs.nixpkgs) lib;

      supportedSystems = [
        "x86_64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
      ];

      # Import nixpkgs and load it into
      # pkgs and apply overlays to it.
      loadNixpgs =
        system:
        import inputs.nixpkgs {
          inherit system;
          overlays = [ ];
        };

      forEachSupportedSystem =
        func:
        lib.genAttrs supportedSystems (
          system:
          let
            pkgs = loadNixpgs system;
            lib = pkgs.lib;
          in
          func { inherit lib pkgs system; }
        );

      defineTreefmt = pkgs: (import ./packages/treefmt) { inherit pkgs inputs; };

    in
    {
      formatter = forEachSupportedSystem ({ pkgs, ... }: defineTreefmt pkgs);

      packages = forEachSupportedSystem (
        { pkgs, ... }:
        let
          # Define our CLI tool.
          cli = pkgs.callPackage ./packages/cli { self = cli; };

        in
        {
          treefmt = defineTreefmt pkgs;
          inherit cli;
        }
      );

      devShells = forEachSupportedSystem (
        { pkgs, system, ... }:
        import ./shells.nix {
          inherit lib pkgs inputs;
          inherit (outputs.packages.${system}) cli;
        }
      );
    };
}
