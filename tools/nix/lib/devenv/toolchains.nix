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

  config =
    let
      toolchains = "${lib.concatStringsSep ", " cfg.toolchains}";
    in
    {
      env = {
        QUITSH_TOOLCHAINS = toolchains;
      };

      enterShell = ''
        quitsh-log info "Entering Quitsh DevShell: Toolchains active '${toolchains}'";
      '';
    };
}
