package errors

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

func New(format string, args ...any) error {
	//TODO: Log here the stack with go-errors package maybe.
	//      Unwrap loop with errors.Unwrap
	//      in main to print all go-errors errors stack in
	//      all errors.
	return fmt.Errorf(format, args...)
}

// Combine combines multiple errors into one error.
func Combine(err error, errs ...error) error {
	// Early returning (optimization)
	if len(errs) == 0 {
		return err
	} else if err == nil && len(errs) == 1 {
		return errs[0]
	}

	// Otherwise append all together and see what turns out.
	return multierror.Append(err, errs...).ErrorOrNil()
}

// AddContext adds a formatted error to any `err`.
func AddContext(err error, message string, args ...any) error {
	if err == nil {
		return nil
	}

	return Combine(New(message, args...), err)
}
