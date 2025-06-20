# Define different shells.
{
  lib,
  pkgs,
  inputs,
  ...
}:
let

  pkgsPinned = {
    go = pkgs.go_1_24;
  };

  # Override the buildGoModule function to use the specified Go package.
  buildGoModule = pkgs.buildGoModule.override { inherit (pkgsPinned) go; };
  buildWithSpecificGo = pkg: pkg.override { inherit buildGoModule; };
  buildWithSpecificGoL = pkg: pkg.override { buildGoLatestModule = buildGoModule; };

  # Creates some options in the devenv module system to be used
  # for quitsh.
  quitshDevenvModule =
    { lib, config, ... }:
    let
      cfg = config.quitsh;
    in
    {
      options.quitsh = {
        toolchain = lib.mkOption {
          type = lib.types.listOf lib.types.str;
          default = [ ];
          description = ''
            The toolchain name quitsh uses to detect the correct toolchain.
          '';
        };
      };

      config = {
        env = {
          QUITSH_TOOLCHAINS = "${lib.concatStringsSep "," cfg.toolchain}";
        };
      };
    };

  toolchains =
    let
      build-go = [
        (
          { config, ... }:
          {
            packages = [
              pkgs.quitsh.bootstrap
              pkgsPinned.go

              # For tests.
              pkgs.process-compose
              pkgs.skopeo
            ];

            env = {
              # Important to set these variables.
              GOROOT = pkgsPinned.go + "/share/go";
              GOPATH = config.env.DEVENV_STATE + "/go";
              GOTOOLCHAIN = "local";
            };

            quitsh.toolchain = [ "build-go" ];
          }
        )
      ];

      lint-go = [
        {
          quitsh.toolchain = [ "lint-go" ];

          packages = [
            pkgs.quitsh.bootstrap
            pkgs.golangci-lint
            pkgsPinned.go
          ];
        }
      ];

      general =
        build-go
        ++ lint-go
        ++ [
          {
            quitsh.toolchain = [ "general" ];

            env = {
              GOTOOLCHAIN = "local";
            };

            packages = [
              pkgs.quitsh.bootstrap

              # Linting and LSP and debuggers.
              (buildWithSpecificGo pkgs.delve)
              (buildWithSpecificGo pkgs.gotools)
              (buildWithSpecificGoL pkgs.gopls)
              (buildWithSpecificGo pkgs.golines)
              pkgs.golangci-lint-langserver
              pkgs.typos-lsp
            ];

            # To make CGO and the debugger delve work.
            # https://nixos.wiki/wiki/Go#Using_cgo_on_NixOS
            # Note: Due to warning when compilin `_FORTIFY_SOURCE`
            languages.go.enableHardeningWorkaround = true;
          }
        ];

      ci = [
        {
          packages = [
            pkgs.quitsh.bootstrap
            # pkgs.quitsh.cli
          ];

          env = {
            GOTOOLCHAIN = "local";
          };

          quitsh.toolchain = [ "ci" ];
        }
      ] ++ build-go;
    in
    {
      inherit
        build-go
        lint-go
        general
        ci
        ;
    };

  # Make a devenv shell from some modules.
  makeShell =
    let
      # This is currently needed for devenv to properly run in pure hermetic
      # mode while still being able to run processes & services and modify
      # (some parts) of the active shell.
      # We read here the root for devenv from the workaround flake input `devenv-root`.
      root = lib.strings.trim (builtins.readFile inputs.devenv-root.outPath);

    in
    pkgs: devenvModules:
    inputs.devenv.lib.mkShell {
      inherit inputs pkgs;
      modules = [
        quitshDevenvModule
        (
          args:
          {
            devenv.flakesIntegration = lib.hiPrio false;
          }
          // lib.optionalAttrs (root != "") {
            # Only apply it if `devenv-root` is defined.
            devenv.root = root;
          }
        )
      ] ++ devenvModules;
    };

in
{
  default = makeShell pkgs toolchains.general;
  ci = makeShell pkgs toolchains.ci;
  build-go = makeShell pkgs toolchains.build-go;
  lint-go = makeShell pkgs toolchains.lint-go;
}
