{ lib, self, ... }:
{
  perSystem =
    { pkgs, ... }:
    {
      packages.cli = pkgs.callPackage self.lib.build.buildQuitsh {
        name = "cli";
        buildGoModule = pkgs.buildGo126Module;

        rootDir = ../../..;
        modRoot = "./tools/cli";
        vendorHash = "sha256-SLlo47cxD+nzV7VbYBQHFsOpdqJV2xIXi1/iA8IwoF8=";

        versionModulePath = "github.com/sdsc-ordes/quitsh/pkg/build.buildVersion";

        meta = {
          description = "The quitsh's own CLI tool to build itself.";
          homepage = "https://github.com/sdsc-ordes/quitsh";
          license = lib.licenses.agpl3Plus;
          maintainers = [ "gabyx" ];
        };
      };
    };
}
