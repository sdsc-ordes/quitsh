package pipeline

import (
	"bytes"
	"io"
	"os"
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type PipelineGitSettings struct {
	// The branch ref or tag or commit SHA.
	Ref string `yaml:"refName"`

	// The current commit which this pipeline runs on.
	CommitSHA string `yaml:"commitSha"`

	// If `Type == MergePipeline`:
	RefSource string   `yaml:"refSource"`
	RefTarget string   `yaml:"refTarget"`
	Labels    []string `yaml:"labels,omitempty"`
}

type PipelineSettings struct {
	Type PipelineType        `yaml:"type"`
	Git  PipelineGitSettings `yaml:"git"`
}

// NewPipelineSettingsFromReader loads the settings from a YAML reader.
func NewPipelineSettingsFromReader[T any, TP config.Initializable[T]](
	reader io.Reader,
) (T, error) {
	return config.LoadFromReader[T, TP](reader)
}

// LoadFromFile loads the settings from a YAML file.
// Attributes `attrs` can be `nil`, in which case they are not loaded
// from the YAML.
func NewPipelineSettingsFromFile[T any, TP config.Initializable[T]](
	file string,
) (T, error) {
	log.Info("Load pipeline settings from file.", "path", file)

	return config.LoadFromFile[T, TP](file)
}

// Write stores the settings to YAML.
func WritePipelineSettings[T any](settings *T, writer io.Writer) error {
	return config.SaveToWriter(settings, writer)
}

// Write stores the settings to a YAML file `file`.
func WritePipelineSettingsToFile[T any](settings *T, file string) error {
	log.Info("Write pipeline settings to file.", "path", file)

	buf := bytes.NewBuffer(nil)
	err := WritePipelineSettings(settings, buf)
	if err != nil {
		return err
	}

	if ci.IsRunning() {
		log.Info(
			"Pipeline settings file content to copy:",
			"content",
			"\n---\n"+buf.String()+"\n---",
		)
	}

	err = os.MkdirAll(path.Dir(file), fs.DefaultPermissionsFile)
	if err != nil {
		return err
	}

	return os.WriteFile(file, buf.Bytes(), fs.DefaultPermissionsFile)
}
