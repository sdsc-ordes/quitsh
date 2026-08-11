package config

type LintSettings struct {
	// Try to fix linting errors.
	Fix bool `yaml:"fix"`
}

// NewLintSettings constructs a new build setting.
func NewLintSettings() LintSettings {
	return LintSettings{}
}
