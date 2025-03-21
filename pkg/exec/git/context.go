package git

import "github.com/sdsc-ordes/quitsh/pkg/exec"

type Context struct {
	*exec.CmdContext
}

// NewCtx returns a new Git command context in the folder `cwd`.
// Mixin are modifier functions to adapt the ctx to custom needs.
func NewCtx(cwd string, mixin ...func(builder *exec.CmdContextBuilder)) Context {
	b := exec.NewCmdCtxBuilder().
		Cwd(cwd).
		BaseCmd("git")

	for _, mix := range mixin {
		mix(&b)
	}

	if len(mixin) == 0 {
		b.Quiet()
	}

	return Context{b.Build()}
}

// NewCtxAtRoot returns a new Git command context at the root of the Git repo starting from `cwd`.
// Returns the `rootDir` as well.
func NewCtxAtRoot(
	cwd string,
	mixin ...func(builder *exec.CmdContextBuilder),
) (Context, string, error) {
	g := NewCtx(cwd)
	rootDir, err := g.RootDir()
	if err != nil {
		return Context{}, "", err
	}

	return NewCtx(rootDir, mixin...), rootDir, nil
}
