package secret

import (
	"encoding/json"
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

// RedactedString represents a redacted string.
type RedactedString string

// Interface implementation guard.
var _ fmt.Stringer = (*RedactedString)(nil)

func (r *RedactedString) String() string {
	return fmt.Sprintf("<redacted-%d-chars>", len(*r))
}

// Interface implementation guard.
var _ json.Marshaler = (*RedactedString)(nil)

func (r *RedactedString) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "\"%s\"", r.String()), nil
}

func (r *RedactedString) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *RedactedString) MarshalBinary() ([]byte, error) {
	return []byte(r.String()), nil
}

// Interface implementation guard.
var _ config.UnmarshalerMapstruct = (*RedactedString)(nil)

// UnmarshalMapstruct implements [config.UnmarshalerMapstruct].
func (r *RedactedString) UnmarshalMapstruct(data any) (err error) {
	secret, ok := data.(string)
	if !ok {
		return errors.New("cannot unmarshal secure secret: got '%T'", data)
	}

	*r = RedactedString(secret)

	return nil
}
