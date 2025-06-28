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
    devenv = {
      url = "github:cachix/devenv/latest";
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

    quitsh = {
      url = "github:sdsc-ordes/quitsh?dir=tools/nix&ref=feat/add-stdin-parse-and-key-value-pairs";
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
