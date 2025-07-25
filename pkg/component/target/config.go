package target

import (
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/component/input"
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

type Config struct {
	ID ID `yaml:"-"`

	Stage     stage.Stage     `yaml:"stage"`
	StagePrio stage.StagePrio `yaml:"-"`

	Steps []step.Config `yaml:"steps"`

	Inputs       []input.ID `yaml:"inputs,omitempty"`
	Dependencies []ID       `yaml:"depends,omitempty"`
}

// Init initializes this config.
func (c *Config) Init(id ID) (err error) {
	c.ID = id

	for i := range c.Steps {
		e := c.Steps[i].Init(step.Index(i))

		if e != nil {
			e = errors.AddContext(e, "could not initialize target config with id '%v'", c.ID)
			err = errors.Combine(err, e)
		}
	}

	return
}

func DefineID(componentName string, inputName string) ID {
	return ID(
		strings.ReplaceAll(componentName, NamespaceSeparator, "-") + "::" +
			strings.ReplaceAll(inputName, NamespaceSeparator, "-"))
}

type IConfig interface {
	ID() string
	Stage() stage.Stages
}
