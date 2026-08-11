{ lib, ... }:
let
  # Reads only some top-level single-line `keywords` from the YAML file
  # at `path`. This is a lightweight parser to avoid IFD.
  readSimple =
    path: keywords:
    let
      lines = lib.splitString "\n" (lib.readFile path);

      findKeyword =
        keyword:
        let
          res = builtins.match "${keyword}: (.*)" (lib.findFirst (l: lib.hasPrefix "${keyword}:" l) "" lines);
        in
        assert lib.assertMsg (res != null) "Keyword '${keyword}' not found!";
        builtins.elemAt res 0;
    in
    lib.foldl' lib.mergeAttrs { } (
      lib.map (k: {
        "${k}" = findKeyword k;
      }) keywords
    );
in
{
  flake.lib.yaml = { inherit readSimple; };
}
