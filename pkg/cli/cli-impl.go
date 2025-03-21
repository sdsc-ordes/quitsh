package cli

import (
	rootcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/root"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	"github.com/spf13/cobra"
)

func (c *cliApp) Config() config.IConfig {
	return c.config
}

func (c *cliApp) ToolchainDispatcher() toolchain.IDispatcher {
	return c.toolchainDispatcher
}

func (c *cliApp) RunnerFactory() factory.IFactory {
	return c.factory
}

func (c *cliApp) RootDir() string {
	root, err := c.resolveRootDir()
	if err != nil {
		log.PanicE(err, "could not determine root dir, something is wrong")
	}

	return root
}

func (c *cliApp) RootCmd() *cobra.Command {
	return c.rootCmd
}

func (c *cliApp) RootArgs() *rootcmd.Args {
	return c.rootArgs
}

func (c *cliApp) Stages() stage.Stages {
	return c.stages
}

func (c *cliApp) Run() error {
	if err := c.rootCmdPreExec(); err != nil {
		return err
	}

	return c.rootCmd.Execute()
}

func (c *cliApp) FindComponents(
	args *general.ComponentArgs,
) (comps []*component.Component, all []*component.Component, rootDir string, err error) {
	rootDir = c.RootDir()

	outBaseDir := ""
	if c.rootArgs.GlobalOutputDir != "" {
		outBaseDir = c.rootArgs.GlobalOutputDir
	} else if c.rootArgs.GlobalOutput {
		outBaseDir = rootDir
	}

	transformConfig := mapTargetNameToStage(c.targetNameToStageMapper, nil)
	transformConfig = setStagePrio(c.stages, transformConfig)

	comps, all, err = general.FindComponents(
		args,
		c.rootArgs.Cwd,
		outBaseDir,
		c.componentPatterns,
		transformConfig)

	return
}

func (c *cliApp) resolveRootDir() (string, error) {
	r := c.rootArgs

	if c.rootDirResolved {
		return r.RootDir, nil
	}

	// Set the global root dir if not set already.
	if r.RootDir == "" {
		var err error
		_, r.RootDir, err = git.NewCtxAtRoot(r.Cwd)
		if err != nil {
			return "", errors.AddContext(
				err,
				"could not determine root directory from Git top-level",
			)
		}
	} else {
		r.RootDir = fs.MakeAbsolute(r.RootDir)
	}

	if !fs.Exists(r.RootDir) {
		return "", errors.New("root directory '%v' doest not exist", r.RootDir)
	}

	log.Debug("Root directory resolved.", "root", r.RootDir)

	c.rootDirResolved = true

	return r.RootDir, nil
}

func setStagePrio(
	stages stage.Stages,
	mixing component.ConfigAdjuster,
) component.ConfigAdjuster {
	return func(conf *component.Config) (err error) {
		if mixing != nil {
			err = mixing(conf)
			if err != nil {
				return
			}
		}

		for _, target := range conf.Targets {
			s, exists := stages.Find(target.Stage)
			target.StagePrio = s

			if !exists {
				return errors.New(
					"target id '%v' contains an unknown stage '%v' (not in '%v')",
					target.ID,
					target.Stage,
					stages,
				)
			}
		}

		return nil
	}
}

func mapTargetNameToStage(
	mapTargetToStage stage.TargetNameToStageMapper,
	mixing component.ConfigAdjuster,
) component.ConfigAdjuster {
	if mapTargetToStage == nil {
		return mixing
	}

	return func(conf *component.Config) (err error) {
		if mixing != nil {
			err = mixing(conf)
			if err != nil {
				return
			}
		}

		for n, target := range conf.Targets {
			if target.Stage != "" {
				continue
			}

			target.Stage, err = mapTargetToStage(n)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
