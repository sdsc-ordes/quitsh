{
  lib,
  config,
  ...
}:
let
  cfg = config.quitsh;
in
{
  options = {
    quitsh.config = lib.mkOption {
      type = lib.types.pathWith {
        inStore = false;
        absolute = null;
      };
      default = "";
      description = ''
        The config file used by default (same as `--config`, must exist).
        If a relative path is given it is searched in all parents.
      '';
    };

    quitsh.configUser = lib.mkOption {
      type = lib.types.pathWith {
        inStore = false;
        absolute = null;
      };
      default = "";
      description = ''
        The user config file used by default (same as `--config-user`, may not exists).
        If a relative path is given it is searched in all parents.
      '';
    };
  };

  config = {
    env =
      lib.optionalAttrs (cfg.config != "") {
        QUITSH_CONFIG = cfg.config;
      }
      // lib.optionalAttrs (cfg.configUser != "") {
        QUITSH_CONFIG_USER = cfg.configUser;
      };
  };
}
