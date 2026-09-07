package cache

import (
	"path"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"deedles.dev/xiter"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"

	"github.com/spf13/cobra"
)

func AddDownloadCmd(cli cli.ICLI, parent *cobra.Command, nixSetts *config.NixSettings) {
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download all flake outputs into the local Nix cache.",
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return download(cli, nixSetts)
		},
	}

	parent.AddCommand(downloadCmd)
}

func download(cli cli.ICLI, nixSett *config.NixSettings) error {
	rootDir := cli.RootDir()
	flakePath := path.Join(rootDir, nixSett.FlakeDirRel)

	log.Infof("Downloading Flake outputs in '%v'.", flakePath)
	packages, err := nix.GetFlakeOutputs(rootDir, flakePath, []string{"devShells", "packages"})
	if err != nil {
		return err
	}

	log.Infof("Downloading '%v' Flake outputs.", len(packages))
	log.Info("Packages: \n" + formatPackage(packages))

	if nixSett.Cache.SSH.HostName == "" {
		return errors.New("Nix cache host name not set.")
	}

	switch {
	case nixSett.Cache.SSH.Enable:
		err = sshDownload(rootDir, packages, nixSett, flakePath)
		if err != nil {
			return err
		}
	default:
		return errors.New("Only SSH is currently supported. Its not enabled.")
	}

	return nil
}

func sshDownload(
	rootDir string,
	packages []*nix.Package,
	nixSett *config.NixSettings,
	flakePath string,
) error {
	sshSett := &nixSett.Cache.SSH
	nixB, agent, err := sshSetup(rootDir, sshSett, false)
	if err != nil {
		return err
	}
	defer func() { _ = agent.Close() }()
	nixCtx := nix.AddFlakeDefaultArguments(rootDir, nixB.BaseArgs("build")).Build()

	url := sshSett.URL(false)
	buildCmd := make([]string, 0, len(packages)+3) //nolint:mnd
	buildCmd = append(
		buildCmd,
		"--extra-substituters", url,
		"-v", "-L",
		"--no-eval-cache",
		"--no-link",
		"--cores",
		strconv.Itoa(runtime.NumCPU()),
	)

	ps := xiter.SortedFunc(slices.Values(packages), func(a *nix.Package, b *nix.Package) int {
		return strings.Compare(a.Name, b.Name)
	})

	for p := range ps {
		buildCmd = append(buildCmd, nix.FlakeInstallable(flakePath, p.AttrPath))
	}

	log.Infof("Building all installables from '%v' to local cache...", url)

	err = nixCtx.Check(buildCmd...)
	if err != nil {
		log.Warnf("Some derivations could not be built from the remote." +
			"Usually means that they are not in the remote store.")
	}

	return nil
}
