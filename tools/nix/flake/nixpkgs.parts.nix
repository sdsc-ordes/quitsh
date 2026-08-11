{
  self,
  ...
}:
{
  perSystem =
    {
      system,
      ...
    }:
    let
      pkgs = self.lib.nixpkgs.importPkgs { inherit system; };
    in
    {
      # Define two arguments `pkgs` and `pkgsStable` available on all flake-parts modules.
      _module.args.pkgs = pkgs;

      legacyPackages.unstable = pkgs;
    };
}
