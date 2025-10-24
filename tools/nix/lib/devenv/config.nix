{
  lib,
  config,
  pkgs,
  ...
}:
let
  cfg = config.quitsh;
in
{
  options = {
    quitsh.config = lib.mkOption {
      type = lib.types.nullOr (
        lib.types.pathWith {
          inStore = false;
          absolute = null;
        }
      );
      default = null;
      description = ''
        The config file (relative to `devenv` root) used by default (same as `--config`, must exist).
      '';
    };

    quitsh.configUser = lib.mkOption {
      type = lib.types.nullOr (
        lib.types.pathWith {
          inStore = false;
          absolute = null;
        }
      );
      default = null;
      description = ''
        The user config file (relative to `devenv` root) used by default (same as `--config-user`, may not exists).
      '';
    };
  };

  config = {
    env =
      lib.optionalAttrs (cfg.config != null) {
        QUITSH_CONFIG = "${config.devenv.root}/${cfg.config}";
      }
      // lib.optionalAttrs (cfg.configUser != null) {
        QUITSH_CONFIG_USER = "${config.devenv.root}/${cfg.configUser}";
      };
  };
}
