package rootcmd

import (
	"fmt"
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Args struct {
	// The config YAML from which we read parameters for the CLI.
	Config string `yaml:"-"`
	// The config YAML (user overlay).
	ConfigUser string `yaml:"-"`

	// Working directory to switch to at startup.
	// This will be used to search for components.
	// This will be used to define the `RootDir` directory.
	Cwd string `yaml:"cwd"`

	// The root directory of quitsh.
	// By default its the Git root directory resolved starting from
	// `Cwd`.
	RootDir string `yaml:"rootDir"`

	// The log level `debug,info,warning,error`.
	LogLevel string `yaml:"logLevel"`

	// Enable environment print on command execution errors.
	EnableEnvPrint bool `yaml:"enableEnvPrint"`

	// Disable any toolchain dispatch which might happen.
	SkipToolchainDispatch bool `yaml:"skipToolchainDispatch"`

	// If we use a global output directory
	// instead of component's specific one.
	GlobalOutput bool `yaml:"globalOutput"`
	// Use a specific output directory.
	GlobalOutputDir string `yaml:"outputDir"`
}

type Settings struct {
	Name        string
	Version     *version.Version
	Description string
}

func (s *Settings) applyDefaults() {
	if s.Name == "" {
		s.Name = "quitsh"
	}
	if s.Version == nil {
		s.Version = build.GetBuildVersion()
	}

	desc := fmt.Sprintf(
		"A script tool to support tooling in monorepos [based on quitsh version: %v].",
		build.GetBuildVersion(),
	) +
		`
    It replaces non-statically typed script languages -> Quit using 'sh'.
    `

	if s.Description == "" {
		s.Description = desc
	}
}

// Create a new `quitsh` CLI with settings `setts` and
// root arguments `rootArgs`. The full argument structure `allArgs` is treated
// as `any` and will be used to parse the configuration files `--config`
// (`--config-user`) into before startup.
func New(
	setts *Settings,
	rootArgs *Args,
	config any) (rootCmd *cobra.Command, preExecFunc func() error) {
	if setts == nil {
		setts = &Settings{}
	}
	setts.applyDefaults()

	var parsedConfig, parsedConfigUser bool
	var version bool

	rootCmd = &cobra.Command{
		Use:           setts.Name,
		Long:          setts.Description,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(_cmd *cobra.Command, _args []string) error {
			e := applyGeneralOptions(rootArgs)
			if e != nil {
				return e
			}

			if parsedConfig {
				log.Debug("Parsed general config.", "path", rootArgs.Config)
			}
			if parsedConfigUser {
				log.Debug("Parsed user config.", "path", rootArgs.ConfigUser)
			}

			log.Debug("Parsed config.", "config", config)

			return nil
		},
		RunE: func(cmd *cobra.Command, _args []string) error {
			return runRoot(cmd, setts, version)
		},
	}

	rootCmd.PersistentFlags().
		StringVar(&rootArgs.Config, "config", "", "The global configuration file")
	rootCmd.PersistentFlags().
		StringVar(&rootArgs.ConfigUser, "config-user", "", "The global user configuration file (overlay), can not exist.")
	rootCmd.PersistentFlags().
		StringVarP(&rootArgs.Cwd,
			"cwd", "C", ".",
			"Set the current working directory "+
				"(note: '--root-dir' = Git root dir evaluated from '--cwd').")
	rootCmd.PersistentFlags().
		StringVarP(&rootArgs.RootDir,
			"root-dir", "R", "", "Set the root directory. "+
				"This is used to define configured relative paths, e.g. flake path etc.")
	rootCmd.PersistentFlags().
		StringVarP(&rootArgs.LogLevel,
			"log-level", "v", "info", "The log level. [debug|info|warn|error]")
	rootCmd.PersistentFlags().
		BoolVar(&rootArgs.EnableEnvPrint,
			"enable-env-print", false, "If env. variables printing for failed commands should be enabled.")
	rootCmd.PersistentFlags().
		BoolVar(&rootArgs.SkipToolchainDispatch,
			"skip-toolchain-dispatch", false,
			"The flag to denote that we are already inside the toolchain. "+
				"Skip any invocation to any toolchain dispatcher.",
		)

	rootCmd.PersistentFlags().
		BoolVar(&rootArgs.GlobalOutput,
			"global-output", false, fmt.Sprintf("If a global output directory (repository root dir + '%s').", fs.OutputDir))

	rootCmd.PersistentFlags().
		StringVar(&rootArgs.GlobalOutputDir,
			"global-output-dir", "", "Use this as global output directory (more simple: use '--global-output').")

	rootCmd.Flags().
		BoolVar(&version, "version", false, "Print the version.")

	rootCmd.SilenceErrors = true

	preExecFunc = func() error {
		var err error

		parsedConfig, parsedConfigUser, err = parseConfigs(
			rootCmd.PersistentFlags(),
			rootArgs,
			config,
		)

		return err
	}

	return rootCmd, preExecFunc
}

func parseConfigs(
	ss *pflag.FlagSet,
	rootArgs *Args,
	config any,
) (parsedConfig, parsedUserConfig bool, err error) {
	// Parse here the --config, and --config-user
	// and init the configs, because that needs to happen before
	// cobra parses the flags.

	s := *ss // Copy to not have side effects with `.parsed`
	s.ParseErrorsWhitelist.UnknownFlags = true

	err = s.Parse(os.Args)
	if err != nil {
		// WARN: any `-h` will get into an error. Noway to disable that.
		// So we set the error to nil, and return.
		err = nil

		return //nolint:nilerr // intentional.
	}

	parsedConfig, err = initConfig(rootArgs.Config, config, true)
	if err != nil {
		return
	}

	parsedUserConfig, err = initConfig(rootArgs.ConfigUser, config, false)
	if err != nil {
		return
	}

	return
}

func initConfig(config string, args any, errorIfNotExists bool) (bool, error) {
	if config == "" {
		return false, nil
	}

	if exists := fs.Exists(config); !exists {
		if errorIfNotExists {
			return false, errors.New("Config file '%s' does not exists", config)
		} else {
			return false, nil
		}
	}

	f, err := os.OpenFile(config, os.O_RDONLY, fs.DefaultPermissionsFile)
	if err != nil {
		return false, errors.New("Could not load config file '%s'", config)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(args)
	if err != nil {
		return false, errors.AddContext(err, "could not decode config file '%s'", config)
	}

	return true, nil
}

func runRoot(rootCmd *cobra.Command, setts *Settings, version bool) error {
	if version {
		fmt.Printf( //nolint:forbidigo // Allowed as no log yet.
			"%s version %v\n",
			setts.Name, setts.Version,
		)

		return nil
	}

	_ = rootCmd.Help()

	return errors.New("no command given")
}

func applyGeneralOptions(r *Args) error {
	// NOTE: No `git` stuff should be done here.
	// As there are other commands which also need to run like `completion`
	// which does not need Git etc.
	if r.LogLevel != "" {
		err := log.SetLevel(r.LogLevel)
		if err != nil {
			return errors.AddContext(err, "could not set log level to '%v'", r.LogLevel)
		}
	}

	r.Cwd = fs.MakeAbsolute(r.Cwd)
	log.Debug("Setting working dir.", "cwd", r.Cwd)
	err := os.Chdir(r.Cwd)
	if err != nil {
		return errors.AddContext(err, "could not change working dir '%v'", r.Cwd)
	}

	if r.GlobalOutput && r.GlobalOutputDir != "" {
		return errors.New("either use '--global-output' or '--output-dir', but not both")
	}

	// Global hack to enable env. printing
	// TODO: Setting globals is not so good!
	exec.EnableEnvPrint = r.EnableEnvPrint

	return nil
}
