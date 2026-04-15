package errorsfilter

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
)

func TestDropReported(t *testing.T) {
	var errA = errors.New("error A")
	var errB = errors.New("error B")
	var errC = errors.New("error C")
	var errD = errors.New("error D")

	tests := []struct {
		name     string
		input    error
		wantNil  bool
		wantErrs []error // errors expected to survive (checked via errors.Is)
		wantDrop []error // errors expected to be dropped
	}{
		{
			name:    "nil input",
			input:   nil,
			wantNil: true,
		},
		{
			name:     "single error passes through",
			input:    errA,
			wantErrs: []error{errA},
		},
		{
			name:     "multierror without ErrReported passes through",
			input:    multierror.Append(nil, errA, errB),
			wantErrs: []error{errA, errB},
		},
		{
			name:    "wrapped error is dropped",
			input:   WrapAsReported(errA),
			wantNil: true,
		},
		{
			name: "flat: two dropped in multierror",
			input: multierror.Append(nil,
				WrapAsReported(errA), // dropped
				errB,
				WrapAsReported(multierror.Append(nil, errC)), // dropped
			),
			wantErrs: []error{errB},
			wantDrop: []error{errA, errC},
		},
		{
			name: "nested: two dropped in multierror",
			input: &multierror.Error{
				Errors: []error{
					WrapAsReported(errA),
					errB,
					&multierror.Error{
						Errors: []error{
							WrapAsReported(errC),
							errD,
						},
					},
				},
			},
			wantErrs: []error{errB, errD},
			wantDrop: []error{errA, errC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterAlreadyReported(tt.input)

			if tt.wantNil {
				require.NoError(t, result)

				return
			}

			require.Error(t, result)

			for _, want := range tt.wantErrs {
				require.ErrorIs(t, result, want)
			}

			for _, dropped := range tt.wantDrop {
				require.NotErrorIs(t, result, dropped)
			}
		})
	}
}
