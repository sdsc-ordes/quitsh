{ remarshal, runCommand, ... }:
let
  # Read a YAML file into a Nix datatype using IFD
  # (Import From Derivation).
  #
  # Similar to:
  # > builtins.fromJSON (builtins.readFile ./somefile)
  # but takes an input file in YAML instead of JSON.
  #
  # readYAML :: Path -> a
  #
  # where `a` is the Nixified version of the input file.
  # TODO: https://github.com/NixOS/nix/pull/7340
  #       When this is merged we can replace this.
  read =
    path:
    let
      jsonOutputDrv = runCommand "from-yaml" {
        nativeBuildInputs = [ remarshal ];
      } "remarshal -if yaml -i \"${path}\" -of json -o \"$out\"";
    in
    builtins.fromJSON (builtins.readFile jsonOutputDrv);
in
{
  inherit read;
}
