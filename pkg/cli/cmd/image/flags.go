package image

// Helper functions for [cobra.Command] when adding an `image` stage command.

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/image"
	"github.com/sdsc-ordes/quitsh/pkg/registry"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type (
	imgTypesParseWrapper struct {
		values *[]image.Type
	}

	copyToParseWrapper struct {
		value *string
	}
)

// SetImageFlags setts all image settings which are set by the global config `imageSettings`.
func SetImageFlags(cmd *cobra.Command, imageSettings *config.ImageSettings) {
	imgType := imgTypesParseWrapper{values: &imageSettings.Build.ImageTypes}

	cmd.Flags().
		VarP(&imgType,
			"image-types", "i",
			fmt.Sprintf("The image types to build (%s) (comma-separated).",
				image.GetImageTypesHelp()))

	cmd.Flags().
		BoolVar(&imageSettings.Build.SkipBuild,
			"skip-build", imageSettings.Build.SkipBuild,
			"If the image should be not be build.")

	AddPushFlagsGeneral(cmd, imageSettings)
}

// AddPushFlagsGeneral adds common flags for general image settings.
func AddPushFlagsGeneral(cmd *cobra.Command, setts *config.ImageSettings) {
	s := cmd.Flags()

	copyTo := copyToParseWrapper{value: &setts.Push.CopyTo}

	s.BoolVar(&setts.Push.Enable,
		"push", setts.Push.Enable,
		"If the image should be pushed to registry.")

	s.Var(
		&setts.Push.RegistryType,
		"registry-type",
		fmt.Sprintf(
			"The registry type specifying the registry to which the image is uploaded (%v).",
			registry.GetAllRegistryTypes(),
		),
	)

	s.StringVar(
		&setts.Push.RegistryDomain,
		"registry-domain",
		setts.Push.RegistryDomain,
		"Overwrite the image registry domain name.",
	)
	s.StringVar(
		&setts.Push.RegistryBasePathFmt,
		"registry-base-name",
		setts.Push.RegistryBasePathFmt,
		"Overwrite the image registry base path fmt.",
	)

	s.BoolVar(&setts.Push.Force,
		"force",
		setts.Push.Force,
		"Under the following condition you can force an upload\n"+
			"because we disallow the push to prevent accidents of overwriting:\n"+
			"- Release registry + !UseReleaseTag + image exist\n"+
			"Pushing/overwriting to temporary registry is always allowed.",
	)

	s.BoolVar(&setts.Push.UseReleaseTag,
		"use-release-tag",
		setts.Push.UseReleaseTag,
		"If the image tag should be a release tag \n"+
			"(e.g. semantic version: '1.2.3' instead of '1.2.3-<git-hash>').\n"+
			"On any other registry than 'release' this is ignored.")

	s.BoolVar(&setts.Push.AddLatestTag,
		"add-latest-tag",
		setts.Push.AddLatestTag,
		"If also a latest tag 'latest' image ref additionally should be added.")

	s.StringVar(&setts.Push.CredentialsEnv.UserEnv,
		"credential-user-env",
		setts.Push.CredentialsEnv.UserEnv,
		"The username environment variable for the registry to upload the image.")
	s.StringVar(&setts.Push.CredentialsEnv.TokenEnv,
		"credential-token-env",
		setts.Push.CredentialsEnv.TokenEnv,
		"The token environment variable for the registry to upload the image.")

	s.BoolVar(&setts.Push.Parallel,
		"parallel",
		setts.Push.Parallel,
		"If the push (currently only) is done in parallel.")

	s.Var(&copyTo,
		"copy-to",
		"The destination transport, either copy to remote `docker://` or "+
			"local `containers-storage:` or `docker-daemon:`.")

	s.BoolVar(&setts.Push.UseHTTPS,
		"use-https",
		setts.Push.UseHTTPS,
		"If HTTPS is used to talk the registry.")
}

// Interface implementation guard.
var _ pflag.Value = (*imgTypesParseWrapper)(nil)

// String implements [pflag.Value].
func (i *imgTypesParseWrapper) String() string {
	return fmt.Sprintf("%q", *i.values)
}

// Set implements [pflag.Value].
func (i *imgTypesParseWrapper) Set(s string) error {
	for v := range strings.SplitSeq(s, ",") {
		vv, err := image.NewType(strings.TrimSpace(v))
		if err != nil {
			return err
		}
		*i.values = append(*i.values, vv)
	}

	return nil
}

// Type implements [pflag.Value].
func (i *imgTypesParseWrapper) Type() string {
	return "string"
}

// Interface implementation guard.
var _ pflag.Value = (*copyToParseWrapper)(nil)

// String implements [pflag.Value].
func (i *copyToParseWrapper) String() string {
	return *i.value
}

// Set implements [pflag.Value].
func (i *copyToParseWrapper) Set(s string) error {
	if s != "containers-storage:" &&
		s != "docker://" &&
		s != "docker-daemon:" {
		return errors.New("argument `CopyTo` transport is wrong, see help")
	}

	*i.value = s

	return nil
}

// Type implements [pflag.Value].
func (i *copyToParseWrapper) Type() string {
	return "string"
}
