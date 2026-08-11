package image

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/registry"
	"github.com/sdsc-ordes/quitsh/pkg/secret"
)

type ImageSettings struct {
	Build ImageSettingsBuild `yaml:"build"`
	Push  ImageSettingsPush  `yaml:"push"`

	Args []string `yaml:"args"` // Additional arguments forwarded to the tool.
}

type ImageSettingsBuild struct {
	SkipBuild bool `yaml:"skip"`

	ImageTypes []Type `yaml:"imageTypes,omitempty"`

	// TODO: BuildTypes and EnvironmentType are not supported right now,
	// we always build production, because of flakes not being
	// configurable, or not with some whoops...
}

type (
	ImageSettingsPush struct {
		Enable bool `yaml:"enable"`

		SourceInfo ImageSourceInfo `yaml:"sourceInfo"`

		// The registry name specifying either `release` or `temporary` to use to
		// upload the image to.
		RegistryType registry.Type `yaml:"registryType"`

		// The default domain and base name.
		RegistryDomain      string `yaml:"registryDomain"`
		RegistryBasePathFmt string `yaml:"registryBasePathFmt"`

		// If the image should have a release tag (only `<version>)
		// or use `<version>-<commitSHA>`.
		UseReleaseTag bool `yaml:"useReleaseTag"`

		// If also a `:latest` tag is pushed.
		AddLatestTag bool `yaml:"addLatestTag"`

		// The transport in skopeo copy
		// (by default `docker` (remote) or local to `containers-storage` and `docker-daemon`)
		// See reference: https://github.com/containers/image/blob/main/docs/containers-transports.5.md
		CopyTo string `yaml:"copyTo" default:"docker://"`

		CredentialsEnv secret.CredentialsEnv `yaml:"credentialsEnv"`

		// Under the following condition you can force an upload
		// because we disallow to prevent accidents of overwriting:
		// - Push + Release + image does not exist yet!
		// - Push + Debug + and image does exist.
		Force bool `yaml:"force"`

		// If we should not use TLS verification (https)
		// when talking to the registry.
		UseHTTPS bool `yaml:"useHTTPS" default:"true"`

		// If the upload is parallelized.
		Parallel bool `yaml:"parallel"`
	}

	ImageSourceInfo struct {
		// The commit reference where this image has been produced.
		CommitRef string `yaml:"commitRef"`
	}
)

// ResolveCredentials resolves the credentials from the env. variables.
func (s *ImageSettingsPush) ResolveCredentials() (secret.Credentials, error) {
	c, e := secret.NewCredentials(s.CredentialsEnv)

	return c, errors.AddContext(e, "could not resolve credentials")
}

// Resolve resolves source information from `git`.
func (s *ImageSourceInfo) Resolve(gitx git.Context) (err error) {
	if s.CommitRef == "" {
		s.CommitRef, err = gitx.Get("rev-parse", "HEAD")
		if err != nil {
			return
		}
	}

	return nil
}
