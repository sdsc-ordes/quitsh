package component

import (
	"path"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/image"
)

// These paths are relative to the component folder.
// Use the below functions to get paths relative to the component.
const (
	ConfigFilename = ".component.yaml"
)

// OutDir returns the output directory of the component.
func (c *Component) OutDir(p ...string) string {
	return c.RelOutPath("", p...)
}

// OutBuildDir returns the directory of the components build folder.
func (c *Component) OutBuildDir(p ...string) string {
	return c.RelOutPath(fs.OutBuildDir, p...)
}

// OutBuildBinDir returns the directory of the components build binary directory.
func (c *Component) OutBuildBinDir(p ...string) string {
	return c.RelOutPath(fs.OutBuildBinDir, p...)
}

// OutBuildShareDir returns the directory of the components build share directory.
func (c *Component) OutBuildShareDir(p ...string) string {
	return c.RelOutPath(fs.OutBuildShareDir, p...)
}

// OutBuildDocsDir returns the directory of the components build docs directory.
func (c *Component) OutBuildDocsDir(p ...string) string {
	return c.RelOutPath(fs.OutBuildDocsDir, p...)
}

// OutCoverageDataDir returns the directory of the coverage data output directory.
func (c *Component) OutCoverageDataDir(p ...string) string {
	return c.RelOutPath(fs.OutCoverageDataDir, p...)
}

// OutCoverageBinDir returns the directory of the coverage binary output directory.
func (c *Component) OutCoverageBinDir(p ...string) string {
	return c.RelOutPath(fs.OutCoverageBinDir, p...)
}

// OutPackageDir returns the directory of the package output directory.
func (c *Component) OutPackageDir(p ...string) string {
	return c.RelOutPath(fs.OutPackageDir, p...)
}

// OutImageDir returns the directory of the image output directory.
func (c *Component) OutImageDir(p ...string) string {
	return c.RelOutPath(fs.OutImageDir, p...)
}

// DocsDir returns the directory of the components docs folder.
func (c *Component) DocsDir(p ...string) string {
	return c.RelPath(fs.DocsDir, p...)
}

// ImagesDir returns the directory of the components images folder.
func (c *Component) ImagesDir(p ...string) string {
	return c.RelPath(fs.ImagesDir, p...)
}

// ImagesContainerfile returns the image file of the component for the image type.
func (c *Component) ImagesContainerfile(imageType image.Type) string {
	return c.ImagesDir(imageType.String(), "Containerfile")
}

// RelPath concats paths relative to the component's root directory.
func (c *Component) RelPath(p string, ps ...string) string {
	l := make([]string, 0, len(p)+1)
	l = append(l, c.root, p)
	l = append(l, ps...)

	return path.Join(l...)
}

// RelOutPath concats paths relative to the component's out directory.
func (c *Component) RelOutPath(p string, ps ...string) string {
	l := make([]string, 0, len(p)+1)
	l = append(l, c.outDir, p)
	l = append(l, ps...)

	return path.Join(l...)
}
