package git

import (
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/hashicorp/go-version"
)

// RootDir returns the top level directory inside `cwd`.
// If in a submodule it returns the submodule root directory.
func (gitx *Context) RootDir() (string, error) {
	s, err := gitx.Get("rev-parse", "--show-toplevel")
	if err != nil {
		return "", errors.AddContext(err, "cannot determine Git repo root directory")
	}

	return s, nil
}

// IsIgnored returns true if `path` is ignored by Git.
func (gitx *Context) IsIgnored(path ...string) (ignored bool, err error) {
	cmd := append([]string{"check-ignore"}, path...)
	err = gitx.CheckWithEC(exec.ExitCode0And1SuccessVar(&ignored), cmd...)

	return
}

// Changes returns all changed paths in `dir`
// (relative to working dir if its a relative path).
// Staged files are not returned!
func (gitx *Context) Changes(dir string, absPath bool) ([]string, error) {
	relativeArg := "--relative"
	if absPath {
		relativeArg = "--no-relative"
	}

	if !path.IsAbs(dir) {
		dir = path.Join(gitx.Cwd(), dir)
	}

	files, err := gitx.GetSplitWithEC(
		exec.ExitCode0And1Success(),
		"-C", dir,
		"diff", relativeArg, "--name-only", "--exit-code", ".")

	if err != nil {
		return nil, errors.AddContext(
			err,
			"could not get changes in dir '%s'",
			dir,
		)
	}

	return files, nil
}

// HasChanges reports if there any unstaged changes in repository.
func (gitx *Context) HasChanges(dir string) (hasChanges bool, err error) {
	noChanges := false
	err = gitx.CheckWithEC(exec.ExitCode0And1SuccessVar(&noChanges),
		"-C", dir,
		"diff", "--exit-code", ".")

	return !noChanges, err
}

// ChangesBetweenRevs computs path changes between two revision specs,
// as `git diff revA revB`.
// You can specify `HEAD~1` and `HEAD` for example.
// To get absolute paths (relative to the repository) use `noRelative`.
// To convert all to absolute paths use `fs.MakeAllAbsolute`.
func (gitx *Context) ChangesBetweenRevs(
	dir string,
	revA, revB string,
	noRelative bool,
) ([]string, error) {
	relativeArg := "--relative"
	if noRelative {
		relativeArg = "--no-relative"
	}

	if !path.IsAbs(dir) {
		dir = path.Join(gitx.Cwd(), dir)
	}

	files, err := gitx.GetSplit(
		"-C", dir,
		"diff", relativeArg, "--name-only", revA, revB)

	if err != nil {
		return nil, errors.AddContext(
			err,
			"could not get changes between '%s' and '%s' in dir '%s'",
			revA,
			revB,
			dir,
		)
	}

	return files, nil
}

// Files returns all files managed by Git in dir `dir`.
func (gitx *Context) Files(dir string, noRelative bool) ([]string, error) {
	if !path.IsAbs(dir) {
		dir = path.Join(gitx.Cwd(), dir)
	}

	var paths []string
	var err error

	if noRelative {
		paths, err = gitx.GetSplit("-C", dir, "ls-files", "--full-name")
	} else {
		paths, err = gitx.GetSplit("-C", dir, "ls-files")
	}

	return paths, err
}

// Check if the current commit is a merge commit.
func (gitx *Context) CommitIsAMerge(ref string) (bool, error) {
	parentsCount, err := gitx.Get("rev-list", "--no-walk", "--count", ref+"^@")
	if err != nil {
		return false, err
	}

	c, err := strconv.Atoi(parentsCount)
	if err != nil {
		return false, err
	}

	return c >= 2, nil //nolint:mnd
}

// CommitContainsTag returns `true` if the commit (ref) `ref`
// contains the tag `tag`.
func (gitx *Context) CommitContainsTag(ref string, tag string) (bool, error) {
	tags, err := gitx.GetSplit("tag", "--list", "--points-at", ref)
	if err != nil {
		return false, err
	}

	return slices.Contains(tags, tag), nil
}

type VersionTag struct {
	Version *version.Version
	SHA1    string
	Ref     string
}

// VersionTags returns all versions from existing Git version tags (e.g. `v1.2.3`)
// in descendin order (sem. version).
//
//nolint:mnd
func (gitx *Context) VersionTags(remote string) ([]VersionTag, error) {
	tags, err := gitx.GetSplit("ls-remote", "--tags", "--refs", "--sort=-v:refname", remote)
	if err != nil {
		return nil, err
	}

	var versions []VersionTag

	for i := range tags {
		s := strings.SplitN(tags[i], "\t", 2)
		if len(s) != 2 { //nolint:mnd
			return nil, errors.New("could not split ref line '%s'", s)
		}

		sha := strings.TrimSpace(s[0])
		ref := strings.TrimSpace(s[1])

		versionTag := strings.TrimPrefix(ref, "refs/tags/")
		if versionTag == ref {
			continue
		}

		ver := GetVersionFromTag(versionTag)
		if ver != nil {
			versions = append(versions, VersionTag{ver, sha, ref})
		}
	}

	// Reverse sorting.
	slices.SortFunc(versions, func(a, b VersionTag) int {
		switch {
		case a.Version.LessThan(b.Version):
			return 1
		case a.Version.Equal(b.Version):
			return 0
		default:
			return -1
		}
	})

	return versions, nil
}

