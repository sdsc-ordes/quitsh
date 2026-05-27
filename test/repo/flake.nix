{
  description = "component-a-test";

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

    quitsh = {
      url = "../../?dir=tools/nix";
    };
  };
  outputs =
    inputs:
    let
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
        let
          # Testing setting an argument.
          sys =
            let
              a = builtins.getEnv "MYARG";
            in
            if a != "" then builtins.warn "Nix set argument: '${a}'" system else system;
        in
        import inputs.nixpkgs {
          system = sys;
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
          func { inherit lib pkgs; }
        );
    in
    {
      devShells = forEachSupportedSystem (
        { pkgs, ... }:
        import ./shells.nix {
          inherit pkgs;
          inherit inputs;
        }
      );
    };
}
