package pipeline

import (
	"slices"

	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// GetChangesOnPipeline reports all changed paths depending on the pipeline type:
// - on merge request and branch pipelines: the diff between the HEAD~1 and HEAD is used.
// - on tag pipelines:
//   - if the tag is a semantic Git version tag the diff between
//     the last semantic Git version tag and the current one is used.
//     (`v1.2.0` -> `v.1.3.0`)
func GetChangesOnPipeline(
	gitx git.Context,
	sett *PipelineSettings,
	remote string,
) ([]string, error) {
	const noRelative = true

	switch sett.Type {
	case MergeRequestPipeline:
		ok, err := gitx.CommitIsAMerge("HEAD")
		if err != nil {
			return nil, err
		}

		if ci.IsRunning() && !ok {
			return nil, errors.New("current commit is not a merge commit, " +
				"you need to merge the source branch into " +
				"the target branch.")
		}
		log.Info("Compute changes.", "old", "HEAD~1", "new", "HEAD")
		files, err := gitx.ChangesBetweenRevs(".", "HEAD~1", "HEAD", noRelative)

		return fs.MakeAllAbsoluteTo(gitx.Cwd(), files...), err
	case BranchPipeline:
		log.Info("Compute changes.", "old", "HEAD~1", "new", "HEAD")
		files, err := gitx.ChangesBetweenRevs(".", "HEAD~1", "HEAD", noRelative)

		return fs.MakeAllAbsoluteTo(gitx.Cwd(), files...), err
	case TagPipeline:
		version := git.GetVersionFromTag(sett.Git.Ref)
		if version == nil {
			// On normal tags we just report the changes between the
			// last commit.
			log.Info("Compute changes.", "old", "HEAD~1", "new", "HEAD")
			files, err := gitx.ChangesBetweenRevs(".", "HEAD~1", "HEAD", noRelative)

			return fs.MakeAllAbsoluteTo(gitx.Cwd(), files...), err
		}

		ok, err := gitx.CommitContainsTag("HEAD", sett.Git.Ref)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, errors.New(
				"current commit '%v' is does not contain a tag '%v'",
				sett.Git.CommitSHA,
				sett.Git.Ref,
			)
		}

		// Get the last version.
		allVersions, err := gitx.VersionTags(remote)
		if err != nil {
			return nil, err
		}

		// Get the next version
		idx := slices.IndexFunc(allVersions, func(v git.VersionTag) bool {
			return v.Version.Equal(version)
		})

		var files []string
		if idx+1 <= len(allVersions) {
			v := &allVersions[idx+1]

			// Fetch the tag.
			if e := gitx.Check("fetch", "origin", v.Ref+":"+v.Ref); e != nil {
				return nil, e
			}

			log.Info("Compute changes.", "old", v.Version, "new", "HEAD")
			files, err = gitx.ChangesBetweenRevs(
				".",
				v.Ref,
				"HEAD",
				noRelative)
		} else {
			log.Info("Did not find any older version. Taking all files.")
			files, err = gitx.Files(".", noRelative)
		}

		return fs.MakeAllAbsoluteTo(gitx.Cwd(), files...), err
	default:
		panic("not implemented")
	}
}
