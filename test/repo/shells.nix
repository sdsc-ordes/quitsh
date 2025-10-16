# Define different shells.
{
  inputs,
  pkgs,
  ...
}:
{
  # Toolchain Shells ============================
  go = inputs.quitsh.lib.mkShell {
    inherit inputs pkgs;
    modules = [
      {
        quitsh.languages.go.enable = true;
        quitsh.toolchains = [ "go" ];
      }
    ];
  };

  cmd-runner = inputs.quitsh.lib.mkShell {
    inherit inputs pkgs;
    modules = [
      {
        quitsh.toolchains = [ "cmd-runner" ];
        packages = [
          pkgs.coreutils
        ];
      }
    ];
  };
  # =============================================
}
