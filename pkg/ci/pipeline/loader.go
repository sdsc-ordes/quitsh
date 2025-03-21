package pipeline

// PipelineSettingsLoader abstracts the different vendors to load
// `PipelineSettings`.
type PipelineSettingsLoader interface {
	// LoadFromEnv loads from the environment
	// (if `env` empty `os.Environ()` is used.)
	LoadFromEnv(env []string) (PipelineSettings, error)
}
