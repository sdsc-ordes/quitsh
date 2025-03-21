package input

import (
	"path"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/debug"
)

type ID string
type BaseDir string

type Config struct {
	ID ID `yaml:"-"`

	// Regex patterns:
	// Prefixed with `!` denotes a exclude pattern.
	// Escaping for a first `!` can be done with `\!`.
	Patterns []string `yaml:"patterns" validate:"required"`

	// If the regex match (`BaseDir`) against the root directory
	// or the components root dir.
	RelativeToRoot bool   `yaml:"relToRoot"`
	BaseDir        string `yaml:"-"`

	includePatterns []string `yaml:"-"`
	excludePatterns []string `yaml:"-"`
}

// Init initializes this config.
func (in *Config) Init(id ID) {
	in.ID = id
	in.SplitIntoIncludeAndExcludes()
}

// DefineID defines the input id, based on the component name and the input name.
func DefineID(componentName string, inputName string) ID {
	return ID(componentName + "::" + inputName)
}

func DefineIDComp(componentName string) ID {
	return ID(componentName)
}

// IsComponent tells if this input id refers to the whole component
// instead of to an input set on the component.
// TODO: Wrap into better types (also 'comp://subdir' would be possible, maybe not needed).
func (i ID) IsComponent() bool {
	return !strings.Contains(string(i), "::")
}

// SplitIntoIncludeAndExcludes splits the patterns into
// include and exclude patterns.
func (in *Config) SplitIntoIncludeAndExcludes() {
	in.includePatterns = make([]string, 0, len(in.Patterns))
	in.excludePatterns = make([]string, 0, len(in.Patterns))

	for i := range in.Patterns {
		l := &in.includePatterns

		startIdx := 0
		if strings.HasPrefix(in.Patterns[i], "!") {
			l = &in.excludePatterns
			startIdx = 1
		}

		// Escaping with `\!`, split it off.
		if strings.HasPrefix(in.Patterns[i], "\\!") {
			startIdx = 1
		}

		(*l) = append((*l), in.Patterns[i][startIdx:])
	}
}

// Exclude returns exclude patterns.
func (in *Config) Excludes() []string {
	return in.excludePatterns
}

// Includes returns include patterns.
func (in *Config) Includes() []string {
	return in.includePatterns
}

// TrimOfBaseDir returns if the `BaseDir` could be trimmed of (its match).
func (in *Config) TrimOfBaseDir(absPath string) (relativePath string, trimmed bool) {
	return BaseDir(in.BaseDir).TrimOffFrom(absPath)
}

// TrimOffFrom returns if the base directory can be trimmed of from `absPath`.
func (in BaseDir) TrimOffFrom(absPath string) (string, bool) {
	debug.Assert(
		path.IsAbs(absPath),
		"path must be absolute '%s', if base path is absolute",
		absPath,
	)

	trim := path.Clean(string(in)) + "/"
	absPath = path.Clean(absPath)

	if !strings.HasPrefix(absPath, trim) {
		return absPath, false
	}

	return strings.TrimPrefix(absPath, trim), true
}
