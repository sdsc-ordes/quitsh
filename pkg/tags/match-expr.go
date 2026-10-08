package tags

import (
	"fmt"
	"go/build/constraint"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
)

type (
	Tag struct {
		s string
	}

	Expr struct {
		expr string
		ex   constraint.Expr
	}
)

// NewTag creates a new tag.
func NewTag(tag string) Tag {
	return Tag{s: strings.ReplaceAll(tag, "-", ".")}
}

// NewExpr creates a new tag expression from `expr`.
// Since we parse it as a Go build line `//go:build <expr>` only `_` and `.` are
// allowed in tags, we allow also "-", which we internally replace with `.`.
func NewExpr(expr string) (ex Expr, err error) {
	ex.expr = strings.ReplaceAll(strings.TrimSpace(expr), "-", ".")
	if ex.expr == "" {
		return
	}

	ex.ex, err = constraint.Parse("//go:build " + ex.expr)

	return
}

// Interface implementation guard.
var _ fmt.Stringer = (*Expr)(nil)

// String returns the expression.
func (e *Expr) String() string {
	return e.expr
}

// Matches reports if this tag expression matches for the given `tags`.
func (e *Expr) Matches(tags []Tag) bool {
	if e.ex == nil {
		return true
	}

	return e.ex.Eval(func(tag string) bool {
		return slices.Contains(tags, Tag{s: tag})
	})
}

// Interface implementation guard.
var _ yaml.InterfaceUnmarshaler = (*Tag)(nil)

// UnmarshalYAML implements [yaml.InterfaceUnmarshaler].
func (v *Tag) UnmarshalYAML(unmarshal func(any) error) (err error) {
	err = unmarshal(&v.s)
	if err != nil {
		return
	}

	return
}

// Interface implementation guard.
var _ yaml.InterfaceMarshaler = (*Tag)(nil)

// MarshalYAML implements [yaml.InterfaceMarshaler].
// NOTE: needs to be value-receiver to be called!
func (v Tag) MarshalYAML() (any, error) {
	return v.s, nil
}

// Interface implementation guard.
var _ yaml.InterfaceUnmarshaler = (*Expr)(nil)

// UnmarshalYAML implements [yaml.InterfaceUnmarshaler].
func (v *Expr) UnmarshalYAML(unmarshal func(any) error) (err error) {
	err = unmarshal(&v.expr)
	if err != nil {
		return
	}

	*v, err = NewExpr(v.expr)
	if err != nil {
		return
	}

	return
}

// Interface implementation guard.
var _ yaml.InterfaceMarshaler = (*Expr)(nil)

// MarshalYAML implements [yaml.InterfaceMarshaler].
// NOTE: needs to be value-receiver to be called!
func (v Expr) MarshalYAML() (any, error) {
	return v.expr, nil
}
