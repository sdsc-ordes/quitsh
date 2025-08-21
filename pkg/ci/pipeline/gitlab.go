package pipeline

import (
	"os"

	strs "github.com/sdsc-ordes/quitsh/pkg/common/strings"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/env"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

const (
	ciCommitSHA                    = "CI_COMMIT_SHA"
	ciCommitRefName                = "CI_COMMIT_REF_NAME"
	ciCommitTag                    = "CI_COMMIT_TAG"
	ciCommitBranch                 = "CI_COMMIT_BRANCH"
	ciMergeRequestSourceBranchName = "CI_MERGE_REQUEST_SOURCE_BRANCH_NAME"
	ciMergeRequestTargetBranchName = "CI_MERGE_REQUEST_TARGET_BRANCH_NAME"
	ciCommitMessage                = "CI_COMMIT_MESSAGE"
	ciCommitTagMessage             = "CI_COMMIT_TAG_MESSAGE"
	ciMergeRequestDesc             = "CI_MERGE_REQUEST_DESCRIPTION"
	ciMergeRequestLabels           = "CI_MERGE_REQUEST_LABELS"
)

type gitlabSettingsLoader struct {
	attrs PipelineAttributes
}

func NewPipelineSettingsLoaderGitlab(attrs PipelineAttributes) PipelineSettingsLoader {
	return &gitlabSettingsLoader{attrs}
}

// LoadFromEnv loads the settings from the environment.
func (p *gitlabSettingsLoader) LoadFromEnv(
	e []string,
) (PipelineSettings, error) {
	log.Info("Loading pipeline settings for Gitlab pipeline from environment.")

	var env env.EnvList = e

	if env == nil {
		env = os.Environ()
	}

	envs := env.FindAll(
		ciCommitSHA,
		ciCommitRefName,
		ciCommitTag,
		ciCommitBranch,
		ciMergeRequestSourceBranchName,
		ciMergeRequestTargetBranchName,
		ciCommitMessage,
		ciCommitTagMessage,
		ciMergeRequestDesc,
		ciMergeRequestLabels,
	)

	var t PipelineType
	var ref, sourceRef, targetRef, targetCommitSHA string
	var labels []string
	commitSHA := envs.Get("CI_COMMIT_SHA").Value

	if v := envs.Get(ciCommitBranch); v.Defined() { //nolint:nestif
		t = BranchPipeline
		ref = v.Value
	} else if v = envs.Get(ciCommitTag); v.Defined() {
		t = TagPipeline
		ref = v.Value
	} else if v = envs.Get(ciMergeRequestSourceBranchName); v.Defined() {
		t = MergeRequestPipeline
		ref = v.Value
		sourceRef = v.Value
		targetRef = envs.Get(ciMergeRequestTargetBranchName).Value

		var err error
		targetCommitSHA, err = findRemoteHead(targetRef)
		if err != nil {
			return PipelineSettings{}, err
		}

		l := envs.Get(ciMergeRequestLabels)
		if l.Defined() && l.Value != "" {
			labels = strs.SplitAndTrim(l.Value, ",")
		}
	} else {
		return PipelineSettings{}, errors.New("could not determine pipeline type")
	}

	log.Info("Parsing CI attributes.")

	var pe error
	switch t {
	case BranchPipeline:
		pe = ParseCIAttributes(p.attrs,
			envs.Get("CI_COMMIT_MESSAGE").Value,
		)
	case MergeRequestPipeline:
		pe = ParseCIAttributes(p.attrs,
			envs.Get("CI_MERGE_REQUEST_DESCRIPTION").Value,
			envs.Get("CI_COMMIT_MESSAGE").Value,
		)
	case TagPipeline:
		pe = ParseCIAttributes(p.attrs,
			envs.Get("CI_COMMIT_TAG").Value,
			envs.Get("CI_COMMIT_MESSAGE").Value,
		)
	default:
		log.Panic("Parsing attributes on this pipeline type is not implemented.")
	}

	if pe != nil {
		return PipelineSettings{}, errors.Combine(
			pe,
			errors.New("could not parse ci attributes on pipeline '%v'",
				t,
			))
	}

	return PipelineSettings{
		Type: t,
		Git: PipelineGitSettings{
			Ref:             ref,
			CommitSHA:       commitSHA,
			SourceRef:       sourceRef,
			TargetRef:       targetRef,
			TargetCommitSHA: targetCommitSHA,
			Labels:          labels,
		},
	}, nil
}

func findRemoteHead(ref string) (string, error) {
	log.Info("Finding remote SHA.", "ref", ref)
	gitx := git.NewCtx("")

	return gitx.RemoteBranchExists(ref)
}
