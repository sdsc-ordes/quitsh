package log

type ILog interface {
	// Trace will log an trace if trace is enabled.
	Trace(msg string, args ...any)

	// Debug will log an info if debug is enabled.
	Debug(msg string, args ...any)

	// Info will log an info.
	Info(msg string, args ...any)

	// Warn will log an info.
	Warn(msg string, args ...any)

	// Warn will log a warning for an error `err`.
	WarnE(err error, msg string, args ...any)

	// Error will log an error.
	Error(msg string, args ...any)

	// Error will log an error for `err`.
	ErrorE(err error, msg string, args ...any)

	// Panic will log and panic.
	Panic(msg string, args ...any)

	// PanicE will log and panic if `err` is not `nil`.
	PanicE(err error, msg string, args ...any)
}
