{ lib, pkgs, ... }:
let
  quitsh-log = pkgs.writeShellApplication {
    name = "log";
    text = builtins.readFile ./scripts/log.sh;
    runtimeInputs = [
      pkgs.coreutils
    ];
  };
in
{
  options = {
    quitsh.log.package = lib.mkOption {
      type = lib.types.package;
      default = quitsh-log;
      description = ''
        The quitsh devShell log derivation containing `bin/log`
        Useful to log in `enterShell` and other places.
      '';
    };
  };
}
