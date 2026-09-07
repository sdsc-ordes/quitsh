package cache

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"deedles.dev/xiter"

	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	"github.com/sdsc-ordes/quitsh/pkg/exec/ssh"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"

	"github.com/spf13/cobra"
)

func AddUploadCmd(cli cli.ICLI, parent *cobra.Command, nixSetts *config.NixSettings) {
	uploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload all flake outputs to the Nix cache.",
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return upload(cli, nixSetts)
		},
	}

	parent.AddCommand(uploadCmd)
}

func upload(cli cli.ICLI, nixSett *config.NixSettings) error {
	rootDir := cli.RootDir()
	flakePath := path.Join(rootDir, nixSett.FlakeDirRel)

	log.Infof("Uploading Flake outputs in '%v'.", flakePath)
	packages, err := nix.GetFlakeOutputs(rootDir, flakePath, []string{"devShells", "packages"})
	if err != nil {
		return err
	}

	log.Infof("Uploading '%v' Flake outputs.", len(packages))
	log.Info("Packages: \n" + formatPackage(packages))

	if nixSett.Cache.SSH.HostName == "" {
		return errors.New("Nix cache host name not set.")
	}

	switch {
	case nixSett.Cache.SSH.Enable:
		err = sshUpload(rootDir, packages, nixSett)
		if err != nil {
			return err
		}
	default:
		return errors.New("Only SSH is currently supported. Its not enabled.")
	}

	return nil
}

func formatPackage(packages []*nix.Package) string {
	var s strings.Builder
	for _, p := range packages {
		fmt.Fprintf(&s, " • '%v'\n   '%v'\n   '%v'\n", p.Name, p.AttrPath, p.StorePath)
	}

	return s.String()
}

func sshUpload(rootDir string, packages []*nix.Package, nixSett *config.NixSettings) error {
	sshSett := &nixSett.Cache.SSH
	nixB, agent, err := sshSetup(rootDir, sshSett, true)
	if err != nil {
		return err
	}
	defer func() { _ = agent.Close() }()
	nixCtx := nixB.Build()
	nixBuildCtx := nix.NewBuildCtx(rootDir)

	url := sshSett.URL(true)
	copyCmd := make([]string, 0, len(packages)+3) //nolint:mnd
	copyCmd = append(copyCmd, "copy", "--to", url)

	buildCmd := make([]string, 0, len(packages))
	buildCmd = append(buildCmd, "--no-link")

	ps := xiter.SortedFunc(slices.Values(packages), func(a *nix.Package, b *nix.Package) int {
		return strings.Compare(a.Name, b.Name)
	})

	for p := range ps {
		buildCmd = append(buildCmd,
			nix.FlakeInstallable(nixSett.FlakeDirRel, p.AttrPath))
		copyCmd = append(copyCmd, p.StorePath)
	}

	log.Info("Building installables (if not in cache)...")
	err = nixBuildCtx.Check(buildCmd...)
	if err != nil {
		return err
	}

	log.Infof("Copying installables to remote '%v'...", url)

	return nixCtx.Check(copyCmd...)
}

//nolint:nestif // Complexity is ok here.
func sshSetup(
	rootDir string,
	sshSett *config.NixSSH,
	write bool,
) (nixB exec.CmdContextBuilder, agent ssh.Agent, err error) {
	if ci.IsRunning() {
		log.Infof("Adding know host key '%v' '%v'.", sshSett.HostName, sshSett.HostPublicKey)
		err = ssh.AddKnownHost(sshSett.HostName, sshSett.HostPublicKey)
		if err != nil {
			return nixB, agent, err
		}
	}

	user := &sshSett.Read
	if write {
		user = &sshSett.Write
	}

	nixB = exec.NewCmdCtxBuilder().Cwd(rootDir)

	if user.UseAgent {
		var e error
		agent, e = ssh.StartAgent(log.Global(), ssh.WithNewInstance(true))
		if e != nil {
			return nixB, agent, e
		}

		// Add the socket to the context.
		nixB.Env(agent.Env()...)
		sshCtx := nixB.Build()

		envKey, e := base64.RawStdEncoding.DecodeString(os.Getenv(user.PrivateKeyEnv))
		if e != nil {
			e = errors.New("Could not base64 decode env. var '%s'.", user.PrivateKeyEnv)

			return nixB, agent, e
		}

		switch {
		case user.PrivateKeyPath != "":
			log.Infof("Adding key '%v' with 'ssh-add'.", user.PrivateKeyPath)
			e = sshCtx.Check("ssh-add", user.PrivateKeyPath)
			if e != nil {
				return nixB, agent, e
			}

		case len(envKey) != 0:
			log.Infof("Adding key from env. var '%v' with 'ssh-add'.", user.PrivateKeyEnv)
			// Add a newline just in case none is there, to not fail.
			e = sshCtx.
				WithStdin(strings.NewReader(string(envKey)+"\n")).
				Check("ssh-add", "-")
			if e != nil {
				return nixB, agent, e
			}
		default:
			log.Warnf("No SSH key given (key path '%s' or env. '%s'), "+
				"expecting SSH key to be added to agent.", user.PrivateKeyPath, user.PrivateKeyEnv)
			if ci.IsRunning() {
				return nixB, agent, errors.New("SSH keys must be given.")
			}
		}
	}

	nixB.BaseCmd("nix")

	return nixB, agent, nil
}
