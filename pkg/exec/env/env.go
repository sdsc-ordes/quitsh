package env

import (
	"regexp"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type EnvValue struct {
	Value string
	// idx represent the index in the `EnvList`. (1-based)
	idx int
}

type EnvList []string

type EnvMap map[string]EnvValue

// Idx returns the index of the value in the original `EnvList`.
func (v *EnvValue) Idx() int {
	return v.idx - 1
}

// Determines if this env. value is found in the `EnvList`.
func (v *EnvValue) Defined() bool {
	return v.Idx() >= 0
}

// Get accesses the map safely, and panics when the key is not in the map.
// This is useful to accidentally mistype stuff.
func (v EnvMap) Get(key string) EnvValue {
	if v, exists := v[key]; exists {
		return v
	}

	log.Panic("Cannot access environment var. with key '%s': not in current map.", key)

	return EnvValue{}
}

// Defined returns true if all env variables are defined.
func (v *EnvMap) AssertDefined() error {
	for k, val := range *v {
		if !val.Defined() {
			return errors.New("credential with key '%s' is not defined or empty", k)
		}
	}

	return nil
}

// Defined returns true if all env variables are defined and not empty.
func (v *EnvMap) AssertNotEmpty() error {
	for k, val := range *v {
		if !val.Defined() || val.Value == "" {
			return errors.New("credential with key '%s' is not defined or empty", k)
		}
	}

	return nil
}

var envRe = regexp.MustCompile("[A-Z_][A-Z0-9_]+")

// AssertProperEnvKeys asserts that all env. keys are proper (and not potentially passwords).
func AssertProperEnvKey(keys ...string) error {
	for _, k := range keys {
		short := k[0 : (len(k)+1)/2]
		if !envRe.MatchString(k) {
			return errors.New("environment key '%s...' is not a conventional env. variable", short)
		}
	}

	return nil
}

// FindIdx finds the value of env. variable with key `key`.
func (l EnvList) FindIdx(key string) EnvValue {
	key += "="

	for i := range l {
		if strings.HasPrefix(l[i], key) {
			return EnvValue{strings.TrimPrefix(l[i], key), i + 1}
		}
	}

	return EnvValue{}
}

// FindAll environment variables given in `keys`.
// If `keys` is nil, all env. variables are returned.
func (l EnvList) FindAll(keys ...string) (res EnvMap) {
	if keys == nil {
		res = make(map[string]EnvValue, len(l))
	} else {
		res = make(map[string]EnvValue, len(keys))
		for i := range keys {
			res[keys[i]] = EnvValue{}
		}
	}

	for i := range l {
		parts := strings.SplitN(l[i], "=", 2) //nolint:mnd

		foundIdx := i + 1
		key := &parts[0]

		if keys != nil {
			if _, exists := res[*key]; !exists {
				continue
			}
		}

		if len(parts) == 2 { //nolint:mnd
			res[*key] = EnvValue{parts[1], foundIdx}
		} else {
			res[*key] = EnvValue{idx: foundIdx}
		}
	}

	return res
}

// NewEnvMap returns the env. map from the environment list.
func NewEnvMapFromList(env EnvList) EnvMap {
	return env.FindAll()
}

// Find finds the env. variable with `key`.
func (l EnvList) Find(key string) string {
	val := l.FindIdx(key)

	return val.Value
}
