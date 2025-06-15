package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SpecialType struct {
	A int `yaml:"special"`
}

func (v *SpecialType) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'BuildType'")
	}

	if d == "banana" {
		v.A = 4242
	}

	return nil
}

// TestApplyConfigKeyValues tests the main function that applies a slice of key-value strings
// to a given config struct.
//
//nolint:funlen
func TestApplyKeyValues(t *testing.T) {
	// Define a target config struct for testing
	type NestedConfig struct {
		C int    `yaml:"c"`
		D bool   `yaml:"d"`
		E string `yaml:"e"`
	}

	type TestConfig struct {
		A string         `yaml:"a"`
		B NestedConfig   `yaml:"b"`
		F []string       `yaml:"f"`
		G map[string]int `yaml:"g"`
		S SpecialType    `yaml:"spez"`
	}

	// Define test cases
	testCases := []struct {
		name           string
		keyValues      []string
		initialConfig  TestConfig
		expectedConfig TestConfig
		wantErr        bool
	}{
		{
			name:           "success - set single top-level string field",
			keyValues:      []string{"a: hello"},
			initialConfig:  TestConfig{},
			expectedConfig: TestConfig{A: "hello"},
		},
		{
			name:           "success - set single nested int field",
			keyValues:      []string{"b.c: 123"},
			initialConfig:  TestConfig{},
			expectedConfig: TestConfig{B: NestedConfig{C: 123}},
		},
		{
			name:           "success - set multiple fields",
			keyValues:      []string{"a: world", "b.d: true"},
			initialConfig:  TestConfig{},
			expectedConfig: TestConfig{A: "world", B: NestedConfig{D: true}},
		},
		{
			name:      "success - override existing values",
			keyValues: []string{"a: new_value", "b.c: 999"},
			initialConfig: TestConfig{
				A: "old_value",
				B: NestedConfig{C: 111, D: true},
			},
			expectedConfig: TestConfig{
				A: "new_value",
				B: NestedConfig{C: 999, D: true}, // D should be untouched
			},
		},
		{
			name:           "success - weakly typed input (string to int)",
			keyValues:      []string{`b.c: "42"`}, // Note the quotes
			initialConfig:  TestConfig{},
			expectedConfig: TestConfig{B: NestedConfig{C: 42}},
		},
		{
			name:           "success - set a slice",
			keyValues:      []string{"f: [one, two, three]"},
			initialConfig:  TestConfig{},
			expectedConfig: TestConfig{F: []string{"one", "two", "three"}},
		},
		{
			name:           "success - set a map",
			keyValues:      []string{"g: {x: 10, y: 20}"},
			initialConfig:  TestConfig{},
			expectedConfig: TestConfig{G: map[string]int{"x": 10, "y": 20}},
		},
		{
			name:           "success - set special type",
			keyValues:      []string{"spez: \"banana\""},
			initialConfig:  TestConfig{},
			expectedConfig: TestConfig{S: SpecialType{A: 4242}},
		},
		{
			name:      "error - unused key",
			keyValues: []string{"nonexistent.key: value"},
			wantErr:   true,
		},
		{
			name:      "error - type mismatch (cannot decode string to int)",
			keyValues: []string{"b.c: not-an-int"},
			wantErr:   true,
		},
		{
			name:      "error - parsing error from helper",
			keyValues: []string{"a..b: c"},
			wantErr:   true,
		},
		{
			name:      "error - attempting to set a struct field with a non-map value",
			keyValues: []string{"b: 123"}, // b should be a map/struct, not an int
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Work on a copy to avoid modifying the test case's initial config
			config := tc.initialConfig

			err := ApplyKeyValues(tc.keyValues, &config)

			if tc.wantErr {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Did not expect an error but got one")
				assert.Equal(t, tc.expectedConfig, config, "Config struct does not match expected state")
			}
		})
	}
}
