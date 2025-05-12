package git

import (
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultGitArgs = []string{ //nolint:gochecknoglobals // allowed in tests
	"-c",
	"user.name=test",
	"-c",
	"user.email=test@example.com",
	"-c",
	"init.defaultBranch=main",
}

func adjustGitCtx(b *exec.CmdContextBuilder) {
	b.BaseArgs(defaultGitArgs...).
		Env("GIT_CONFIG_GLOBAL=",
			"GIT_CONFIG_SYSTEM=")
}

func setupGitRepo(t *testing.T) (repoCtx Context) {
	repo := t.TempDir()
	repoCtx = NewCtx(repo, adjustGitCtx)
	e := repoCtx.Chain().
		Check("init").
		Check("commit", "--allow-empty", "-m", "init1").
		Check("commit", "--allow-empty", "-m", "init2").
		Check("commit", "--allow-empty", "-m", "init3").
		Error()
	require.NoError(t, e)

	return
}

func setupGitRepoWithServer(t *testing.T) (repoCtx Context, serverCtx Context) {
	server := t.TempDir()
	serverCtx = NewCtx(server, adjustGitCtx)
	e := serverCtx.Chain().
		Check("init", "--bare").
		Error()
	require.NoError(t, e)

	repo := t.TempDir()
	repoCtx = NewCtx(repo, adjustGitCtx)
	e = repoCtx.Chain().
		Check("clone", server, repo).
		Check("commit", "--allow-empty", "-m", "init1").
		Error()
	require.NoError(t, e)

	return
}

func TestGetChanges(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)
	d := gitx.Cwd()

	containsChanges := func() {
		hasCh, ec := gitx.HasChanges(d)
		require.NoError(t, ec)
		assert.True(t, hasCh)

		changes, ec := gitx.Changes(d, false)
		require.NoError(t, ec)
		assert.Len(t, changes, 2)
		assert.Contains(t, changes, "file.txt")
		assert.Contains(t, changes, "a/b/file.txt")

		changes, ec = gitx.Changes(path.Join(d, "a", "b"), true)
		require.NoError(t, ec)
		assert.Len(t, changes, 1)
		assert.Contains(t, changes, "a/b/file.txt")

		changes, ec = gitx.Changes(path.Join(d, "a", "b"), false)
		require.NoError(t, ec)
		assert.Len(t, changes, 1)
		assert.Contains(t, changes, "file.txt")
	}

	e := os.MkdirAll(path.Join(d, "a", "b"), fs.DefaultPermissionsDir)
	require.NoError(t, e)

	file := path.Join(d, "file.txt")
	f, e := os.Create(file)
	require.NoError(t, e)
	file2 := path.Join(d, "a/b/file.txt")
	f2, e := os.Create(file2)
	require.NoError(t, e)

	e = gitx.Chain().
		Check("add", ".").
		Check("commit", "-m", "init").
		Error()
	require.NoError(t, e)

	paths, e := gitx.Files("a", false)
	require.NoError(t, e)
	assert.Contains(t, paths, "b/file.txt")

	paths, e = gitx.Files("a", true)
	require.NoError(t, e)
	assert.Contains(t, paths, "a/b/file.txt")

	changes, e := gitx.Changes(d, false)
	require.NoError(t, e)
	assert.Empty(t, changes)

	_, e = io.WriteString(f, "asdf")
	require.NoError(t, e)
	f.Close()
	_, e = io.WriteString(f2, "asdf")
	require.NoError(t, e)
	f2.Close()

	containsChanges()
}

func TestIsIgnored(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)
	d := gitx.Cwd()

	file := path.Join(d, ".gitignore")
	f, e := os.Create(file)
	require.NoError(t, e)
	_, e = io.WriteString(f, "file.txt")
	require.NoError(t, e)
	f.Close()
	require.NoError(t,
		gitx.Chain().
			Check("add", ".").
			Check("commit", "-m", "init").
			Error())

	_, e = os.Create(path.Join(d, "file.txt"))
	require.NoError(t, e)
	_, e = os.Create(path.Join(d, "file1.txt"))
	require.NoError(t, e)

	ignored, e := gitx.IsIgnored("file.txt")
	require.NoError(t, e)
	assert.True(t, ignored)

	ignored, e = gitx.IsIgnored("file1.txt")
	require.NoError(t, e)
	assert.False(t, ignored)
}

func TestGetTags(t *testing.T) {
	t.Parallel()
	repoCtx, _ := setupGitRepoWithServer(t)

	e := repoCtx.Chain().
		Check("commit", "--allow-empty", "-m", "1").
		Check("tag", "v1.0.0-beta+build").
		Check("commit", "--allow-empty", "-m", "2").
		Check("tag", "v1.0.0").
		Check("commit", "--allow-empty", "-m", "3").
		Check("tag", "v1.2.0").
		Check("push", "--tags").
		Check("ls-remote", "origin").
		Error()
	require.NoError(t, e)

	versions, e := repoCtx.VersionTags("origin")
	require.NoError(t, e)
	require.Len(t, versions, 3)
	assert.Equal(t, "1.2.0", versions[0].Version.String())
	assert.Equal(t, "refs/tags/v1.2.0", versions[0].Ref)

	assert.Equal(t, "1.0.0", versions[1].Version.String())
	assert.Equal(t, "refs/tags/v1.0.0", versions[1].Ref)

	assert.Equal(t, "1.0.0-beta+build", versions[2].Version.String())
	assert.Equal(t, "refs/tags/v1.0.0-beta+build", versions[2].Ref)
}

