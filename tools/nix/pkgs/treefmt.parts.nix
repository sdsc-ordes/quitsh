{ inputs, ... }:
{
  imports = [
    inputs.treefmt-nix.flakeModule
  ];

  perSystem =
    { config, pkgs, ... }:
    let
      treefmt = config.treefmt.build.wrapper;
    in
    {
      # Define formatter for `nix fmt`.
      formatter = treefmt;

      packages = {
        inherit treefmt;
      };

      treefmt = {
        inherit pkgs;
        # Used to find the project root
        # For worktrees we need either `.git` or a file.
        projectRootFile = "README.md";

        settings.global = {
          excludes = [ ];
        };

        # Enable the following formatters.
        programs.gofmt.enable = true;
        programs.goimports.enable = true;
        programs.golines.enable = true;

        # Markdown, JSON, YAML, etc.
        programs.prettier.enable = true;
        settings.formatter.prettier.excludes = [
          ".golangci.yaml" # this is a symlink, which prettier cannot deal with
          ".yamllint.yaml" # this is a symlink, which prettier cannot deal with
        ];

        # Shellscripts (which we should not have!)
        programs.shfmt = {
          enable = true;
          indent_size = 4;
        };
        programs.shellcheck = {
          enable = true;
        };
        settings.formatter.shellcheck = {
          options = [
            "-e"
            "SC1091"
          ];
        };

        # Nix.
        programs.deadnix.enable = false;
        programs.statix.enable = false;
        programs.nixfmt.enable = true;
      };
    };
}
