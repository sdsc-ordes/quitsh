package pipeline

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

type PipelineType int

const (
	BranchPipeline       PipelineType = 0
	MergeRequestPipeline PipelineType = 1
	TagPipeline          PipelineType = 2

	BranchPipelineName       = "branch"
	MergeRequestPipelineName = "merge-request"
	TagPipelineName          = "tag"
)

func NewPipelineType(s string) (PipelineType, error) {
	switch s {
	case BranchPipelineName:
		return BranchPipeline, nil
	case MergeRequestPipelineName:
		return MergeRequestPipeline, nil
	case TagPipelineName:
		return TagPipeline, nil
	}

	return 0, fmt.Errorf("wrong pipeline type '%s'", s)
}

// GetAllPipelineTypes returns all possible image types.
func GetAllPipelineTypes() []PipelineType {
	return []PipelineType{BranchPipeline, MergeRequestPipeline, TagPipeline}
}

// Interface implementation guard.
var _ fmt.Stringer = (*PipelineType)(nil)

// String implements [fmt.Stringer] .
func (v PipelineType) String() string {
	switch v {
	case BranchPipeline:
		return BranchPipelineName
	case MergeRequestPipeline:
		return MergeRequestPipelineName
	case TagPipeline:
		return TagPipelineName
	}

	panic("Not implemented.")
}

// Interface implementation guard.
var _ yaml.InterfaceUnmarshaler = (*PipelineType)(nil)

// UnmarshalYAML implements [yaml.InterfaceUnmarshaler].
func (v *PipelineType) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string
	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewPipelineType(s)

	return
}

// Interface implementation guard.
var _ yaml.InterfaceMarshaler = (*PipelineType)(nil)

// MarshalYAML implements [yaml.InterfaceMarshaler].
func (v PipelineType) MarshalYAML() (any, error) {
	return v.String(), nil
}
