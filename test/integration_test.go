//go:build test && integration

package test

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/exec"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureOutput redirects output of stderr/stdout.
func captureOutput(f func() error) (string, string, error) {
	origStdout := os.Stdout
	origStderr := os.Stderr

	rStdout, wStdout, _ := os.Pipe()
	rStderr, wStderr, _ := os.Pipe()
	defer wStdout.Close()
	defer wStderr.Close()

	os.Stdout = wStdout
	os.Stderr = wStderr
	err := f()
	os.Stdout = origStdout
	os.Stderr = origStderr
	wStdout.Close()
	wStderr.Close()

	out, _ := io.ReadAll(rStdout)
	outErr, _ := io.ReadAll(rStderr)

	return strings.TrimSpace(string(out)), strings.TrimSpace(string(outErr)), err
}

func setup(t *testing.T) (quitsh *exec.CmdContext) {
	binDir := os.Getenv("QUITSH_BIN_DIR")
	require.DirExists(t, binDir, "QUITSH_BIN_DIR=%s must exist.", binDir)

	quitshExe := path.Join(binDir, "quitsh-integration-test")
	require.FileExists(t, quitshExe, "exe '%v' must exist.", quitshExe)

	covDir := os.Getenv("QUITSH_COVERAGE_DIR")
	require.DirExists(t, covDir, "QUITSH_COVERAGE_DIR=%s must exist.", covDir)

	return exec.NewCmdCtxBuilder().
		BaseCmd(quitshExe).
		Cwd(".").
		EnableCaptureError().
		Env("GOCOVERDIR=" + covDir).
		Build()
}

func TestCLIVersion(t *testing.T) {
	cli := setup(t)
	stdout, _, err := captureOutput(func() error {
		return cli.Check("--version")

	})

	require.NoError(t, err)
	assert.Contains(t, stdout, fmt.Sprintf("version %v", build.GetBuildVersion()))
}

func TestCLIList(t *testing.T) {
	cli := setup(t)
	_, stderr, err := captureOutput(func() error {
		return cli.Check("list", "-C", "../test") // test going back and forward again into test.

	})

	require.NoError(t, err)
	assert.Contains(t, stderr, "component-a")
}

func TestCLIExecTarget(t *testing.T) {
	cli := setup(t)
	_, stderr, err := captureOutput(func() error {
		return cli.Check(
			"exec-target",
			"--log-level",
			"debug",
			"component-a::build",
		)
	})

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}

func TestCLIExecTarget2(t *testing.T) {
	cli := setup(t)
	_, stderr, err := captureOutput(func() error {
		return cli.Check(
			"exec-target",
			"--log-level",
			"debug",
			"component-a::banana",
		)
	})

	require.NoError(t, err, "Stderr:\n"+stderr)

	assert.Contains(t, stderr, "Hello from integration test Go runner")
	assert.Contains(t, stderr, "🌻")
	assert.FileExists(t, path.Join(cli.Cwd(), "repo/component-a/.output/build/bin/cmd"))
}
