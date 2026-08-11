package fixhash

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	strs "github.com/sdsc-ordes/quitsh/pkg/common/strings"
	"github.com/sdsc-ordes/quitsh/pkg/concurrent"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"

	"github.com/spf13/cobra"
)

const getOutputHash = `
	let
	    f = builtins.getFlake "git+file:///{{ .RepoPath }}?dir={{ .FlakeRelDir }}";
        lib = f.inputs.nixpkgs.lib;
	    drv = f.outputs.{{ .AttrPath }};
	in
	if ! lib.isDerivation drv then
    {
        hash = "none";
    }
    else if (builtins.hasAttr "goModules" drv) then
    {
        fodAttrPath = "goModules";
        hash = drv.goModules.outputHash;
    }
    else if (builtins.hasAttr "cargoDeps" drv) then
    {
        fodAttrPath = "cargoDeps.vendorStaging";
        hash = drv.cargoDeps.vendorStaging.outputHash;
    }
    else if (builtins.hasAttr "pnpmDeps" drv) then
    {
        fodAttrPath = "pnpmDeps";
        hash = drv.pnpmDeps.outputHash;
    }
    else
    {
        hash = "none";
    }
	`

const buildFixedOutputDrv = `
	let
	    f = builtins.getFlake "git+file:///{{ .RepoPath }}?dir={{ .FlakeRelDir }}";
	    drv = f.outputs.{{ .AttrPath }};
	in
	drv.overrideAttrs (f: b: { outputHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; })
	`

const longDescLint = `
Fix hashes of Nix's fixed-output-derivation on components.
`

func AddCmd(cli cli.ICLI, parent *cobra.Command, nixSetts *config.NixSettings) {
	failIfChanges := false

	fixHashCmd := &cobra.Command{
		Use:   "fix-hash",
		Short: "Fix hash on fixed-output-derivations.",
		Long:  longDescLint,
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return fixHash(cli, failIfChanges, nixSetts.FlakeDirRel)
		},
	}

	fixHashCmd.Flags().
		BoolVarP(&failIfChanges,
			"fail-on-change", "c", false, "Fail if any changes were made.")

	parent.AddCommand(fixHashCmd)
}

// fixHash will evaluate over all Nix packages and check
// on fixed-output derivation the `outputHash` and recomputes it (sets it to `lib.flakeHash`)
// re builds the derivation and get `hash mismatch` errors to
// assemble a list of `old-hash -> new-hash` replacements over all Nix files in the repository.
// This is a workaround to be able to easily update hashes like the `vendorHash`
// for example on
// - `buildGoModule` (attr `goModules` on `drv`)
// - `buildNodeModule` (attr `pnpmDeps` on `drv`):
//
// To check the outputHash on `goModules` fixed-output derivation:
// ```shell
// nix eval --expr "
//
//		let
//		  f = builtins.getFlake
//	       "git+file:///persist/repos/custodian?dir=tools/nix&ref=$(git branch --show)" +
//	       "&rev=$(git rev-parse HEAD)";
//
//		  drv = f.outputs.packages.x86_64-linux.quitsh;
//		in
//		drv.goModules.outputHash or null
//
// " --impure
// ```
// and to build the derivation with empty hash:
//
// ```shell
// nix eval --expr "
//
//		let
//		  f = builtins.getFlake
//	       "git+file:///persist/repos/custodian?dir=tools/nix&ref=$(git branch --show)" +
//	       "&rev=$(git rev-parse HEAD)";
//
//		  drv = f.outputs.packages.x86_64-linux.quitsh;
//		in
//		drv.goModules.overrideAttrs (f: b: { outputHash = \"sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"; })
//
// " --impure
// ```.
func fixHash(cli cli.ICLI, failIfChanges bool, flakeDirRel string) error {
	rootDir := cli.RootDir()
	gitx := git.NewCtx(rootDir)

	flakePath := path.Join(rootDir, flakeDirRel)
	packages, err := nix.GetFlakePackages(rootDir, flakePath)
	if err != nil {
		return err
	}

	hashOldNew := make(map[string]string)
	err = fixHashes(gitx, rootDir, packages, hashOldNew, flakeDirRel)
	// Eagerly replace all stuff, despite of errors.
	errS := searchAndReplace(rootDir, hashOldNew)
	err = errors.Combine(err, errS)

	log.Info("Tried to replace as many FOD hashes as possible.")

	if err != nil {
		return err
	} else if len(hashOldNew) != 0 && failIfChanges {
		return errors.New("not all hashes in files are up to date but replaced as much as possible")
	}

	return nil
}

