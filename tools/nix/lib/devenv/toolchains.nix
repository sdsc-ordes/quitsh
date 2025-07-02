{
  lib,
  pkgs,
  config,
  ...
}:
let
  cfg = config.quitsh;

  merge = pkgs.writeShellApplication {
    name = "merge";
    text = builtins.readFile ./scripts/merge-toolchain-env.sh;
    runtimeInputs = [
      pkgs.coreutils
      pkgs.gnused
    ];
  };
in
{
  options = {
    quitsh = {
      toolchains = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        description = ''
          The toolchain names ('[a-z0-9-]+') quitsh uses to detect the correct toolchain.
          An env. variable `QUITSH_TOOLCHAINS` will get populated by these names.
          Works with nesteed shells.
        '';
      };
    };
  };

  config =
    let
      toolchains = "${lib.concatStringsSep "," cfg.toolchains}";
    in
    {
      assertions = [
        {
          assertion = lib.lists.all (n: (lib.strings.match "[a-z0-9-]+" n != null)) cfg.toolchains;
          message = ''
            Toolchain names do not comply with `[a-z]+`: ${toolchains}
          '';
        }
      ];

      enterShell = ''
        export QUITSH_TOOLCHAINS=$("${merge}/bin/merge" "${toolchains}")
        ${cfg.log.package}/bin/log info "Quitsh toolchains active: '$QUITSH_TOOLCHAINS'";
      '';
    };
}
