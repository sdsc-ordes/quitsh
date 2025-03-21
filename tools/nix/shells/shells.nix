# Define different shells.
{
  lib,
  pkgs,
  inputs,
  ...
}:
let
  go = pkgs.go_1_23;

  toolchains = {
    go = [
      (
        { ... }:
        {
          packages = [
            pkgs.git
            go
            # For tests.
            pkgs.process-compose
          ];

          env = {
            QUITSH_TOOLCHAINS = "go";
          };
        }
      )
    ];

    general = [
      (
        { ... }:
        {
          packages = [
            pkgs.quitsh.bootstrap

            # Main languages.
            go

            # Linting and LSP and debuggers.
            pkgs.delve
            pkgs.gopls
            pkgs.golines
            pkgs.typos-lsp

            # Our build tool (quitsh framework).
            pkgs.quitsh.cli

            pkgs.process-compose
          ];

          languages.go.enableHardeningWorkaround = true;

          env = {
            QUITSH_TOOLCHAINS = "general";
          };
        }
      )
    ];

    ci = [
      (
        { ... }:
        {
          packages = [
            pkgs.quitsh.bootstrap
            pkgs.quitsh.cli
          ];

          env = {
            QUITSH_TOOLCHAINS = "ci";
          };
        }
      )
    ];
  };

  # Make a devenv shell from some modules.
  makeShell =
    pkgs: devenvModules:
    let
      # This is currently needed for devenv to properly run in pure hermetic
      # mode while still being able to run processes & services and modify
      # (some parts) of the active shell.
      # We read here the root for devenv from the workaround flake input `devenv-root`.
      root = lib.strings.trim (builtins.readFile inputs.devenv-root.outPath);
    in
    inputs.devenv.lib.mkShell {
      inherit inputs pkgs;
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
  default = makeShell pkgs toolchains.general;
  ci = makeShell pkgs toolchains.ci;
  go = makeShell pkgs toolchains.go;
}
