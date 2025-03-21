{
  lib,
  remarshal,
  runCommand,
  buildGo123Module,
  installShellFiles,
  testers,
  git,
  self,
}:
let
  name = "cli";
  yaml = import ./yaml.nix { inherit remarshal runCommand; };
  fs = lib.fileset;

  files = fs.fromSource ../../..;
  test = fs.fromSource ../../../test;
  src = fs.toSource {
    root = ../../..;
    fileset = fs.difference files test;
  };
in
buildGo123Module rec {
  pname = name;
  version = (yaml.read ../../../.component.yaml).version;
  inherit src;

  modRoot = "./tools/cli";
  # subPackages = [ "tools/cli/cmd/cli" ];

  vendorHash = "sha256-dGdC34S+IWA25cGW/CBZ8yrhMQ73OW3G3fq9fUQFYiU=";
  proxyVendor = true;

  nativeBuildInputs = [ installShellFiles ];
  nativeCheckInputs = [ git ];

  ldflags =
    let
      modulePath = "custodian/tools/custodian-cli";
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
    package = self;
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
}
