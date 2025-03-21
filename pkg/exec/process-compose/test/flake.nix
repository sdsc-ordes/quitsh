{
  description = "test-process-compose";

  nixConfig = {
    allow-import-from-derivation = "true";
  };

  inputs = {
    # Nixpkgs
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

    # The devenv module to create good development shells.
    devenv = {
      url = "github:cachix/devenv";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    devenv-root = {
      url = "file+file:///dev/null";
      flake = false;
    };
  };
  outputs =
    { nixpkgs, devenv, ... }@inputs:
    let
      inherit (nixpkgs) lib;

      supportedSystems = [
        "x86_64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
      ];

      loadNixpgs =
        system:
        import inputs.nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };

      # Function which generates an attribute set '{ x86_64-linux = func {inherit lib pkgs}; ... }'.
      forAllSystems =
        func:
        lib.genAttrs supportedSystems (
          system:
          let
            pkgs = loadNixpgs system;
            lib = pkgs.lib;
          in
          func { inherit lib pkgs system; }
        );

      # Define a devShell for testing with mongodb service.
      makeShell =
        pkgs:
        let
          # This is currently needed for devenv to properly run in pure hermetic
          # mode while still being able to run processes & services and modify
          # (some parts) of the active shell.
          # We read here the root for devenv from the workaround flake input `devenv-root`.
          root = lib.strings.trim (builtins.readFile inputs.devenv-root.outPath);
        in
        inputs.devenv.lib.mkShell {
          inherit pkgs inputs;
          modules = [
            (args: {
              devenv.root = lib.mkIf (root != "") root;

              services.httpbin = {
                enable = true;
                bind = [
                  "127.0.0.1:9912"
                ];
              };

            })
          ];
        };

    in
    {
      devShells = forAllSystems (
        { pkgs, ... }:
        {
          test = makeShell pkgs;
        }
      );

      legacyPackages = forAllSystems (
        { pkgs, ... }:
        {
          mynamespace.shells.test = makeShell pkgs;
        }
      );
    };
}
