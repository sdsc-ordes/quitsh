package containerfilerunner

import (
	"fmt"
	"os"
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/image"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const ContainerfileRunnerID = "quitsh::image-containerfile"

type (
	ContainerfileBuildRunner struct {
		config   *RunnerConfigContainerfile
		settings *config.ImageSettings
		options  *opts
	}
)

// NewContainerfileBuildRunner constructs a new ContainerfileBuildRunner with its own config.
func NewContainerfileBuildRunner(
	config step.AuxConfig,
	settings *config.ImageSettings,
	options *opts,
) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &ContainerfileBuildRunner{
		config:   common.Cast[*RunnerConfigContainerfile](config),
		settings: settings,
		options:  options,
	}, nil
}

// Interface implementation guard.
var _ runner.IRunner = (*ContainerfileBuildRunner)(nil)

// ID implements [runner.IRunner].
func (*ContainerfileBuildRunner) ID() runner.RegisterID {
	return ContainerfileRunnerID
}

// Run implements [runner.IRunner].
func (r *ContainerfileBuildRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()
	config := comp.Config()
	log.Info("Starting Containerfile image build for component.", "component", config.Name)

	if r.settings.Build.SkipBuild && !r.settings.Push.Enable {
		return errors.New("you need at least to build or push the images")
	}

	// Either build the images given on the command line
	// or build the images given in the component config.
	var imageTypes []image.Type
	if r.settings.Build.ImageTypes != nil {
		imageTypes = r.settings.Build.ImageTypes
	} else {
		imageTypes = image.GetAllImageTypes()
	}

	pkgs := image.ImagePackages{}
	for i := range imageTypes {
		p := comp.ImagesContainerfile(imageTypes[i])
		if !fs.Exists(p) {
			continue
		}

		err := r.settings.Push.SourceInfo.Resolve(ctx.Git())
		if err != nil {
			return errors.AddContext(err, "could not resolve source information")
		}

		imgPkgName := r.options.packageName(comp.Name(), imageTypes[i])
		imageRef, e := image.NewImageRef(
			r.settings.Push.RegistryDomain,
			r.settings.Push.RegistryBasePathFmt,
			imgPkgName,
			comp.Version(),
			r.settings.Push.RegistryType,
			r.settings.Push.SourceInfo.CommitRef,
			r.settings.Push.UseReleaseTag,
		)
		if e != nil {
			return e
		}

		pkgs = append(pkgs,
			&image.ImagePackage{
				Version:       comp.Version().String(),
				Name:          imgPkgName,
				ContainerFile: p,
				ImageType:     imageTypes[i],
				ImageFile:     comp.OutImageDir(imgPkgName),
				ImageRefs:     []image.ImageRefField{{Ref: imageRef}},
				// not used
				NixPackage:     "",
				NixInstallable: "",
			})
	}

	if len(pkgs) == 0 {
		return errors.New("no image container files found in '%s' for types %q",
			comp.ImagesDir(), imageTypes)
	}

	if !r.settings.Build.SkipBuild {
		err := buildImages(log, ctx.Root(), pkgs)
		if err != nil {
			return err
		}
	}

	if r.settings.Push.Enable {
		err := image.UploadImages(log, r.settings, pkgs)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildImages(
	log log.ILog,
	rootDir string,
	pkgs image.ImagePackages,
) error {
	// We build with a special storage driver,
	// all buildah commands need to know this.
	buildctx := exec.NewCmdCtxBuilder().
		Cwd(rootDir).
		BaseCmd("buildah").
		Env(
			"BUILDAH_FORMAT=docker",
			"BUILDAH_ISOLTAION=chroot",
			"STORAGE_DRIVER=vfs").
		Build()

	for i := range pkgs {
		log.Info("Building image.",
			"image",
			pkgs[i].ImageFile,
			"ref", pkgs[i].ImageRefs[0].Ref.String())

		cmd := []string{
			"build",
			"--build-arg", fmt.Sprintf("BUILD_VERSION=%v", pkgs[i].Version),
			"-f", pkgs[i].ContainerFile,
			"-t", pkgs[i].ImageRefs[0].Ref.String(),
			"--isolation=chroot",
			"--storage-driver=vfs",
			"--format=docker",
			rootDir,
		}

		err := buildctx.Check(cmd...)
		if err != nil {
			return err
		}

		dest := pkgs[i].ImageFile
		err = os.MkdirAll(path.Dir(dest), fs.DefaultPermissionsDir)
		if err != nil {
			return err
		}

		if fs.Exists(dest) {
			err = os.Remove(dest)
			if err != nil {
				return err
			}
		}

		cmd = []string{
			"push",
			"--storage-driver=vfs",
			"--format=docker",
			pkgs[i].ImageRefs[0].Ref.String(),
			"docker-archive://" + dest,
		}

		err = buildctx.Check(cmd...)
		if err != nil {
			return err
		}
	}

	return nil
}
