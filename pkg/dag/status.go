package dag

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

const ExecStatusSuccess = 0
const ExecStatusFailed = 1
const ExecStatusNotRun = 2

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

	RunnerStatuses []RunnerStatus

	Summary struct {
		lock     sync.Mutex
		statuses RunnerStatuses

		allErrors error
	}
)

func (s *Summary) AddStatus(r RunnerStatus) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.statuses = append(s.statuses, r)
	if r.Error != nil {
		s.allErrors = errors.Combine(s.allErrors, r.Error)
	}
}

func (s RunnerStatuses) Log() {
	var sb strings.Builder
	sb.WriteString("Summary:\n")

	const failedS = "❌"
	const successS = "🌻"
	const notRun = "🚫"
	var statusS string

	slices.SortFunc(s, func(a, b RunnerStatus) int {
		return cmp.Compare(a.TargetID, b.TargetID)
	})

	for i := range s {
		var s = &s[i]

		switch s.Status {
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
			s.CompName,
			s.TargetID,
			s.StepIdx,
			s.RunnerID,
		)
	}

	log.Info(sb.String())
}
