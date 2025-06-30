package gorunner

import (
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	gox "github.com/sdsc-ordes/quitsh/pkg/exec/go"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const GoBuildRunnerID = "quitsh::build-go"

type GoBuildRunner struct {
	runnerConfig *RunnerConfigBuild
	settings     config.IBuildSettings
}

// NewGoBuildRunner constructs a new GoBuildRunner with its own config.

func NewGoBuildRunner(config any, settings config.IBuildSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GoBuildRunner{
		runnerConfig: common.Cast[*RunnerConfigBuild](config),
		settings:     settings,
	}, nil
}

func (*GoBuildRunner) ID() runner.RegisterID {
	return GoBuildRunnerID
}

// The difference between `go install` and `go build` in Go lies in their purpose
// and the outputs they produce:
//
// ### **`go build`**
//
// - **Purpose**: Compiles the code and produces a binary.
// - **Output**:
//   - For commands (i.e., `package main`), it generates an **executable binary**
//     in the current directory (or another location if specified with `-o`).
//   - For libraries (non-main packages), it just checks for compilation errors but
//     does not produce any output file.
//
// - **Where**: Temporary directory unless explicitly specified with `-o`.
// - **Use Case**: Useful for quick testing and local compilation.
//
// ---
//
// ### **`go install`**
//
//   - **Purpose**: Compiles the code and installs the binary or library to the **Go
//     module cache** or `$GOBIN` directory (by default `$GOPATH/bin`).
//   - **Output**:
//   - For commands (i.e., `package main`), it installs the binary to `$GOBIN`.
//   - For libraries (non-main packages), it builds and caches the compiled package
//     in the module cache (`$GOMODCACHE`).
//   - **Where**: `$GOBIN` (for executables) or module cache (for libraries).
//   - **Use Case**: Used for permanent installation of executables and for caching
//     dependencies in projects.
func (r *GoBuildRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()

	log.Info("Starting Go build for component.", "component", comp.Name())

	fs.AssertDirs(comp.OutBuildBinDir())

	// Set the output path and disable GOWORK:
	// We build in each component without looking
	// at `go.work` fs.
	var binDir string
	if r.settings.Coverage() {
		binDir = comp.OutCoverageBinDir()
	} else {
		binDir = comp.OutBuildBinDir()
	}

	goctx := gox.NewCtxBuilder().
		Cwd(comp.Root()).
		Env("GOBIN="+binDir,
			"GOWORK=off",
			"GOTOOLCHAIN=local").
		Build()

	modInfo, err := gox.GetModuleInfo(comp.Root())
	if err != nil {
		return err
	}

	// Build everything into `outputDir`.
	flags, tagArgs := GetBuildFlags(
		log,
		comp.Root(),
		r.settings.BuildType(),
		r.settings.EnvironmentType(),
		r.settings.Coverage(),
		false,
		modInfo,
		comp.Version(),
		r.runnerConfig.VersionModule,
		r.runnerConfig.BuildTags,
		false,
	)

	log.Info("Run Go generate.")
	cmd := append([]string{"generate"}, tagArgs...)
	cmd = append(cmd, "./...")
	err = goctx.Check(cmd...)
	if err != nil {
		log.ErrorE(err, "Go generate failed.")

		return err
	}

	log.Info("Run Go install.")

	cmd = append([]string{"install"}, flags...)
	cmd = append(cmd, r.settings.Args()...)
	cmd = append(cmd, path.Join(comp.Root(), "..."))
	err = goctx.Check(cmd...)

	if err != nil {
		log.ErrorE(err, "Go install failed.")

		return err
	}

	return nil
}