func TestGetChangesRevs(t *testing.T) {
	t.Parallel()
	repoCtx := setupGitRepo(t)
	d := repoCtx.Cwd()

	e := os.WriteFile(path.Join(d, "test"), []byte("bla"), fs.DefaultPermissionsFile)
	require.NoError(t, e)

	e = repoCtx.Chain().
		Check("add", ".").
		Check("commit", "--allow-empty", "-m", "1").
		Error()
	require.NoError(t, e)

	changes, e := repoCtx.ChangesBetweenRevs(".", "HEAD~1", "HEAD", false)
	require.NoError(t, e)
	assert.Len(t, changes, 1)
	assert.Contains(t, changes, "test")
}

func TestCommitIsAMerge(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)

	e := gitx.Chain().
		Check("checkout", "-b", "test").
		Check("commit", "--allow-empty", "-m", "1").
		Error()
	require.NoError(t, e)

	res, e := gitx.CommitIsAMerge("HEAD")
	require.NoError(t, e)
	assert.False(t, res)

	e = gitx.Chain().
		Check("checkout", "main").
		Check("commit", "--allow-empty", "-m", "main").
		Check("merge", "--no-ff", "test").
		Error()
	require.NoError(t, e)

	res, e = gitx.CommitIsAMerge("HEAD")
	require.NoError(t, e)
	assert.True(t, res)
}

func TestCommitContainsTag(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)
	res, e := gitx.CommitIsAMerge("HEAD")
	require.NoError(t, e)
	assert.False(t, res)

	e = gitx.Chain().
		Check("tag", "v1.2.3", "-m", "asf").
		Check("tag", "v1.2.4", "-m", "asf").
		Error()
	require.NoError(t, e)

	res, e = gitx.CommitContainsTag("HEAD", "v1.2.5")
	require.NoError(t, e)
	assert.False(t, res)

	res, e = gitx.CommitContainsTag("HEAD", "v1.2.3")
	require.NoError(t, e)
	assert.True(t, res)
	res, e = gitx.CommitContainsTag("HEAD", "v1.2.4")
	require.NoError(t, e)
	assert.True(t, res)
}

func TestIsRefAHeadOf(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)

	ahead, count, err := gitx.IsRefAheadOf("HEAD~1", "HEAD")
	require.NoError(t, err)
	assert.True(t, ahead)
	assert.Equal(t, 1, count)

	ahead, count, err = gitx.IsRefAheadOf("HEAD~2", "HEAD")
	require.NoError(t, err)
	assert.True(t, ahead)
	assert.Equal(t, 2, count)

	ahead, count, err = gitx.IsRefAheadOf("HEAD", "HEAD~2")
	require.NoError(t, err)
	assert.False(t, ahead)
	assert.Equal(t, 0, count)
}

func TestIsRefReachable(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)

	reachable, err := gitx.IsRefReachable("HEAD", "HEAD~1")
	require.NoError(t, err)
	assert.True(t, reachable)

	reachable, err = gitx.IsRefReachable("HEAD~2", "HEAD")
	require.NoError(t, err)
	assert.False(t, reachable)
}

func TestFetchUntil(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)

	lastCommit, err := gitx.CurrentRev()
	require.NoError(t, err)

	for i := range 10 {
		e := gitx.Check("commit", "--allow-empty", "-m", fmt.Sprintf("test dummy %v", i))
		require.NoError(t, e)
	}

	found, err := gitx.IsRefReachable("HEAD", lastCommit)
	require.NoError(t, err)
	assert.True(t, found)

	d := t.TempDir()
	err = gitx.Check("clone", "--depth", "1", "file://"+gitx.Cwd(), d)
	require.NoError(t, err)
	gitx = NewCtx(d, adjustGitCtx)

	found, err = gitx.IsRefReachable("HEAD", lastCommit)
	require.NoError(t, err)
	assert.False(t, found)

	found, err = gitx.FetchUntilCommit(
		"origin",
		lastCommit,
		FetchUntilOpts{Factor: 1, StartDepth: 2, MaxDepth: 10},
	)
	require.NoError(t, err)
	assert.True(t, found)

	found, err = gitx.IsRefReachable("HEAD", lastCommit)
	require.NoError(t, err)
	assert.True(t, found)
}

func TestLocalRefExists(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)

	shaExpect, e := gitx.CurrentRev()
	require.NoError(t, e)
	sha, e := gitx.LocalRefExists("main")
	require.NoError(t, e)
	assert.Equal(t, shaExpect, sha)
}

func TestCurrentRef(t *testing.T) {
	t.Parallel()
	gitx := setupGitRepo(t)
	ref, e := gitx.CurrentRef()
	require.NoError(t, e)
	assert.Equal(t, "main", ref)
}
