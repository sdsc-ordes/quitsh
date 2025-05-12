{
  lib,
  remarshal,
  runCommand,
  buildGo124Module,
  installShellFiles,
  testers,
  git,
}:
let
  name = "cli";
  yaml = import ./yaml.nix { inherit remarshal runCommand; };
  fs = lib.fileset;

  rootDir = ../../../..;

  files = fs.fromSource rootDir;
  test = fs.fromSource (rootDir + "/test");
  src = fs.toSource {
    root = rootDir;
    fileset = fs.difference files test;
  };

  cli = buildGo124Module rec {
    pname = name;
    version = (yaml.read (rootDir + "/.component.yaml")).version;
    inherit src;

    modRoot = "./tools/cli";

    vendorHash = "sha256-3T9TE2Igyg+UpJ1SWzQDmoRVAd7Jfxm+RGYB/oU9ADA=";
    proxyVendor = true;

    nativeBuildInputs = [ installShellFiles ];
    nativeCheckInputs = [ git ];

    ldflags =
      let
        modulePath = "github.com/sdsc-ordes/quitsh";
      in
      [
        "-s"
        "-w"
        "-X ${modulePath}/pkg/build.buildVersion=${version}"
      ];

    checkFlags =
      let
        # Disable tests requiring integration tools
        skippedTests = [
          "TestProcessComposeDevenv"
        ];
      in
      [ "-skip=^${builtins.concatStringsSep "$|^" skippedTests}$" ];

    postInstall = ''
      installShellCompletion --cmd cli \
        --bash <($out/bin/${name} completion bash) \
        --fish <($out/bin/${name} completion fish) \
        --zsh <($out/bin/${name} completion zsh)
    '';

    passthru.tests.version = testers.testVersion {
      package = cli;
      command = "${name} --version";
      inherit version;
    };

    meta = with lib; {
      description = "The quitsh's own CLI tool to build itself.";
      homepage = "https://data-custodian.gitlab.io/custodian";
      license = licenses.agpl3Plus;
      maintainers = [ "gabyx" ];
      mainProgram = name;
    };
  };
in
cli