// TODO: Make it less long.
//
//nolint:gocognit // FIXME: Make shorter.
func fixHashes(
	gitx git.Context,
	rootDir string,
	packages map[string]*nix.Package,
	hashOldNew map[string]string,
	flakeDirRel string,
) error {
	nixx := nix.NewCtxBuilder().Cwd(rootDir).Build()

	hasChanges, err := gitx.HasChanges(gitx.Cwd())
	if err != nil {
		return err
	}

	if hasChanges {
		log.Warn("You SHOULD have not unstaged changes before running this function!")
		log.Warn("Staging everything now.")

		e := gitx.Check("add", "-A")
		if e != nil {
			return e
		}
	}

	updateMutex := sync.Mutex{}

	fix := func(p *nix.Package) error {
		data := map[string]string{
			"AttrPath":           p.AttrPath,
			"RepoPath":           gitx.Cwd(),
			"FlakeRelDir":        flakeDirRel,
			"FixedOutputDrvName": "",
		}

		out, e := nix.EvalTemplate(
			nixx,
			getOutputHash,
			data,
			// nix.WithEvalNoCache(),
			nix.WithEvalImpure(),
			nix.WithEvalOutputJSON(),
		)
		if e != nil {
			return e
		}

		type OldHash struct {
			Hash        string `json:"hash"`
			FODAttrPath string `json:"fodAttrPath"`
		}
		var oldHash OldHash

		e = json.Unmarshal([]byte(out), &oldHash)
		if e != nil {
			return errors.New("Could not unmarshal response from Nix evaluation output: '%v'.", out)
		}

		if oldHash.Hash == "none" || oldHash.Hash == "" {
			log.Info(
				"Package has none of {'goModules', 'pnpmDeps', 'cargoDeps'} fixed-output derivations -> Skip",
				"package",
				p.AttrPath,
			)

			return nil
		}

		data["AttrPath"] = p.AttrPath + "." + oldHash.FODAttrPath
		log.Info("Build fixed-output derivation to get new hash", "package", data["AttrPath"])
		res, e := exec.WithTemplate(
			nixx,
			buildFixedOutputDrv,
			data,
			func(c *exec.CmdContext, file string) (string, error) {
				return c.GetCombined("build", "--file", file)
			},
		)
		if e == nil || !strings.Contains(res, "mismatch") || !strings.Contains(res, "hash") {
			return errors.New(
				"Building with empty hash should result in an hash mismatch error!, out:\n%s",
				res,
			)
		}

		var newHash string
		for _, l := range strs.SplitLines(res) {
			if strings.Contains(l, "got:") {
				newHash = strings.TrimSpace(strings.Replace(l, "got:", "", 1))
			}
		}

		if oldHash.Hash != newHash {
			log.Warn(
				"Fixed output derivation requires new hash.",
				"package",
				p.AttrPath,
				"oldHash",
				oldHash,
				"newHash",
				newHash,
			)

			updateMutex.Lock()
			defer updateMutex.Unlock()

			hashOldNew[oldHash.Hash] = newHash
		} else {
			log.Info(
				"Fixed output derivation has current hash.",
				"package",
				p.AttrPath,
				"hash",
				newHash,
			)
		}

		return nil
	}

	return concurrent.Map(maps.Values(packages), fix)
}

func searchAndReplace(rootDir string, hashOldNew map[string]string) error {
	if len(hashOldNew) == 0 {
		return nil
	}

	log.Info("Hashes to update:", "hashes", hashOldNew)

	files, _, err := fs.FindFiles(
		rootDir,
		fs.WithPathFilterPatterns(
			[]string{"**/*.nix"},
			[]string{".devenv/**/*", ".direnv/**/*"},
			true),
	)
	if err != nil {
		return errors.AddContext(err, "Could not find all Nix files")
	}

	log.Info("Replace over Nix files.", "count", len(files))

	correct := func(file string) error {
		log.Debugf("Replace hash in file '%s'.", file)
		bs, e := os.ReadFile(file)
		if e != nil {
			return e
		}

		replaced := false
		for old, new := range hashOldNew {
			sub := []byte(old)
			idx := bytes.Index(bs, sub)
			if idx >= 0 {
				replaced = true
				after := slices.Clone(bs[idx+len(old):])
				bs = append(bs[:idx], new...)
				bs = append(bs, after...)
			}
		}

		if replaced {
			log.Debugf("Replaced hashes in file '%s'.", file)
			err = os.WriteFile(file, bs, fs.DefaultPermissionsFile) //nolint:gosec // Ok!
			if err != nil {
				return err
			}
		}

		return e
	}

	return concurrent.Map(slices.Values(files), correct)
}
