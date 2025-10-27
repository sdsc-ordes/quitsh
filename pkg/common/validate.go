package common

import val "github.com/go-playground/validator/v10"

var validator *val.Validate //nolint:gochecknoglobals // its ok to use as singleton

// Validator returns the validator singleton.
func Validator() *val.Validate {
	if validator == nil {
		validator = val.New()
	}

	return validator
}
