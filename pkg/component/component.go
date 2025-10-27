package component

import (
	"path"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"

	"github.com/hashicorp/go-version"
)

// Component represents a component in the mono-repo.
type Component struct {
	config *Config

	root       string
	configFile string
	outDir     string
}

func NewComponent(config *Config, root string, configFile string, outBaseDir string) Component {
	root = fs.MakeAbsolute(root)

	// Set the out directory.
	var outDir string
	if path.IsAbs(outBaseDir) {
		outDir = path.Join(outBaseDir, fs.OutputDir, config.Name)
	} else {
		outDir = path.Join(root, fs.OutputDir)
	}

	return Component{root: root, configFile: configFile, config: config, outDir: outDir}
}

type ComponentCreator = func(config *Config, rootDir string, configFile string) (*Component, error)
type ConfigAdjuster = func(config *Config) error

// NewComponentCreator creates a factory method which creates components.
// It will transform the config if a `transformConfig` function is given.
func NewComponentCreator(outBaseDir string, transformConfig ConfigAdjuster) ComponentCreator {
	return func(c *Config, root string, configFile string) (*Component, error) {
		if transformConfig != nil {
			err := transformConfig(c)
			if err != nil {
				return nil, err
			}
		}

		comp := NewComponent(c, root, configFile, outBaseDir)

		return &comp, nil
	}
}

// Name returns the name of the component.
func (c *Component) Name() string {
	return c.config.Name
}

// Language returns the language of the component.
func (c *Component) Language() string {
	return c.config.Language
}

// Version returns the language of the component.
func (c *Component) Version() *version.Version {
	return &c.config.Version.Version
}

// Config returns the config of the component.
func (c *Component) Config() *Config {
	return c.config
}

// ConfigFile returns the config file path.
func (c *Component) ConfigFile() string {
	return c.configFile
}

// Root returns the root directory of the component.
func (c *Component) Root() string {
	return c.root
}

// String returns a string representation of the component.
func (c *Component) String() string {
	return c.Name()
}
