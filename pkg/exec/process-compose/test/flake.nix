{
  description = "test-process-compose";

  nixConfig = {
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

    process-compose-flake.url = "github:Platonic-Systems/process-compose-flake";
    services-flake.url = "github:juspay/services-flake";
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
          func {
            inherit
              inputs
              lib
              pkgs
              system
              ;
          }
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
                  exec = "exec ${pkgs.coreutils}/bin/true";
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
          test-devenv = makeShell pkgs;
        }
      );

      legacyPackages = forAllSystems (
        { pkgs, ... }:
        let
          servicesMod = (import inputs.process-compose-flake.lib { inherit pkgs; }).evalModules {
            modules = [
              inputs.services-flake.processComposeModules.default
              {
                services.mailhog."mailhog" = {
                  enable = true;
                  smtp.port = 9999;
                  ui.port = 9998;
                };
              }
            ];
          };
        in
        {
          mynamespace.test-devenv = makeShell pkgs;
          mynamespace.test-process-compose-flake = servicesMod.config.outputs.package;
          mynamespace.test-process-compose-flake-config = servicesMod.config;
        }
      );
    };
}
