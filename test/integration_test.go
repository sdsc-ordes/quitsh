//go:build test && integration

package test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (quitsh exec.CmdContextBuilder) {
	binDir := os.Getenv("QUITSH_BIN_DIR")
	require.DirExists(t, binDir, "QUITSH_BIN_DIR=%s must exist.", binDir)

	quitshExe := path.Join(binDir, "quitsh-integration-test")
	require.FileExists(t, quitshExe, "exe '%v' must exist.", quitshExe)

	covDir := os.Getenv("QUITSH_COVERAGE_DIR")
	require.DirExists(t, covDir, "QUITSH_COVERAGE_DIR=%s must exist.", covDir)

	// Remove output
	f := path.Join("repo/component-a/.output")
	err := os.RemoveAll(f)
	require.NoError(t, err)

	return exec.NewCmdCtxBuilder().
		BaseCmd(quitshExe).
		BaseArgs("--root-dir", "repo").
		Cwd(".").
		EnableCaptureError().
		Env("GOCOVERDIR=" + covDir)
}

func TestCLIVersion(t *testing.T) {
	cli := setup(t).Build()
	stdout, err := cli.Get("--version")

	require.NoError(t, err)
	assert.Contains(t, stdout, fmt.Sprintf("version %v", build.GetBuildVersion()))
}

func TestCLIList(t *testing.T) {
	cli := setup(t).Build()

	stdout, err := cli.Get("list", "-C", "../test", "--output", "-")

	require.NoError(t, err)
	assert.Contains(t, stdout, "component-a")
}

func TestCLIExecTarget(t *testing.T) {
	cli := setup(t).Build()

	_, stderr, err := cli.GetStdErr(
		"exec-target",
		"--tag", "do-echo",
		"--log-level",
		"debug",
		"component-a::build",
	)

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.NotContains(t, stderr, "Hurrey building release version")
	assert.NotContains(t, stderr, "excluded-step-should-not-be-run")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}

func TestCLIExecTargetWithExcludedStep(t *testing.T) {
	cli := setup(t).Build()

	_, stderr, err := cli.GetStdErr(
		"exec-target",
		"--tag", "do-echo",
		"--tag", "only-when-a",
		"--log-level",
		"debug",
		"component-a::build",
	)

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.NotContains(t, stderr, "Hurrey building release version")
	assert.Contains(t, stderr, "excluded-step-should-not-be-run")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}

func TestCLISetConfigValues(t *testing.T) {
	cli := setup(t).Build()

	_, stderr, err := cli.GetStdErr(
		"exec-target",
		"--log-level",
		"debug",
		"-k", "build.buildType: release",
		"component-a::build",
	)

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.Contains(t, stderr, "Hurrey building release version")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}

func TestCLISetConfigValuesStdin(t *testing.T) {
	cli := setup(t).Build()

	config := "build:\n  buildType: release"
	r := strings.NewReader(config)

	_, stderr, err := cli.WithStdin(r).GetStdErr(
		"--config", "-",
		"exec-target",
		"--log-level",
		"debug",
		"component-a::build",
	)

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.Contains(t, stderr, "Hurrey building release version")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}

func TestCLIExecTarget2(t *testing.T) {
	cli := setup(t).Build()

	_, stderr, err := cli.GetStdErr(
		"exec-target",
		"--log-level",
		"debug",
		"component-a::build-banana",
	)

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.Contains(t, stderr, "🌻")
	assert.NotContains(t, stderr, "Nix set argument: '")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}

func TestCLIExecTargetParallel(t *testing.T) {
	cli := setup(t).Build()

	_, stderr, err := cli.GetStdErr(
		"exec-target",
		"--log-level",
		"debug",
		"--parallel",
		"component-a::build-banana",
		"component-a::lint",
	)

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}

func TestCLIExecTarget2Arg(t *testing.T) {
	cli := setup(t).Env(
		"QUITSH_NIX_NO_PURE_EVAL=true",
		"MYARG=banana").Build()

	_, stderr, err := cli.GetStdErr(
		"exec-target",
		"--log-level",
		"debug",
		"component-a::build-banana",
	)

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
	assert.Contains(t, stderr, "Nix set argument: 'banana'")
}

func TestCLIProcessCompose(t *testing.T) {
	_, rootDir, err := git.NewCtxAtRoot(".")
	require.NoError(t, err)

	cwd := path.Join(rootDir, "pkg/exec/process-compose/test")
	cli := setup(t).Cwd(cwd).BaseArgs("--root-dir", cwd).Build()

	_, stderr, err := cli.GetStdErr(
		"--root-dir", ".",
		"--log-level",
		"debug",
		"process-compose",
		"start",
		"--flake-dir", ".",
		"--wait-for", "httpbin",
		"mynamespace.shells.test",
	)
	require.NoError(t, err, "Stderr:\n"+stderr)
	assert.Contains(t, stderr, "Inspect processes with")
	assert.Contains(t, stderr, "Stop processes with")

	stdout, _, err := cli.GetStdErr(
		"--root-dir", ".",
		"--log-level",
		"debug",
		"process-compose",
		"exec",
		"--flake-dir", ".",
		"mynamespace.shells.test",
		"list",
	)
	require.NoError(t, err, "Process compose list failed")
	assert.Contains(t, stdout, "httpbin")

	defer func() {
		_, _, err := cli.GetStdErr(
			"--root-dir", ".",
			"--log-level",
			"debug",
			"process-compose",
			"stop",
			"--flake-dir", ".",
			"mynamespace.shells.test",
		)
		require.NoError(t, err, "Could not stop process-compose.")
	}()

}
