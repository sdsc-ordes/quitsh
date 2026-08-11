{ config, ... }:
let
  inherit (config.flake.lib) yaml;
in
{
  # Defines a common build function to build quitsh instances more easily.
  flake.lib.build.buildQuitsh =
    {
      lib,
      buildGoModule,
      installShellFiles,
      testers,
      git,

      # The source `fileset` of the quitsh instance.
      name ? "quitsh",
      # The root directory (Git repo root.).
      rootDir ? "",
      # The optional full source fileset (see `lib.fileset`).
      source ? null,
      # The Go module root relative to the source.
      modRoot ? "tools/quitsh",
      # The Go dependency vendor hash. If empty vendoring is used!.
      vendorHash ? null,
      # Skip some go tests.
      skippedTests ? [ ],
      # The module path to the `buildVersion` which is set at compile time like
      # `github.com/sdsc-ordes/quitsh/pkg/build.buildVersion`.
      versionModulePath ? "",
      # Add meta stuff.
      meta ? { },
      # Additional Go module args to merge.
      buildGoModuleArgs ? { },
    }:
    let
      fs = lib.fileset;

      src =
        assert lib.assertMsg (source != null || rootDir != "") "Source or root dir must be given.";
        if source == null then
          let
            files = fs.fromSource rootDir;
            test = fs.fromSource (rootDir + "/test");
          in
          fs.toSource {
            root = rootDir;
            fileset = fs.difference files test; # Remove integration tests.
          }
        else
          source;

      version = (yaml.readSimple (rootDir + "/.component.yaml") [ "version" ]).version;

      cli = buildGoModule (
        lib.recursiveUpdate {
          inherit name;
          pname = name;
          inherit src modRoot;

          inherit vendorHash;
          proxyVendor = true;

          nativeBuildInputs = [ installShellFiles ];
          nativeCheckInputs = [ git ];

          ldflags = lib.optionals (versionModulePath != "") [
            "-s"
            "-w"
            "-X ${versionModulePath}/pkg/build.buildVersion=${version}"
          ];

          checkFlags =
            let
              # Disable tests requiring integration tools
              inherit skippedTests;
            in
            [ "-skip=^${builtins.concatStringsSep "$|^" skippedTests}$" ];

          postInstall = ''
            installShellCompletion --cmd cli \
              --bash <($out/bin/${name} completion bash) \
              --fish <($out/bin/${name} completion fish) \
              --zsh <($out/bin/${name} completion zsh)
          '';

          passthru.tests = lib.optionalAttrs (versionModulePath != "") {
            version = testers.testVersion {
              package = cli;
              command = "${name} --version";
              inherit version;
            };
          };

          meta = {
            mainProgram = name;
          }
          // meta;
        } buildGoModuleArgs
      );
    in
    cli;
}
