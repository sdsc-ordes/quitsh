package cli

import (
	"strings"

	sets "github.com/sdsc-ordes/quitsh/pkg/common/set"
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"
	nixtoolchain "github.com/sdsc-ordes/quitsh/pkg/toolchain/nix"

	"github.com/hashicorp/go-version"
)

type Option func(c *cliApp) error

// WithVersion setts the version on the CLI application.
func WithVersion(version *version.Version) Option {
	return func(c *cliApp) error {
		c.settings.Version = version

		return nil
	}
}

// WithName setts the name on the CLI application.
func WithName(name string) Option {
	return func(c *cliApp) error {
		c.settings.Name = name

		return nil
	}
}

// WithDescription setts the description on the CLI application.
func WithDescription(desc string) Option {
	return func(c *cliApp) error {
		c.settings.Description = desc

		return nil
	}
}

// WithStageTypes sets the stage types you want to use over all the project.
// Each stage also comes with a priority, so you have to order them here
// accordingly how they would appear in execution order.
// This is used internally to guard wrong configuration
// and may be used for sorting targets into stage.
func WithStages(ss ...stage.Stage) Option {
	return func(c *cliApp) error {
		all := sets.NewUnordered[stage.Stage]()

		for i := range ss {
			s := stage.StagePrio{Stage: ss[i], Priority: i}

			if all.Exists(s.Stage) {
				return errors.New("stages are not unique: '%v'", ss)
			}

			all.Insert(s.Stage)
			c.stages = append(c.stages, s)
		}

		return nil
	}
}

// WithComponentPatterns adds glob patterns
// e.g. includes or excludes with `!component-a`
// which will be added by default to the `FindComponents` command.
func WithComponentPatterns(patterns ...string) Option {
	return func(c *cliApp) error {
		c.componentPatterns = patterns

		return nil
	}
}

// WithTargetToStageMapper installs a custom target name to stage mapper.
// This is useful to not name the `stage` keyword in every target config.
func WithTargetToStageMapper(mapper stage.TargetNameToStageMapper) Option {
	return func(c *cliApp) error {
		if mapper == nil {
			return errors.New("target name to stage mapper must not be nil")
		}
		c.targetNameToStageMapper = mapper

		return nil
	}
}

// WithTargetToStageMapper installs a default target name to stage mapper.
// This is useful to not name the `stage` keyword in every target config.
//   - If the target name contains a suffix equal to a stage name,
//     this stage name will be default assigned (if not set).
func WithTargetToStageMapperDefault() Option {
	return func(c *cliApp) error {
		c.targetNameToStageMapper = func(targetName string) (stage.Stage, error) {
			debug.Assert(len(c.stages) != 0, "stages are not set")

			for i := range c.stages {
				if strings.HasSuffix(targetName, string(c.stages[i].Stage)) {
					return c.stages[i].Stage, nil
				}
			}

			return "", nil
		}

		return nil
	}
}

// WithToolchainDispatcher sets a toolchain dispatcher which is used to
// dispatches to the toolchain a runner specifies.
func WithToolchainDispatcher(dispatcher toolchain.IDispatcher) Option {
	return func(c *cliApp) error {
		c.toolchainDispatcher = dispatcher

		return nil
	}
}

// WithToolchainDispatcherNix sets a toolchain dispatcher which will execute a
// runner by launching the following command inside a Nix development shell with name
// `toolchain`:
// ```shell
//
//	argsv[0] execute --running-in-toolchain --config temp.yaml
//
// ```
// Where the parameters for `exec` are marshaled to `temp.yaml`.
// NOTE: When you use this option, you need to add the `execute` command
// with `exec.AddCmd` to the root command.
func WithToolchainDispatcherNix(
	flakeDir string,
	argsSelector nixtoolchain.ArgsSelector,
) Option {
	return func(c *cliApp) error {
		cmd := []string{"exec-runner", "--skip-toolchain-dispatch"}
		disp := nixtoolchain.NewDispatcher(flakeDir, cmd, argsSelector)
		c.toolchainDispatcher = &disp

		return nil
	}
}

// WithConfigFilename sets the components config filename which is used
// to find components, default is `comp.ConfigFileName`.
func WithConfigFilename(filename string) Option {
	return func(c *cliApp) error {
		c.configFilename = filename

		return nil
	}
}
