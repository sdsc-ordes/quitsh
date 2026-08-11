package image

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"carvel.dev/imgpkg/pkg/imgpkg/lockconfig"

	"github.com/containers/image/v5/docker/reference"
	"github.com/goccy/go-yaml"
	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/skopeo"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/registry"
)

func UploadImages(log log.ILog, sett *ImageSettings, pkgs ImagePackages) error {
	var err error

	ctx := skopeo.NewCtx(skopeo.WithEnableTLS(sett.Push.UseHTTPS))

	if sett.Push.CopyTo == "docker://" {
		// Only login when pushing to remote registry.
		logout, e := login(log, sett, ctx)
		if e != nil {
			return e
		}

		defer func() {
			if logout != nil {
				loe := logout()
				err = errors.Combine(err, loe)
			}
		}()
	}

	if sett.Push.AddLatestTag {
		err = addAuxiliaryRefsLatest(log, pkgs, "latest")
		if err != nil {
			return err
		}
	}

	err = processPackages(log, pkgs, ctx, sett)
	if err != nil {
		return err
	}

	err = writeImageInfo(log, pkgs)
	if err != nil {
		return err
	}

	err = writeImgpkgLock(log, pkgs)
	if err != nil {
		return err
	}

	return nil
}

func writeImgpkgLock(log log.ILog, pkgs ImagePackages) error {
	log.Info("Write 'imgpkg` lock files.")

	lock := lockconfig.NewEmptyImagesLock()

	for _, p := range pkgs {
		mainImgRef := p.ImageRefs[0].Ref

		refNamed, ok := mainImgRef.(ImageRefNamed)
		if !ok {
			return errors.New("image ref '%v' is not a 'ImageRefNamed' type", mainImgRef)
		}

		imgName := path.Base(refNamed.Name())
		if imgName != p.Name {
			return errors.New(
				"image base name '%v' should correspond to the image id '%v' "+
					"we use for 'ImagesLock` CRD (from imgpkg), otherwise 'kbld' will fail",
				imgName, p.Name)
		}

		ref := lockconfig.ImageRef{
			Image: p.ImageRefDigest.Ref.String(),
			Annotations: map[string]string{
				"kbld.carvel.dev/id": p.Name,
			},
		}
		lock.AddImageRef(ref)

		yamlFile := path.Join(
			path.Dir(p.ImageFile),
			fmt.Sprintf("%s-%s.imgpkg.lock.yaml", p.Component, p.ImageType),
		)

		err := lock.WriteToPath(yamlFile)
		if err != nil {
			return errors.AddContext(err, "could not write imgpgk lock file")
		}
	}

	return nil
}

