package config

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var _ viper.Viper
var _ mapstructure.DecodeHookFunc
var _ mapstructure.Decoder

// UnmarshalMapstruct is an interface which types can implement to
// unmarshal themself when the `mapstructure` library
// is used (as in `[viper.Viper.Unmarshal]`).
type UnmarshalerMapstruct interface {
	UnmarshalMapstruct(data any) error
}

// UnmarshalMapstructDecodeHook is the hook for the `Unmarshal` function
// [mapstructure.Decoder].
func UnmarshalMapstructDecodeHook(_from, to reflect.Type, data any) (any, error) {
	result := reflect.New(to).Interface()

	unmarshaller, ok := result.(UnmarshalerMapstruct)
	if !ok {
		// Not of this interface.
		return data, nil
	}

	if err := unmarshaller.UnmarshalMapstruct(data); err != nil {
		return nil, err
	}

	return result, nil
}
