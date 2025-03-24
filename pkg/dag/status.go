package dag

import (
	"fmt"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

type RunnerStatus struct {
	Failed   bool
	CompName string
	TargetID target.ID
	StepIdx  step.Index
	RunnerID runner.RegisterID
}

type RunnerStatuses []RunnerStatus

func (s RunnerStatuses) Log() {
	var sb strings.Builder
	sb.WriteString("Summary:\n")

	const failedS = "❌"
	const successS = "🌻"
	var failed string

	for i := range s {
		var s = &s[i]

		if s.Failed {
			failed = failedS
		} else {
			failed = successS
		}

		fmt.Fprintf(
			&sb,
			"- %v: Component '%v', target id: '%v', step idx: '%v', runner id: '%v'\n",
			failed,
			s.CompName,
			s.TargetID,
			s.StepIdx,
			s.RunnerID,
		)
	}

	log.Info(sb.String())
}
