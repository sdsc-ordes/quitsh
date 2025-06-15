package config

import (
	"fmt"
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CustomType struct {
	private string // This field is not seen by mapstructure because its private.
}

func (s *CustomType) UnmarshalMapstruct(data any) error {
	data, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map[string]interface{} but got %T", data)
	}

	type customTypeHelper struct {
		Private string `mapstructure:"bla"`
	}

	var h customTypeHelper
	if err := mapstructure.Decode(data, &h); err != nil {
		return err
	}

	*s = CustomType{private: h.Private}

	return nil
}

func TestMapstructureDecodeHook(t *testing.T) {
	v := map[string]any{
		"bla": "password",
	}

	var res CustomType

	// Unmarshall the custom type with the hook.
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(UnmarshalMapstructDecodeHook),
		Result:     &res,
	}

	decoder, err := mapstructure.NewDecoder(config)
	require.NoError(t, err)

	err = decoder.Decode(v)
	require.NoError(t, err)

	assert.Equal(t, "password", res.private)
}
