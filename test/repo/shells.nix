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

  runner-exec = inputs.quitsh.lib.mkShell {
    inherit inputs pkgs;
    modules = [
      {
        quitsh.toolchains = [ "runner-exec" ];
        packages = [
          pkgs.coreutils
        ];
      }
    ];
  };
  # =============================================
}
