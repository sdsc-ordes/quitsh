package rootcmd

import (
	"fmt"
	"io"
	"os"

	e "errors"

	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Root arguments.
// NOTE: ALl fields need proper default values (here mostly empty).
type Args struct {
	// The config YAML from which we read parameters for the CLI.
	Config string `yaml:"-"`
	// The config YAML (user overlay).
	ConfigUser string `yaml:"-"`
	// Config key,value arguments to override nested config
	// values by paths e.g. `a.b.c: {"a":3}` on the command line.
	ConfigKeyValues []string `yaml:"-"`

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

// Create a new `quitsh` root command with settings `setts` and
// root arguments `rootArgs`. The full argument structure `allArgs` is treated
// as `any` and will be used to parse the configuration files `--config`
// (`--config-user`) into before startup.
//
// WARNING: The default values here set and in any subcommand, are unimportant!
// We load the config (the ground truth) always afterwards and before
// any arguments are parsed.
// The sequence is as follows:
//   - Cobra sets defaults valutes in command definitions (unimportant).
//   - The `preExecFunc`: We pass the config (hopefully defaulted),
//     load potentially from `--config`, `--config-user` and `--config-values`.
//   - Cobra executes and sets CLI arguments to override stuff as a final step.
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

	addPersistendFlags(rootCmd.PersistentFlags(), rootArgs)

	rootCmd.Flags().
		BoolVar(&version, "version", version, "Print the version.")

	rootCmd.SilenceErrors = true

	preExecFunc = func() error {
		var err error
		parsedConfig, parsedConfigUser, err = parseConfigs(config)

		return err
	}

	return rootCmd, preExecFunc
}

func addPersistendFlags(flags *pflag.FlagSet, args *Args) {
	flags.
		StringVar(&args.Config, "config", args.Config,
			"The global configuration file. If set to '-' then stdin is read.")
	flags.
		StringVar(&args.ConfigUser, "config-user", args.ConfigUser,
			"The global user configuration file (overlay), can not exist.")
	flags.
		StringArrayVarP(
			&args.ConfigKeyValues,
			"config-value",
			"k",
			args.ConfigKeyValues,
			"Config key,values pairs to override nested config values, e.g. `\"a.b.c: {\\\"a\\\": 3}\"`.",
		)

	flags.
		StringVarP(&args.Cwd,
			"cwd", "C", args.Cwd,
			"Set the current working directory "+
				"(note: '--root-dir' = Git root dir evaluated from '--cwd').")
	flags.
		StringVarP(&args.RootDir,
			"root-dir", "R", args.RootDir, "Set the root directory. "+
				"This is used to define configured relative paths, e.g. flake path etc.")
	flags.
		StringVarP(&args.LogLevel,
			"log-level", "v", args.LogLevel, "The log level. [debug|info|warn|error]")
	flags.
		BoolVar(
			&args.EnableEnvPrint,
			"enable-env-print",
			args.EnableEnvPrint,
			"If env. variables printing for failed commands should be enabled.",
		)
	flags.
		BoolVar(&args.SkipToolchainDispatch,
			"skip-toolchain-dispatch", args.SkipToolchainDispatch,
			"The flag to denote that we are already inside the toolchain. "+
				"Skip any invocation to any toolchain dispatcher.",
		)

	flags.
		BoolVar(&args.GlobalOutput,
			"global-output", args.GlobalOutput,
			fmt.Sprintf("If a global output directory (repository root dir + '%s').", fs.OutputDir))

	flags.
		StringVar(&args.GlobalOutputDir,
			"global-output-dir", args.GlobalOutputDir,
			"Use this as global output directory (more simple: use '--global-output').")
}

func parseConfigs(
	conf any,
) (parsedConfig, parsedUserConfig bool, err error) {
	// Parse here the --config, and --config-user and `--config-values`
	// and init the config, because that needs to happen before
	// cobra parses the flags and set defaults.
	//
	// Priorities:
	// 2. Env. variables, not yet implemented.
	// 1. Config from `--config` and `--config-user` and `--config-values.`
	// 0. Command line arguments.

	var args Args
	s := pflag.NewFlagSet("default", pflag.ContinueOnError)
	addPersistendFlags(s, &args)

	s.ParseErrorsWhitelist.UnknownFlags = true

	err = s.Parse(os.Args)
	if e.Is(err, pflag.ErrHelp) {
		// WARN: any `-h` will get into an error. Noway to disable that.
		// So we set the error to nil, and return.
		return false, false, nil
	}

	parsedConfig, err = initConfig(args.Config, conf, true)
	if err != nil {
		return
	}

	parsedUserConfig, err = initConfig(args.ConfigUser, conf, false)
	if err != nil {
		return
	}

	err = config.ApplyKeyValues(args.ConfigKeyValues, conf)
	if err != nil {
		return
	}

	return
}

func initConfig(configPath string, conf any, errorIfNotExists bool) (bool, error) {
	if configPath == "" {
		return false, nil
	}

	log.Debugf("Parse config from '%s'", configPath)

	var f io.Reader
	switch configPath {
	case "-":
		f = os.Stdout
	default:
		exists := fs.Exists(configPath)

		if !exists {
			if errorIfNotExists {
				return false, errors.New("Config file '%s' does not exists", configPath)
			} else {
				return false, nil
			}
		}

		ff, err := os.OpenFile(configPath, os.O_RDONLY, fs.DefaultPermissionsFile)
		if err != nil {
			return false, errors.New("Could not load config file '%s'", configPath)
		}
		defer ff.Close()

		f = ff
	}

	decoder := yaml.NewDecoder(f)
	err := decoder.Decode(conf)
	if err != nil {
		return false, errors.AddContext(err, "could not decode config file '%s'", configPath)
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
