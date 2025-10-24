# Define different shells.
{
  lib,
  pkgs,
  inputs,
  ...
}:
let

  pkgsPinned = {
    go = pkgs.go_1_24;
  };

  toolchains =
    let
      build-go = [
        {
          packages = [
            pkgs.quitsh.bootstrap

            # For tests.
            pkgs.process-compose
            pkgs.skopeo
          ];

          quitsh.languages.go = {
            enable = true;
            package = pkgsPinned.go;
          };

          quitsh.toolchains = [ "build-go" ];
        }
      ];

      lint-go = [
        {
          quitsh.toolchains = [ "lint-go" ];

          packages = [
            pkgs.quitsh.bootstrap
            pkgs.golangci-lint
            pkgsPinned.go
          ];
        }
      ];

      general =
        build-go
        ++ lint-go
        ++ [
          {
            quitsh.toolchains = [ "general" ];

            quitsh.config = "tools/configs/quitsh/config.yaml";
            quitsh.configUser = "tools/configs/quitsh/config.user.yaml";

            quitsh.languages.go.enable = true;

            packages = [
              pkgs.quitsh.bootstrap

              pkgs.golangci-lint-langserver
              pkgs.typos-lsp

              pkgs.hyperfine
            ];
          }

        ];

      ci = [
        {
          packages = [
            pkgs.quitsh.bootstrap
            # pkgs.quitsh.cli
          ];

          quitsh.languages.go.enable = true;
          quitsh.toolchains = [ "ci" ];
        }
      ] ++ build-go;
    in
    {
      inherit
        build-go
        lint-go
        general
        ci
        ;
    };

  # Make a devenv shell from some modules.
  makeShell =
    modules:
    lib.quitsh.mkShell {
      inherit
        inputs
        pkgs
        modules
        ;
    };

in
{
  default = makeShell toolchains.general;
  ci = makeShell toolchains.ci;
  build-go = makeShell toolchains.build-go;
  lint-go = makeShell toolchains.lint-go;
}
