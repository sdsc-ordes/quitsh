{
  inputs,
  pkgs,
  lib,
  namespace,
  ...
}:
(import ../shells.nix {
  inherit
    inputs
    pkgs
    lib
    namespace
    ;
}).build-go
