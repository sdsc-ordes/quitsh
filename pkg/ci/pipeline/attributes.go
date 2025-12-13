package pipeline

import (
	"strings"

	strs "github.com/sdsc-ordes/quitsh/pkg/common/strings"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/goccy/go-yaml"
)

type PipelineAttributes interface {
	Clone() PipelineAttributes
	Merge(other PipelineAttributes)
}

// ParseCIAttributes parses the first block in each `ss`
// like:
//
// ```markdown
//
//	```yaml {ci}
//	   value: true
//	   list: [ 1, 3]
//	```
//
// ```
// into `PipelineAttributes` and merges all of them together into `settings`.
func ParseCIAttributes(settings PipelineAttributes, ss ...string) error {
	foundHeaders := false
	const startPrefix = "```yaml {ci}"

	for i := range ss {
		yamlBlock := findCodeBlock(ss[i], startPrefix)
		if yamlBlock == "" {
			continue
		}
		log.Debug("Found YAML block.", "block", yamlBlock)

		decoder := yaml.NewDecoder(strings.NewReader(yamlBlock), yaml.Strict())
		setts := settings.Clone()

		e := decoder.Decode(setts)
		if e != nil {
			return errors.AddContext(e,
				"ci attributes YAML not parsable:\n---\n%s\n---",
				yamlBlock,
			)
		}

		foundHeaders = true
		settings.Merge(setts)
	}

	if !foundHeaders {
		log.Info("Could not find any ci attributes YAML block.")
	}

	return nil
}

// findCodeBlock returns the YAML code block in string `s`.
// Only returns errors if a found YAML block starts but is not delimited.
func findCodeBlock(s string, startPrefix string) string {
	state := 0 // 0: before start, 1: inside, 2: after
	startIdx := 0
	endIdx := 0

	lines := strs.SplitLines(s)

lineloop:
	for i := range lines {
		switch state {
		case 0:
			if lines[i] == startPrefix {
				startIdx = i + 1
				state = 1
			}
		case 1:
			if lines[i] == "```" {
				endIdx = i
				state = 2
			}
		default:
			break lineloop
		}
	}

	if state != 2 { //nolint:mnd
		// nothing found, or endTag not detected.
		return ""
	}

	debug.Assert(startIdx < endIdx, "not a correct YAML detection")

	return strings.Join(lines[startIdx:endIdx], "\n")
}
