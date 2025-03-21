package exec

// ExitCode0And1Success is a exit code handler which reports success for
// exit code 1 and 0.
func ExitCode0And1Success() ExitCodeHandler {
	return func(e *CmdError) error {
		switch {
		case e == nil:
			fallthrough
		case e.ExitCode() == 1:
			return nil
		default:
			return e
		}
	}
}

// ExitCode0And1SuccessVar is a exit code handler which reports success for
// exit code 1 and 0 and sets the boolean `success`.
func ExitCode0And1SuccessVar(success *bool) ExitCodeHandler {
	return func(e *CmdError) error {
		switch {
		case e == nil:
			*success = true

			return nil
		case e.ExitCode() == 1:
			*success = false

			return nil
		default:
			return e
		}
	}
}
