# Define different shells.
{
  pkgs,
  ...
}:
{
  # Toolchain Shells ============================
  go = pkgs.mkShell {
    QUITSH_TOOLCHAINS = "go";
    packages = [
      pkgs.go
    ];
  };
  # =============================================
}
