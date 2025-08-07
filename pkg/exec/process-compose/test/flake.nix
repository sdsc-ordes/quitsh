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
      url = "github:cachix/devenv/main";
      inputs.nixpkgs.follows = "nixpkgs-devenv";
    };
    # We have to lock somehow the pkgs in `mkShell` here:
    # https://github.com/cachix/devenv/issues/1797
    # `nixpkgs` is used in the devShell modules.
    nixpkgs-devenv.url = "github:cachix/devenv-nixpkgs/rolling";
    devenv-root = {
      url = "file+file:///dev/null";
      flake = false;
    };

  };
  outputs =
    { nixpkgs, ... }@inputs:
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

              process.managers.process-compose = {
                package = pkgs.process-compose;
              };

              # Has no ready probe.
              services.httpbin = {
                enable = true;
                bind = [
                  "127.0.0.1:9912"
                ];
              };

              # Has ready probe.
              processes = {
                keycloak = {
                  exec = "${pkgs.coreutils}/bin/tail -f /dev/null";
                  process-compose = {
                    readiness_probe.exec.command = "${pkgs.coreutils}/bin/true";
                    depends_on.httpbin.condition = "process_started";
                  };
                };
                completed = {
                  exec = "${pkgs.coreutils}/bin/true";
                };
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
