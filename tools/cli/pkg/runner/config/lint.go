package config

type LintSettings struct {
	// Additional arguments forwarded to the test tool.
	Args []string `yaml:"args"`
}

// NewLintSettings constructs a new build setting.
func NewLintSettings(
	args []string,
) LintSettings {
	return LintSettings{
		Args: args,
	}
}
