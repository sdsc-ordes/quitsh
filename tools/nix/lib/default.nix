{
  lib,
  inputs,
  ...
}:
{
  # Make a devenv shell from some modules.
  # makeShell =
  #   pkgs: devenvModules:
  #   let
  #     # This is currently needed for devenv to properly run in pure hermetic
  #     # mode while still being able to run processes & services and modify
  #     # (some parts) of the active shell.
  #     # We read here the root for devenv from the workaround flake input `devenv-root`.
  #     root = lib.strings.trim (builtins.readFile inputs.devenv-root.outPath);
  #   in
  #   inputs.devenv.lib.mkShell {
  #     inputs = builtins.trace inputs.self inputs;
  #     inherit pkgs;
  #     modules = [
  #       (
  #         { ... }:
  #         {
  #           devenv.flakesIntegration = lib.hiPrio false;
  #         }
  #         // lib.optionalAttrs (root != "") {
  #           # Only apply it if `devenv-root` is defined.
  #           devenv.root = root;
  #         }
  #       )
  #     ] ++ devenvModules;
  #   };
}
