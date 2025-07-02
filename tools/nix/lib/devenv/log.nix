{ pkgs, ... }:
let
  quitsh-log = pkgs.writeShellApplication {
    name = "quitsh-log";
    text = builtins.readFile ./scripts/log.sh;
    runtimeInputs = [
      pkgs.coreutils
    ];
  };
in
{
  config = {
    packages = [ quitsh-log ];
  };
}
