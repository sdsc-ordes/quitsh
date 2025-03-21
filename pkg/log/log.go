package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/mattn/go-isatty"
)

func ciRunning() bool {
	return os.Getenv("CI") == "true"
}

// Our global default logger. Yes singletons are code-smell,
// but we allow it for the logging functionality.
var levelVar slog.LevelVar        //nolint:gochecknoglobals // Accepted as only for logging.
var globalLogger = slog.Default() //nolint:gochecknoglobals // Accepted as only for logging.

// Setup sets up the default loggers .
func Setup(level string) error {
	err := SetLevel(level)
	if err != nil {
		return err
	}

	initLog()

	return nil
}

func initLog() {
	addSource := levelVar.Level() == slog.LevelDebug || false
	writer := os.Stderr

	slogOpts := &slog.HandlerOptions{
		AddSource: addSource,
		Level:     &levelVar,
	}

	haveColor := ciRunning() || isatty.IsTerminal(writer.Fd())

	opts := &devslog.Options{
		HandlerOptions:    slogOpts,
		NoColor:           !haveColor,
		SortKeys:          true,
		NewLineAfterLog:   true,
		MaxSlicePrintSize: 1000000, //nolint:mnd
	}

	handler := devslog.NewHandler(writer, opts)

	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)
}

const TraceLevel = slog.LevelDebug - 10

// Sets the log level on the logger.
func SetLevel(level string) error {
	var l slog.Level
	if level == "trace" {
		l = TraceLevel
	} else {
		err := l.UnmarshalText([]byte(level))

		if err != nil {
			return err
		}
	}

	levelVar.Set(l)

	if l <= slog.LevelDebug {
		// On debug (or smaller), we need to reinitialize for sources.
		initLog()
	}

	return nil
}

// Checks if debug is enabled on the log.
func IsDebug() bool {
	return levelVar.Level() == slog.LevelDebug
}

// createRecord creates a record with the proper source level.
func createRecord(skipFns int, level slog.Level, msg string, args ...any) slog.Record {
	var pc uintptr

	if levelVar.Level() == slog.LevelDebug {
		var pcs [1]uintptr
		runtime.Callers(
			3+skipFns, //nolint: mnd
			pcs[:],
		) // skip [Callers, createRecord, callers function]
		pc = pcs[0]
	}

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)

	return r
}

// Trace will log an trace if trace is enabled.
func Trace(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	ctx := context.Background()
	const level = TraceLevel

	if !globalLogger.Enabled(ctx, level) {
		return
	}
	r := createRecord(0, level, msg)
	_ = slog.Default().Handler().Handle(ctx, r)
}

// Debug will log an info if debug is enabled.
func Debug(msg string, args ...any) {
	ctx := context.Background()
	const level = slog.LevelDebug

	if !globalLogger.Enabled(ctx, level) {
		return
	}
	r := createRecord(0, level, msg, args...)
	_ = slog.Default().Handler().Handle(ctx, r)
}

// Info will log an info.
func Info(msg string, args ...any) {
	ctx := context.Background()
	const level = slog.LevelInfo

	if !globalLogger.Enabled(ctx, level) {
		return
	}
	r := createRecord(0, level, msg, args...)
	_ = slog.Default().Handler().Handle(ctx, r)
}

// Warn will log an info.
func Warn(msg string, args ...any) {
	warnS(0, msg, args...)
}

// Warn will log a warning for an error `err`.
func WarnE(err error, msg string, args ...any) {
	a := make([]interface{}, 0, 2+len(args)) //nolint: mnd
	a = append(a, "error", err)
	a = append(a, args...)
	warnS(0, msg, a...)
}

// Error will log an error.
func Error(msg string, args ...any) {
	errorS(0, msg, args...)
}

// Error will log an error for `err`.
func ErrorE(err error, msg string, args ...any) {
	errorES(0, err, msg, args...)
}

// Panic will log and panic.
func Panic(msg string, args ...any) {
	errorS(0, msg, args...)
	panic(msg)
}

// PanicE will log and panic if `err` is not `nil`.
func PanicE(err error, msg string, args ...any) {
	if err != nil {
		errorES(0, err, msg, args...)
		panic(err)
	}
}

func warnS(skipFns int, msg string, args ...any) {
	ctx := context.Background()
	const level = slog.LevelWarn

	if !globalLogger.Enabled(ctx, level) {
		return
	}
	r := createRecord(skipFns+1, level, msg, args...)
	_ = slog.Default().Handler().Handle(ctx, r)
}

func errorS(skipFns int, msg string, args ...any) {
	ctx := context.Background()
	const level = slog.LevelError

	if !globalLogger.Enabled(ctx, level) {
		return
	}
	r := createRecord(skipFns+1, level, msg, args...)
	_ = slog.Default().Handler().Handle(ctx, r)
}

func errorES(skipFns int, err error, msg string, args ...any) {
	a := make([]interface{}, 0, 2+len(args)) //nolint: mnd
	a = append(a, args...)
	a = append(a, "error", err)
	errorS(skipFns+1, msg, a...)
}