// Checks if a local branch `ref` exists and returns the SHA1.
//
//nolint:mnd
func (gitx *Context) LocalRefExists(ref string) (sha1 string, err error) {
	var s string
	s, err = gitx.Get("show-ref", "refs/heads/"+ref)
	if err != nil {
		return
	}

	shaAndRef := strings.SplitN(s, " ", 2)
	if len(shaAndRef) == 2 {
		sha1 = shaAndRef[0]
	} else {
		err = errors.New("could not split into ref '%s'", s)
	}

	return
}

// IsRefAheadOf checks if ref `refB` is ahead of `refA` and by how many commits.
// Returns `false` and `0` if `refB` is not ahead of `refA`.
func (gitx *Context) IsRefAheadOf(refA string, refB string) (bool, int, error) {
	shaA, err := gitx.Get("rev-parse", refA)
	if err != nil {
		return false, 0, err
	}
	shaB, err := gitx.Get("rev-parse", refB)
	if err != nil {
		return false, 0, err
	}

	c, err := gitx.Get("rev-list", "--count", "^"+shaA, shaB) // same as "refA..refB"
	if err != nil {
		return false, 0, err
	}

	count, err := strconv.Atoi(c)
	if err != nil {
		return false, 0, err
	}

	return count != 0, count, nil
}

// GetVersionFromTag returns the version if the tag is a release tag (e.g. v1.3.4).
func GetVersionFromTag(tag string) *version.Version {
	t := strings.TrimPrefix(tag, "v")
	if t == tag {
		return nil
	}

	v, err := version.NewSemver(tag)
	if err == nil {
		return v
	}

	return nil
}

// CurrentRef returns the current Git symbolic reference (branch).
// e.g. `feat/bla`.
func (gitx *Context) CurrentRef() (string, error) {
	ref, err := gitx.Get("symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}

	return ref, nil
}

// CurrentRev returns the current revision SHA1.
func (gitx *Context) CurrentRev() (string, error) {
	rev, err := gitx.Get("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}

	return rev, nil
}

// ObjectExists checks if an object in Git addressed by the SHA1 `sha1` exists.
func (gitx *Context) ObjectExists(sha1 string) (exists bool, err error) {
	err = gitx.CheckWithEC(
		exec.ExitCode0And1SuccessVar(&exists),
		"cat-file", "-e", sha1)

	return
}

// IsRefReachable reports if `ref` (can be branch/tag/commit) is contained starting
// from `fromRef`.
func (gitx *Context) IsRefReachable(fromRef string, ref string) (reachable bool, err error) {
	// Check if ref does even exist in the repo.
	exists, err := gitx.ObjectExists(ref)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	err = gitx.CheckWithEC(
		exec.ExitCode0And1SuccessVar(&reachable),
		"merge-base",
		"--is-ancestor",
		ref,
		fromRef,
	)

	return
}

type FetchUntilOpts struct {
	Factor     int
	StartDepth int
	MaxDepth   int
}

// FetchUntilCommit fetches exponentially increasing the
// shallow boundary of the remote (`*4`)
// until the `commitSHA` is found in `git rev-list `.
// Only tries until `maxDepth` has been reached.
// Meaningful values are: `factor = 4`, `startDepth = 10`, `maxDepth = 1000`.
func (gitx *Context) FetchUntilCommit(
	remote string,
	commitSHA string,
	opts FetchUntilOpts,
) (found bool, err error) {
	opts.Factor = max(1, opts.Factor)
	opts.StartDepth = max(1, opts.StartDepth)
	opts.MaxDepth = max(0, opts.MaxDepth)

	accum := 0
	for accum <= opts.MaxDepth {
		log.Info("Fetch on remote with more depth.", "remote", remote, "deepen", opts.StartDepth)

		err = gitx.Check("fetch", "origin", "--deepen="+strconv.Itoa(opts.StartDepth))

		if err != nil {
			return
		}

		found, err = gitx.IsRefReachable("HEAD", commitSHA)
		if err != nil {
			return
		}

		if found {
			return
		}

		accum += opts.StartDepth
		opts.StartDepth *= opts.Factor
	}

	return
}