// writeImageInfo writes two summary files, one YAML and
// one with the image ref (plain text).
func writeImageInfo(log log.ILog, pkgs ImagePackages) (err error) {
	log.Info("Write image info files.")

	writeFile := func(file string, v any) error {
		w, e := os.Create(file)
		if e != nil {
			return e
		}
		defer w.Close()
		enc := yaml.NewEncoder(w)

		return enc.Encode(v)
	}

	for _, p := range pkgs {
		yamlFile := path.Join(
			path.Dir(p.ImageFile),
			fmt.Sprintf("%s-%s.imginfo.yaml", p.Component, p.ImageType),
		)

		err = writeFile(yamlFile, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func login(
	log log.ILog,
	sett *ImageSettings,
	ctx skopeo.Context,
) (logout func() error, err error) {
	creds, err := sett.Push.ResolveCredentials()
	if err != nil {
		if ci.IsRunning() {
			return
		}

		log.Warn("No credentials provided for skopeo. " +
			"Relying on skopeo login performed.")
		err = nil

		return
	}

	log.Info("Skopeo login.")
	logout, err = ctx.Login(creds, sett.Push.RegistryDomain)
	if err != nil {
		return
	}

	return
}

//nolint:gocognit // TODO: refactor another time.
func processPackages(
	log log.ILog,
	pkgs ImagePackages,
	ctx skopeo.Context,
	sett *ImageSettings) error {
	var wg sync.WaitGroup
	errChan := make(chan error) // Unbuffered channel for errors

	// Error collector goroutine.
	var totalError error
	done := make(chan struct{})
	go func() {
		for e := range errChan {
			if e != nil {
				totalError = errors.Combine(totalError, e)
			}
		}
		close(done) // Signal completion when errChan is closed
	}()

	// Upload goroutine.
	upload := func(pkg *ImagePackage, imgRef ImageRef, auxiliary bool) {
		defer wg.Done()

		src := pkg.ImageFile

		// Set the destination image.
		dest := sett.Push.CopyTo + imgRef.String()

		// Only check if exists for the first original image, not auxiliary
		// ones like latest etc.
		// Also only do it when using "docker" transport.
		if !auxiliary &&
			sett.Push.CopyTo != "docker-daemon:" &&
			sett.Push.CopyTo != "containers-storage:" {
			exists := checkIfExists(log, ctx, dest)
			if exists {
				log.Warn("Image seems to exist.")
			} else {
				log.Info("Image seems not to exist.")
			}

			e := assertUseForceForPush(log, sett, exists, dest)
			if e != nil {
				errChan <- e

				return
			}
		}

		src = "docker-archive://" + src
		log.Info("Uploading image.", "src", src, "dest", dest)

		cpCtx := ctx.CopyCtx()

		cmd := []string{"--insecure-policy", src, dest}

		// TODO: Sign the image with cosign
		e := cpCtx.Check(cmd...)

		if e != nil {
			e = errors.AddContext(e, "could not upload image '%s' -> '%s'", src, dest)
			errChan <- e

			return
		}
	}

	// Process all uploads.
	for _, pkg := range pkgs {
		for j := range pkg.ImageRefs {
			auxiliary := j > 0
			wg.Add(1)

			if sett.Push.Parallel {
				go upload(pkg, pkg.ImageRefs[j].Ref, auxiliary)
			} else {
				upload(pkg, pkg.ImageRefs[j].Ref, auxiliary)
			}
		}

		log.Info("Get image digest.")
		err := pkg.SetDigest(ctx, false, sett.Push.CopyTo)
		if err != nil {
			return err
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()
	<-done // Wait until the error collector goroutine finishes.

	return totalError
}

// checkIfExists tries best-effort to detect if image exists, its not easy over skoepeo
// cause it handles different registries and not all report correctly.
func checkIfExists(log log.ILog, ctx skopeo.Context, dest string) (exists bool) {
	log.Info("Checking if image already exists.")

	inspCtx := ctx.InspectCtx()
	exists = true

	// When repository does not exist yet ->
	// we get exit code 1, an we need to handle
	// this case somehow.
	// TODO: Use a better tool/approach to check if the image is there, registry agnostic
	// (its not about gitlab!)
	// The workaround below is really ugly and prone to break!
	resp, e := inspCtx.GetCombinedWithEC(
		func(cmdError *exec.CmdError) error {
			if cmdError != nil && cmdError.ExitCode() == 2 {
				exists = false

				return nil
			}

			return cmdError
		},
		dest)

	if e != nil {
		if strings.Contains(resp, "403") ||
			strings.Contains(resp, "resource is denied") {
			// we have 403 which is ok, the registry does not yet exist.
			exists = false
		}

		log.WarnE(e,
			"could not check if image exists (fallback: true)",
			"image", dest,
			"response", resp,
		)
	}

	return exists
}

// addAuxiliaryRefsLatest adds auxiliary image refs with tag `tag`.
func addAuxiliaryRefsLatest(log log.ILog, pkgs ImagePackages, tag string) error {
	for i := range pkgs {
		debug.Assert(len(pkgs[i].ImageRefs) != 0, "Should have at least one image ref.")

		imgRef := pkgs[i].ImageRefs[0]
		namedRef, ok := imgRef.Ref.(ImageRefNamed)
		if !ok {
			log.Panic("Image ref '%v' is not a named image ref.", imgRef)
		}

		latestRef, err := reference.WithTag(namedRef, tag)
		if err != nil {
			return errors.AddContext(err, "could not add latest tag to image ref '%v'", imgRef)
		}

		pkgs[i].ImageRefs = append(pkgs[i].ImageRefs, ImageRefField{Ref: latestRef})
	}

	return nil
}

func assertUseForceForPush(
	log log.ILog,
	sett *ImageSettings,
	exists bool,
	imageName string,
) error {
	if sett.Push.RegistryType != registry.RegistryRelease {
		// To the test registry, allow anything.
		return nil
	}

	// Any other registry...
	if !exists { //nolint:nestif // intentional.
		if !ci.IsRunning() &&
			sett.Push.RegistryType == registry.RegistryRelease &&
			sett.Push.UseReleaseTag &&
			!sett.Push.Force {
			log.Error("The image does NOT EXIST and you want to push a RELEASE image! "+
				"WARNING: Luke you can use the force (--force) flag to proceed.",
				"image", imageName)

			return errors.New("image push denied")
		}

		log.Info(
			"Going to push image to registry.",
			"registry",
			sett.Push.RegistryType.String(),
			"image",
			imageName,
		)
		if !ci.IsRunning() {
			log.Warn("Waiting 10secs to proceed. -> Cancel now it that is a mistake.")
			time.Sleep(10 * time.Second) //nolint:mnd
		}
	} else {
		if sett.Push.UseReleaseTag {
			log.Error("The image EXISTS and you want to OVERWRITE a release image! "+
				"This is not allowed for safety, the image needs to be deleted manually!",
				imageName)

			return errors.New("image push denied")
		} else if !sett.Push.Force {
			log.Error("The image EXISTS and you want to OVERWRITE a image! "+
				"WARNING: Luke you can use the force (--force) flag to proceed.",
				"image", imageName)

			return errors.New("image push denied")
		}

		log.Warn("The image EXISTS and you are going to OVERWRITING an image!",
			"registry",
			sett.Push.RegistryType.String(),
			"image", imageName)

		if !ci.IsRunning() {
			log.Warn("Waiting 10secs to proceed. -> Cancel now it that is a mistake.")
			time.Sleep(10 * time.Second) //nolint:mnd
		}
	}

	return nil
}
