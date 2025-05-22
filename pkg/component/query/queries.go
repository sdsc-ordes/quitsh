package query

import (
	"fmt"
	"path"
	"strings"

	comp "github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/bmatcuk/doublestar/v4"
)

type queryOptions struct {
	dirFilter      fs.DirFilter
	configFileName string
}

// Find finds all components in directory `root` and loads them.
// Some directories are by default ignored.
//
//nolint:gocognit,funlen
func Find(
	rootDir string,
	filter func(compName string, root string) (matches bool, err error),
	creator comp.ComponentCreator,
	opts ...Option,
) (comps []*comp.Component, all []*comp.Component, err error) {
	if creator == nil {
		log.Panic("Component creator not given.")
	}

	// Apply options.
	queryOpts := queryOptions{}
	for i := range opts {
		err = opts[i](&queryOpts)

		if err != nil {
			return nil, nil, err
		}
	}

	if queryOpts.dirFilter == nil {
		err = WithPathFilterDefault()(&queryOpts)

		if err != nil {
			return nil, nil, err
		}
	}

	if queryOpts.configFileName == "" {
		queryOpts.configFileName = comp.ConfigFilename
	}

	rootDir = fs.MakeAbsolute(rootDir)
	files, traversedFiles, err := fs.FindFiles(
		rootDir,
		[]string{queryOpts.configFileName},
		nil,
		fs.WithPathFilter(queryOpts.dirFilter),
		fs.WithMatchFileNameOnly(),
	)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("Traversed fs.", "count", traversedFiles)
	log.Debug("Found components.", "count", len(files))

	visitedComps := map[string]string{}

	gitx := git.NewCtx(rootDir)
	for _, componentFile := range files {
		root := path.Dir(componentFile)

		ignored, e := gitx.IsIgnored(componentFile)
		if e != nil {
			log.WarnE(e, "could not check if file '%s' is ignored.")
		}

		if ignored {
			log.Trace("Component ignored by Git.", "root", root)

			continue
		}

		c, e := config.LoadFromFile[comp.Config](componentFile)
		if e != nil {
			log.Warn("Could not load config.", "config", componentFile)
			err = errors.Combine(err, e)

			continue
		}

		log.Debug("Loaded component.", "component", c.Name, "root", root)

		if p, exists := visitedComps[c.Name]; exists {
			return nil, nil, errors.New(
				"component with name '%v' at path '%v'"+
					"already exists at path '%v'", c.Name, root, p,
			)
		}

		comp, e := creator(&c, root)
		if e != nil {
			err = errors.Combine(err, e)

			continue
		}

		all = append(all, comp)
		visitedComps[c.Name] = root

		if filter != nil {
			matches, ef := filter(c.Name, root)
			if ef != nil {
				err = errors.Combine(err, ef)

				continue
			}

			if !matches {
				log.Debug("Ignoring filtered component.", "name", c.Name)

				continue
			}
		}

		comps = append(comps, comp)
	}

	return comps, all, err
}

// splitIntoIncludeAndExcludes splits the patterns into
// include and exclude patterns.
func splitIntoIncludeAndExcludes(patterns []string) (incls []string, excls []string) {
	incls = make([]string, 0, len(patterns))
	excls = make([]string, 0, len(patterns))

	for i := range patterns {
		l := &incls

		startIdx := 0
		if strings.HasPrefix(patterns[i], "!") {
			l = &excls
			startIdx = 1
		}

		// Escaping with `\!`, split it off.
		if strings.HasPrefix(patterns[i], "\\!") {
			startIdx = 1
		}

		(*l) = append((*l), patterns[i][startIdx:])
	}

	return
}

// Find finds components in `rootDir` with names matched by `patterns`.
func FindByPatterns(
	rootDir string,
	patterns []string,
	minCount int,
	creator comp.ComponentCreator,
) (comps []*comp.Component, all []*comp.Component, err error) {
	incls, excls := splitIntoIncludeAndExcludes(patterns)
	log.Info("Find components by patterns.",
		"includes", incls, "excludes", excls, "root", rootDir)

	filter := func(name string, _ string) (matches bool, err error) {
		include := false
		exclude := false
		for _, pattern := range incls {
			matches, err = doublestar.Match(pattern, name)
			if err != nil {
				return
			} else if matches {
				include = true

				break
			}
		}

		for _, pattern := range excls {
			matches, err = doublestar.Match(pattern, name)
			if err != nil {
				return
			} else if matches {
				exclude = true

				break
			}
		}

		return include && !exclude, nil
	}

	comps, all, err = Find(rootDir, filter, creator)
	if err != nil {
		log.ErrorE(err, "Could not find and load components.", "root", rootDir)
	} else if len(comps) < minCount {
		log.Error("Could not find min. components",
			"root", rootDir, "count", len(comps), "min-count", minCount)

		err = fmt.Errorf(
			"min. count '%v' components not found in "+
				"'%v' (found only '%v')",
			minCount, rootDir, len(comps))
	}

	return comps, all, err
}

// Find the matching component inside directory `dir`.
// Note: Only `WithComponentConfigFilename` makes sense for `opts`.
func FindInside(
	dir string,
	creator comp.ComponentCreator,
	opts ...Option,
) (*comp.Component, error) {
	if creator == nil {
		log.Panic("Component creator not given.")
	}

	// Apply options.
	queryOpts := queryOptions{}
	for i := range opts {
		err := opts[i](&queryOpts)

		if err != nil {
			return nil, err
		}
	}

	if queryOpts.configFileName == "" {
		queryOpts.configFileName = comp.ConfigFilename
	}

	d := fs.MakeAbsolute(dir)
	if !fs.Exists(d) {
		return nil, fmt.Errorf("directory does not exists '%v'", d)
	}

	for fs.Exists(d) {
		f := path.Join(d, queryOpts.configFileName)
		log.Debug(f)

		if fs.Exists(f) {
			c, err := config.LoadFromFile[comp.Config](f)

			if err != nil {
				return nil, err
			}

			return creator(&c, d)
		}

		prev := d
		d = path.Dir(d)
		log.Debug("Searching up.", "d", d)
		if d == prev {
			break
		}
	}

	return nil, fmt.Errorf(
		"could not find any component upwards starting from '%v'",
		dir,
	)
}
