package recache

import (
	"regexp"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type List []*regexp.Regexp

type Cache struct {
	fullMatch bool
	patterns  map[string]*regexp.Regexp
}

// Create a new regex cache.
func NewCache(fullMatch bool) Cache {
	return Cache{fullMatch: fullMatch, patterns: make(map[string]*regexp.Regexp)}
}

// Get gets a pattern `pattern` or it compiles one and adds it to the cache.
func (r *Cache) Get(patterns ...string) (regexes List, err error) {
	for _, pattern := range patterns {
		re, exists := r.patterns[pattern]
		if exists {
			regexes = append(regexes, re)

			continue
		}

		re, e := r.Add(pattern)
		if e != nil {
			err = errors.Combine(err, e)

			continue
		}

		regexes = append(regexes, re)
	}

	return
}

// Add adds a regex pattern.
func (r *Cache) Add(pattern string) (*regexp.Regexp, error) {
	if r.fullMatch {
		if !strings.HasPrefix(pattern, "^") {
			log.Warn(
				"Regex is only allowed to be a full match -> adding '^' (change to hide this warning)",
				"regex",
				pattern,
			)
			pattern = "^" + pattern
		}
		if !strings.HasSuffix(pattern, "$") {
			log.Warn(
				"Regex is only allowed to be a full match -> adding '$' (change to hide this warning).",
				"regex",
				pattern,
			)
			pattern += "$"
		}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.AddContext(err, "could not compile regex '%s'", pattern)
	}

	r.patterns[pattern] = re

	return re, nil
}

// Match returns `true` if any regex matches for `s`.
func (r List) Match(s string) bool {
	for idx := range r {
		if r[idx].MatchString(s) {
			log.Debug("Regex matches.", "regex", r[idx].String())

			return true
		}
	}

	return false
}
