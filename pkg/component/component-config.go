package component

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/input"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	Name    string  `yaml:"name"    validate:"required"`
	Version Version `yaml:"version" validate:"required"`

	Language string `yaml:"language" validate:"required"`

	Inputs  map[string]*input.Config  `yaml:"inputs"`
	Targets map[string]*target.Config `yaml:"targets"`
}

// Init implements the `Initializable` interface.
func (c *Config) Init() (err error) {
	err = validator.New().Struct(c)

	// Init target.
	for targetName, t := range c.Targets {
		e := t.Init(target.DefineID(c.Name, targetName))
		if e != nil {
			err = errors.Combine(err, e)
		}
	}

	// Init input.
	for name, in := range c.Inputs {
		in.Init(input.DefineID(c.Name, name))
	}

	return
}

// TargetById finds the target by the respective name in the config.
func (c *Config) TargetByID(id target.ID) *target.Config {
	for _, t := range c.Targets {
		if t.ID == id {
			return t
		}
	}

	return nil
}

// TargetByName finds the target by the respective name in the config.
func (c *Config) TargetByName(name string) *target.Config {
	if t, exists := c.Targets[name]; exists {
		return t
	}

	return nil
}
