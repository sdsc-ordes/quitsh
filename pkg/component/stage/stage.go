package stage

import (
	"fmt"
	"slices"
)

type Stage string

// Interface implementation guard.
var _ fmt.Stringer = (*Stage)(nil)

// String implements [fmt.Stringer].
func (s Stage) String() string {
	return (string)(s)
}

type StagePrio struct {
	// The stage.
	Stage Stage

	// The stage priority, lower priority come before other stage.
	Priority int
}

// IsAfter returns true if the stage is right after `other`.
// Meaning the priority is +1.
func (s StagePrio) IsAfter(other StagePrio) bool {
	return s.Priority == other.Priority+1
}

// IsBefore returns true if the stage is right before `other`.
// Meaning the priority is +1.
func (s StagePrio) IsBefore(other StagePrio) bool {
	return s.Priority+1 == other.Priority
}

type Stages []StagePrio

type TargetNameToStageMapper func(targetName string) (Stage, error)

// NewDefaults default stages.
func NewDefaults() Stages {
	return Stages{
		{Lint, 0},
		{Build, 1},
		{Test, 2},
		{Image, 3},
		{Deploy, 4},
		{Aux, 100},
	}
}

// Len implements the sort.Sort interface.
func (s Stages) Len() int {
	return len(s)
}

// Less implements the sort.Sort interface.
func (s Stages) Less(i, j int) bool {
	return s[i].Priority < s[j].Priority
}

// Swap implements the sort.Sort interface.
func (s Stages) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns `-1` if `s` has lower priority than `other`, and `0` if equal and `1` if greater.
func (s *StagePrio) Less(other *StagePrio) int {
	switch {
	case s.Priority < other.Priority:
		return -1
	case s.Priority > other.Priority:
		return 1
	default:
		return 0
	}
}

// Contains returns if the stage `stage` is contained.
func (s Stages) Contains(stage Stage) bool {
	return slices.IndexFunc(s, func(st StagePrio) bool { return st.Stage == stage }) >= 0
}

// Find returns the stage with name `stage`.
func (s Stages) Find(stage Stage) (st StagePrio, exists bool) {
	idx := slices.IndexFunc(s, func(st StagePrio) bool { return st.Stage == stage })

	exists = idx >= 0
	if exists {
		st = s[idx]
	}

	return
}
