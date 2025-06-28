{ lib, config, ... }:
let
  cfg = config.quitsh;
in
{
  options = {
    quitsh = {
      toolchains = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        description = ''
          The toolchain names quitsh uses to detect the correct toolchain.
          Corresponds to an env. variable `QUITSH_TOOLCHAINS` which get populated.
        '';
      };
    };
  };

  config = {
    env = {
      QUITSH_TOOLCHAINS = "${lib.concatStringsSep "," cfg.toolchains}";
    };
  };
}
