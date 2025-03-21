{
  inputs,
  pkgs,
  ...
}:
let
  # Configure formatter.
  treefmtEval = inputs.treefmt-nix.lib.evalModule pkgs ./treefmt.nix;
  treefmt = treefmtEval.config.build.wrapper;
in
treefmt
