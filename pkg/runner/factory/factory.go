package factory

import (
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

type IFactory interface {
	Register(
		id runner.RegisterID,
		entry ...runner.RunnerData,
	) error

	RegisterToKey(
		key runner.RegisterKey,
		id runner.RegisterID,
	) error

	CreateByKey(
		key runner.RegisterKey,
		toolchain string,
		rawConfig step.AuxConfigRaw,
	) (runners []RunnerInstance, err error)

	CreateByID(
		id runner.RegisterID,
		toolchain string,
		rawConfig step.AuxConfigRaw,
	) (runners []RunnerInstance, err error)
}

func NewFactory() IFactory {
	return &factory{
		byIDs:  make(runnerMapID),
		byKeys: make(runnerMapKeys),
	}
}

type runnerMapID = map[runner.RegisterID][]runner.RunnerData
type runnerMapKeys = map[runner.RegisterKey]runner.RegisterID

type factory struct {
	// Registered runners by ids.
	byIDs runnerMapID

	// Registered runners by stage and short name.
	byKeys runnerMapKeys
}

type RunnerInstance struct {
	RunnerID  runner.RegisterID
	Runner    runner.IRunner
	Toolchain string
}

// Create returns the runner instances created for
// `runnerName` for step `step`.
func (fac *factory) CreateByKey(
	key runner.RegisterKey,
	toolchain string,
	rawConfig step.AuxConfigRaw,
) ([]RunnerInstance, error) {
	id, exists := fac.byKeys[key]

	if !exists {
		return nil, errors.New(
			"could not find runner name '%v' for stage '%v', is the runner registered?",
			key.Name(),
			key.Stage(),
		)
	}

	return fac.CreateByID(id, toolchain, rawConfig)
}

func (fac *factory) CreateByID(
	id runner.RegisterID,
	toolchain string,
	rawConfig step.AuxConfigRaw,
) (runners []RunnerInstance, err error) {
	entries, exists := fac.byIDs[id]

	if !exists {
		return nil, errors.New(
			"could not find runner id '%v', is the runner registered?",
			id,
		)
	}

	for _, entry := range entries {
		if toolchain == "" {
			toolchain = entry.DefaultToolchain
		}

		var config any
		config, err = loadRunnerConfig(id, entry.RunnerConfigUnmarshal, rawConfig)
		if err != nil {
			return
		}

		var runner runner.IRunner

		runner, err = entry.Creator(config)
		if err != nil {
			err = errors.AddContext(err,
				"could not create runner with id '%v'", id)

			return
		}

		runners = append(runners, RunnerInstance{id, runner, toolchain})
	}

	return
}

func loadRunnerConfig(
	id runner.RegisterID,
	unmarshaller step.RunnerConfigUnmarshaller,
	rawConfig step.AuxConfigRaw,
) (config any, err error) {
	if unmarshaller == nil {
		return nil, nil //nolint:nilnil // Intended to return a nil config on no unmarshaller.
	}

	config, err = unmarshaller(rawConfig)

	if err != nil {
		return nil,
			errors.AddContext(
				err,
				"could not create runner config for runner id '%v'", id,
			)
	} else if config == nil {
		return nil, fmt.Errorf("you cannot return nil from config unmarshaller for runner id '%v'", id)
	}

	return config, nil
}

// Register adds a bunch of runners to the runner factory.
func (fac *factory) Register(
	id runner.RegisterID,
	entry ...runner.RunnerData,
) error {
	log.Debug("Registering runner '%v'.", id)

	_, exists := fac.byIDs[id]
	if exists {
		return errors.New("you cannot register another runner with id '%v'", id)
	}

	for _, e := range entry {
		if e.Creator == nil {
			return errors.New("the runner creator for runner id '%v' is nil", id)
		}
	}

	fac.byIDs[id] = entry

	return nil
}

// RegisterRunnerToKey registers runner id to the key (stage name and runner name).
func (fac *factory) RegisterToKey(key runner.RegisterKey, id runner.RegisterID) error {
	_, exists := fac.byKeys[key]
	if exists {
		return errors.New("you cannot register another same key '%v' for runner id '%v'", key, id)
	}

	fac.byKeys[key] = id

	return nil
}
