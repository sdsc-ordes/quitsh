package versionupcmd

import (
	"bytes"
	"os"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/version"

	"github.com/spf13/cobra"
)

type versionUpArgs struct {
	compArgs general.ComponentArgs

	buildMeta  string
	prerelease string
}

func AddCmd(cl cli.ICLI, parent *cobra.Command) {
	var upArgs versionUpArgs
	versionUpCmd := &cobra.Command{
		Use:     "version-up [patch|minor|major]",
		Short:   "Bump the semantic versions on components.",
		PreRunE: cobra.MinimumNArgs(1),
		RunE: func(_cmd *cobra.Command, args []string) error {
			return versionUp(cl, args[0], &upArgs)
		},
	}

	versionUpCmd.Flags().
		StringArrayVarP(&upArgs.compArgs.ComponentPatterns,
			"components", "c", []string{"*"}, "Components matched by these patterns are listed.")

	versionUpCmd.Flags().
		StringVar(&upArgs.buildMeta,
			"build-meta", "",
			"The build meta part of the semantic version.",
		)

	versionUpCmd.Flags().
		StringVar(&upArgs.buildMeta,
			"prerelease-meta", "",
			"The prerelease meta part of the semantic version.",
		)

	parent.AddCommand(versionUpCmd)
}

func versionUp(cl cli.ICLI, level string, c *versionUpArgs) error {
	comps, _, _, err := cl.FindComponents(&c.compArgs)
	if err != nil {
		return err
	}

	allowedTypes := []string{"patch", "minor", "major"}
	if !slices.Contains(allowedTypes, level) {
		return errors.New("Version bump level '%v' is not one of '%v'", level, allowedTypes)
	}

	log.Infof("Do a %s version update on all components", level)
	for i := range comps {
		var e error

		vv := &comps[i].Config().Version.Version
		newVersion, e := version.Bump(vv,
			level,
			c.prerelease,
			c.buildMeta)

		if e != nil {
			return errors.AddContext(e,
				"could not version up component '%s'", comps[i].Name())
		}

		log.Info("Component:",
			"name", comps[i].Config().Name,
			"version-before", comps[i].Config().Version.String(),
			"version", newVersion)

		fileName := comps[i].ConfigFile()
		f, e := os.OpenFile(fileName, os.O_RDWR, fs.DefaultPermissionsFile)
		if e != nil {
			return e
		}
		defer f.Close()

		var node ast.Node
		cm := make(yaml.CommentMap)
		dec := yaml.NewDecoder(f, yaml.CommentToMap(cm))
		e = dec.Decode(&node)
		if e != nil {
			return e
		}

		p, e := yaml.PathString("$.version")
		if e != nil {
			return e
		}
		versionNode, e := p.FilterNode(node)
		if e != nil {
			log.Warn("Node '.version' is not found in '%v'", fileName)

			continue
		}

		if s, ok := versionNode.(*ast.StringNode); ok {
			s.Value = newVersion.String()
		} else {
			return errors.New("`.version` is not a string node in '%v'", fileName)
		}

		buf := bytes.NewBuffer(nil)
		enc := yaml.NewEncoder(
			buf,
			yaml.Indent(2), //nolint:mnd
			yaml.Flow(true),
			yaml.WithComment(cm))
		e = enc.Encode(node)
		if e != nil {
			return errors.AddContext(e, "could not marshal to file '%v'", fileName)
		}

		e = os.WriteFile(fileName, buf.Bytes(), fs.DefaultPermissionsFile)
		if e != nil {
			return e
		}
	}

	return nil
}
