# Define different shells.
{
  lib,
  pkgs,
  inputs,
  cli,
  ...
}:
let
  go = pkgs.go_1_23;

  toolchains = {
    go = [
      (args: {
        packages = [
          pkgs.git
          go
          # For tests.
          pkgs.process-compose
        ];

        env = {
          QUITSH_TOOLCHAINS = "go";
        };
      })
    ];

    general = [
      (args: {
        packages = [
          # Essentials.
          pkgs.git
          pkgs.just

          # Main languages.
          go

          # Linting and LSP and debuggers.
          pkgs.delve
          pkgs.gopls
          pkgs.golines
          pkgs.typos-lsp

          # cli

          pkgs.process-compose
        ];

        languages.go.enableHardeningWorkaround = true;

        env = {
          QUITSH_TOOLCHAINS = "general";
        };
      })
    ];

    ci = [
      (args: {
        packages = [
          pkgs.git
          pkgs.git-lfs
          pkgs.just
          cli
        ];

        env = {
          QUITSH_TOOLCHAINS = "ci";
        };
      })
    ];
  };

  # Make a devenv shell from some modules.
  makeShell =
    devenvModules:
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
        (
          { ... }:
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
  ci = makeShell toolchains.ci;
  default = makeShell toolchains.general;
  go = makeShell toolchains.go;
}
