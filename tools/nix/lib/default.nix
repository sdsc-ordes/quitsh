{
  lib,
  ...
}:
let
  # Devenv modules for Quitsh.
  devenvModules = [
    (import ./devenv/go.nix)
    (import ./devenv/toolchains.nix)
  ];
in
{
  devenv = {
    # All quitsh's devenv modules.
    modules = devenvModules;
  };

  # Make a devenv shell (from `inputs.devenv`) with the following features:
  # - `devenv.root` set to the flake input `devenv-root` defined by
  #
  #    ```nix
  #       devenv-root = {
  #         url = "file+file:///dev/null";
  #         flake = false;
  #       };
  #     ```
  #   and set with `nix develop --no-pure-eval --override-input devenv-root "path:.devenv/state/pwd"
  #   where `.devenv/state/pwd` is the current project root directory. This is a workaround to allow
  #   pure evaluation.
  # - Using `pkgs` or import from flake input `inputs.nixpkgs-devenv` for the `system` if not given.
  # - All quitsh's `devenvModules` applied.
  # - Additional `modules` added.
  # - Allow unfree packages.
  # - Flake integration set to `true`
  mkShell =
    {
      inputs,
      modules ? [ ],
      pkgs ? null,
      system ? null,
    }:
    let
      # This is currently needed for devenv to properly run in pure hermetic
      # mode while still being able to run processes & services and modify
      # (some parts) of the active shell.
      # We read here the root for devenv from the workaround flake input `devenv-root`.
      root = lib.strings.trim (builtins.readFile inputs.devenv-root.outPath);

      pkgsForDevenv =
        if pkgs == null then
          let
            assertNotNull = lib.assertMsg system != null "System must be given";
          in
          import inputs.nixpkgs-devenv {
            config.allowUnfree = true;
            inherit system;
          }
        else
          pkgs;

    in
    inputs.devenv.lib.mkShell {
      inherit inputs;
      pkgs = pkgsForDevenv;

      modules =
        devenvModules
        ++ [
          (
            {
              devenv.flakesIntegration = true;
            }
            // lib.optionalAttrs (root != "") {
              # Only apply it if `devenv-root` is defined.
              devenv.root = root;
            }
          )
        ]
        ++ modules;
    };
}
