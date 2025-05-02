package exec

import (
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/lookpath"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandCtxNormal(t *testing.T) {
	dir := t.TempDir()
	ctx := NewCommandCtx(dir)

	pwd, err := ctx.Get("pwd")
	require.NoError(t, err)
	assert.Equal(t, dir, pwd)

	s, err := ctx.Get("ls", "this-is-not-existing")
	assert.Empty(t, s)
	require.ErrorContains(t, err, "this-is-not-existing")

	s, err = ctx.Get("ls", "-al", dir)
	assert.NotEmpty(t, s)
	require.NoError(t, err)
}

func TestCommandCtxBuilderAddArgs(t *testing.T) {
	ctx := NewCmdCtxBuilder().BaseCmd("ls").
		BaseArgs("-a").
		BaseArgs("--banana").
		Build()

	_, err := ctx.Get(".")
	require.ErrorContains(t, err, "--banana")

	ctx = NewCmdCtxBuilder().BaseCmd("ls").
		BaseArgs("-a").
		BaseArgs("-l").
		Build()

	_, err = ctx.Get(".")
	require.NoError(t, err)
}

func TestCommandCtxExitCode(t *testing.T) {
	dir := t.TempDir()

	ctx := NewCmdCtxBuilder().Cwd(dir).
		BaseCmd("ls").
		ExitCodeHandler(func(err *CmdError) error {
			switch {
			case err == nil:
				return nil
			case err.ExitCode() == 2:
				return errors.New("handled exit code 2")
			default:
				return err
			}
		}).
		Build()

	err := ctx.Check(dir)
	require.NoError(t, err)

	err = ctx.Check("lkajsdflkjasdf")
	require.ErrorContains(t, err, "handled exit code 2")

	// Overwrite the error handler.
	err = ctx.CheckWithEC(
		func(_ *CmdError) error { return errors.New("new error here") },
		"lkajsdflkjasdf",
	)
	require.ErrorContains(t, err, "new error here")
}

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

func TestCommandCtxPipeStdout(t *testing.T) {
	dir := t.TempDir()

	// Test output.
	output, _, err := captureOutput(func() error {
		ctx := NewCommandCtx(dir)

		return ctx.Check("pwd")
	})

	output = strings.TrimSpace(output)
	require.NoError(t, err)
	assert.Equal(t, dir, output)

	// Test no output.
	output, _, err = captureOutput(func() error {
		ctx := NewCmdCtxBuilder().Cwd(dir).Quiet().Build()

		return ctx.Check("pwd")
	})

	require.NoError(t, err)
	assert.Equal(t, "", output)

	// Test that there is output.
	output, _, err = captureOutput(func() error {
		ctx := NewCmdCtxBuilder().Cwd(dir).Quiet().NoQuiet().Build()

		return ctx.Check("pwd")
	})

	require.NoError(t, err)
	assert.Equal(t, dir, output)
}

func TestCommandCtxPipeStdErr(t *testing.T) {
	dir := t.TempDir()

	// Test output.  w
	_, output, err := captureOutput(func() error {
		ctx := NewCommandCtx(dir)

		return ctx.Check("ls", "this-is-not-existing")
	})

	output = strings.TrimSpace(output)
	require.Error(t, err)
	assert.Contains(t, output, "this-is-not-existing")

	// Test no output.
	_, output, err = captureOutput(func() error {
		ctx := NewCmdCtxBuilder().Cwd(dir).Quiet().Build()

		return ctx.Check("ls", "this-is-not-existing")
	})

	require.Error(t, err)
	assert.NotContains(t, output, "this-is-not-existing")

	// Test there is output.
	_, output, err = captureOutput(func() error {
		ctx := NewCmdCtxBuilder().Cwd(dir).Build()

		return ctx.Check("ls", "this-is-not-existing")
	})

	require.Error(t, err)
	assert.Contains(t, output, "this-is-not-existing")
}

func TestCommandCtxEnv(t *testing.T) {
	ctx := NewCmdCtxBuilder().Env("A=banana", "B=monkey", "C=monkey").Build()
	out, err := ctx.Get("env")
	require.NoError(t, err)
	assert.Contains(t, out, "A=banana")
	assert.Contains(t, out, "B=monkey")
	assert.Contains(t, out, "C=monkey")
	// We need the default environment, it must be set.
	assert.Contains(t, out, "PATH")

	ctx = NewCmdCtxBuilder().Env("A=banana", "B=monkey", "C=monkey").EnvRemove("A", "B").Build()
	out, err = ctx.Get("env")
	require.NoError(t, err)
	assert.NotContains(t, out, "A=banana")
	assert.NotContains(t, out, "B=banana")
	assert.Contains(t, out, "C=monkey")
	// We need the default environment, it must be set.
	assert.Contains(t, out, "PATH")
}

func TestCommandCtxEnvPure(t *testing.T) {
	ctx := NewCmdCtxBuilder().EnvEmpty().Env("A=banana", "B=monkey").Build()
	out, err := ctx.Get("env")
	require.NoError(t, err)
	assert.Contains(t, out, "A=banana")
	assert.Contains(t, out, "B=monkey")
	// We need the default environment, it must be set.
	assert.NotContains(t, out, "PATH")
}

func TestCommandCtxLookUp(t *testing.T) {
	// Copy full resolved file to new location.
	dir := t.TempDir()
	shNew := path.Join(dir, "sh")
	sh, err := exec.LookPath("sh")
	require.NoError(t, err)
	sh, err = filepath.EvalSymlinks(sh)
	require.NoError(t, err)

	t.Logf("sh: %v", sh)
	t.Logf("sh: %v", shNew)
	err = fs.CopyFileOrDir(sh, shNew, false)
	require.NoError(t, err)
	assert.True(t, fs.Exists(shNew))

	// Check lookup in `dir`
	shLP, err := lookpath.Look("sh", dir)
	require.NoError(t, err)
	assert.Equal(t, shNew, shLP)
}

func TestCommandCtxLookUpSymlink(t *testing.T) {
	// Copy full symlink to new location.
	dir := t.TempDir()
	shNew := path.Join(dir, "sh")
	sh, err := exec.LookPath("sh")
	require.NoError(t, err)

	err = os.Symlink(sh, shNew)
	require.NoError(t, err)
	assert.True(t, fs.Exists(shNew))

	// Check lookup in `dir`
	shLP, err := lookpath.Look("sh", dir)
	require.NoError(t, err)
	assert.Equal(t, shNew, shLP)

	// Compare to ctx.
	ctx := NewCmdCtxBuilder().
		Paths(dir).
		Build()
	assert.NotEmpty(t, ctx.env.Find("PATH"))
	assert.True(t, ctx.enableLookPath)

	baseCmd, args, err := ctx.getCommand([]string{"sh"})

	require.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, shNew, baseCmd)
}

func TestCommandGetSplit(t *testing.T) {
	dir := t.TempDir()
	ctx := NewCmdCtxBuilder().
		Cwd(dir).
		Build()

	out, e := ctx.GetSplit("true")
	require.NoError(t, e)
	assert.Empty(t, out)

	out, e = ctx.GetSplit("sh", "-c", "echo aa; echo b")
	require.NoError(t, e)
	assert.Len(t, out, 2)
	assert.Equal(t, "aa", out[0])
	assert.Equal(t, "b", out[1])
}
