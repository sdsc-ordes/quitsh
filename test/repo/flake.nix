{
  description = "component-a-test";

  nixConfig = {
    extra-trusted-substituters = [
      # Nix community's cache server
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];

    allow-import-from-derivation = "true";
  };

  inputs = {
    # Nixpkgs
    nixpkgs.url = "github:nixos/nixpkgs/nixos-24.11";
  };
  outputs =
    inputs:
    let
      inherit (inputs.nixpkgs) lib;

      supportedSystems = [
        "x86_64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
      ];

      # Import nixpkgs and load it into
      # pkgs and apply overlays to it.
      loadNixpgs =
        system:
        let
          # Testing setting an argument.
          sys =
            let
              a = builtins.getEnv "MYARG";
            in
            if a != "" then builtins.warn "Nix set argument: '${a}'" system else system;
        in
        import inputs.nixpkgs {
          system = sys;
          overlays = [ ];
        };

      forEachSupportedSystem =
        func:
        lib.genAttrs supportedSystems (
          system:
          let
            pkgs = loadNixpgs system;
            lib = pkgs.lib;
          in
          func { inherit lib pkgs; }
        );
    in
    {
      devShells = forEachSupportedSystem ({ pkgs, ... }: import ./shells.nix { inherit pkgs; });
    };
}
