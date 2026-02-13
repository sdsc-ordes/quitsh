package dag

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

const ExecStatusNotRun = 0
const ExecStatusFailed = 1
const ExecStatusSuccess = 2

type (
	ExecStatus int

	RunnerStatus struct {
		Status ExecStatus
		Error  error

		CompName string
		TargetID target.ID
		StepIdx  step.Index
		RunnerID runner.RegisterID
	}

	RunnerStatuses []*RunnerStatus

	Summary struct {
		statuses  RunnerStatuses
		allErrors error
	}
)

// AddStatus adds all runner statuses to the summary.
func (s *Summary) AddStatus(r ...*RunnerStatus) {
	for _, stat := range r {
		s.statuses = append(s.statuses, stat)
		s.allErrors = errors.Combine(s.allErrors, stat.Error)
	}
}

// Log prints the log of the summary.
func (s RunnerStatuses) Log() {
	var sb strings.Builder
	sb.WriteString("Summary:\n")

	const failedS = "❌"
	const successS = "🌻"
	const notRun = "🚫"
	var statusS string

	slices.SortFunc(s, func(a, b *RunnerStatus) int {
		return cmp.Compare(a.TargetID, b.TargetID)
	})

	for _, stat := range s {

		switch stat.Status {
		case ExecStatusSuccess:
			statusS = successS
		case ExecStatusNotRun:
			statusS = notRun
		case ExecStatusFailed:
			statusS = failedS
		}

		fmt.Fprintf(
			&sb,
			"- %v: Component '%v', target id: '%v', step idx: '%v', runner id: '%v'\n",
			statusS,
			stat.CompName,
			stat.TargetID,
			stat.StepIdx,
			stat.RunnerID,
		)
	}

	log.Info(sb.String())
}
