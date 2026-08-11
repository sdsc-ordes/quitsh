{
  ...
}:
{
  flake.lib.devenv.modules = [
    (import ./go.nix)
    (import ./toolchains.nix)
    (import ./log.nix)
    (import ./config.nix)
  ];
}
