package version

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBump(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		level      string
		prerelease string
		buildMeta  string
		want       string
		wantErr    bool
	}{
		{
			name:  "bump patch",
			input: "1.2.3",
			level: "patch",
			want:  "1.2.4",
		},
		{
			name:  "bump minor",
			input: "1.2.3",
			level: "minor",
			want:  "1.3.0",
		},
		{
			name:  "bump major",
			input: "1.2.3",
			level: "major",
			want:  "2.0.0",
		},
		{
			name:       "bump patch with prerelease",
			input:      "1.2.3",
			level:      "patch",
			prerelease: "beta.1",
			want:       "1.2.4-beta.1",
		},
		{
			name:      "bump patch with build metadata",
			input:     "1.2.3",
			level:     "patch",
			buildMeta: "exp.sha.5114f85",
			want:      "1.2.4+exp.sha.5114f85",
		},
		{
			name:       "bump patch with prerelease and build metadata",
			input:      "1.2.3",
			level:      "patch",
			prerelease: "rc.1",
			buildMeta:  "build.11",
			want:       "1.2.4-rc.1+build.11",
		},
		{
			name:       "bump minor with prerelease",
			input:      "1.2.3",
			level:      "minor",
			prerelease: "alpha.1",
			want:       "1.3.0-alpha.1",
		},
		{
			name:      "bump minor with build metadata",
			input:     "1.2.3",
			level:     "minor",
			buildMeta: "20251001",
			want:      "1.3.0+20251001",
		},
		{
			name:       "bump minor with prerelease and build metadata",
			input:      "1.2.3",
			level:      "minor",
			prerelease: "rc.2",
			buildMeta:  "sha.abcdef",
			want:       "1.3.0-rc.2+sha.abcdef",
		},
		{
			name:       "bump major with prerelease",
			input:      "1.2.3",
			level:      "major",
			prerelease: "beta.7",
			want:       "2.0.0-beta.7",
		},
		{
			name:      "bump major with build metadata",
			input:     "1.2.3",
			level:     "major",
			buildMeta: "build.20251001",
			want:      "2.0.0+build.20251001",
		},
		{
			name:       "bump major with prerelease and build metadata",
			input:      "1.2.3",
			level:      "major",
			prerelease: "rc.99",
			buildMeta:  "exp.sha.999999",
			want:       "2.0.0-rc.99+exp.sha.999999",
		},
		{
			name:    "invalid level",
			input:   "1.2.3",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := version.NewVersion(tt.input)
			require.NoError(t, err)

			got, err := Bump(v, tt.level, tt.prerelease, tt.buildMeta)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.String())
			}
		})
	}
}
