package strs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitLines(t *testing.T) {
	s := `a
b
c`
	ls := SplitLines(s)
	assert.Equal(t, []string{"a", "b", "c"}, ls)

	s = "a\nb\r\nc\n"
	ls = SplitLines(s)
	assert.Equal(t, []string{"a", "b", "c", ""}, ls)
}

func TestIndent(t *testing.T) {
	s := `a
b
c`
	ls := Indent(s, "| ")
	assert.Equal(t, "| a\n| b\n| c", ls)

	s = "a\nb\r\nc\n"
	ls = Indent(s, "| ")
	assert.Equal(t, "| a\n| b\n| c\n| ", ls)
}

func TestSplitAndTrim(t *testing.T) {
	s := "a,b,  c, ban nana,d"
	ls := SplitAndTrim(s, ",")
	assert.Equal(t, []string{"a", "b", "c", "ban nana", "d"}, ls)

	s = "a|b,d|c"
	ls = SplitAndTrim(s, "|")
	assert.Equal(t, []string{"a", "b,d", "c"}, ls)
}
