{
  config,
  inputs,
  ...
}:
{
  # Define different shells.
  perSystem =
    {
      self',
      pkgs,
      ...
    }:
    let
      inherit (config.flake.lib) mkShell;

      pkgsPinned = {
        go = pkgs.go_1_26;
      };

      toolchains =
        let
          build-go = [
            {
              packages = [
                # For tests.
                pkgs.process-compose
                pkgs.imgpkg
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
                pkgs.golangci-lint
                pkgsPinned.go
              ];
            }
          ];

          lint-trivy = [
            {
              quitsh.toolchains = [ "lint-trivy" ];
              packages = [
                pkgs.trivy
              ];
            }
          ];

          ci = [
            {
              packages = [
                self'.packages.bootstrap
                pkgs.openssh # For tests.
              ];

              quitsh.languages.go.enable = true;
              quitsh.toolchains = [ "ci" ];

              # Disable all process-compose stuff.
              # Set to native manager to not have PC_ env. variables.
              process.manager.implementation = "native";
            }
          ]
          ++ build-go;

          general =
            ci
            ++ lint-go
            ++ [
              {
                quitsh.toolchains = [ "general" ];

                quitsh.config = "tools/configs/quitsh/config.yaml";
                quitsh.configUser = "tools/configs/quitsh/config.user.yaml";

                quitsh.languages.go.enable = true;

                packages = [
                  self'.packages.bootstrap

                  pkgs.golangci-lint-langserver
                  pkgs.typos-lsp

                  pkgs.hyperfine
                ];
              }

            ];

        in
        {
          inherit
            build-go
            lint-go
            general
            ci
            lint-trivy
            ;
        };

      # Make a devenv shell from some modules.
      makeShell =
        modules:
        mkShell {
          inherit
            inputs
            pkgs
            modules
            ;
        };

    in
    {
      devShells = {
        default = makeShell toolchains.general;
        ci = makeShell toolchains.ci;
        build-go = makeShell toolchains.build-go;
        lint-go = makeShell toolchains.lint-go;
        lint-trivy = makeShell toolchains.lint-trivy;
      };
    };
}
