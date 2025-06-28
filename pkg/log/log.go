package log

import (
	"os"
	"time"

	chlog "github.com/charmbracelet/log"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

func ciRunning() bool {
	return os.Getenv("CI") == "true"
}

const TraceLevel = chlog.DebugLevel - 10

var ForceColorInCI = true

// Our global default logger. Yes singletons are code-smell,
// but we allow it for the logging functionality.
var globalLogger = logger{l: chlog.New(os.Stderr)} //nolint:gochecknoglobals // Accepted as only for logging.

// Setup sets up the default loggers .
func Setup(level string) (err error) {
	l, err := initLog(level)
	if err != nil {
		return
	}

	globalLogger = l

	return
}

func initLog(level string) (logger, error) {
	if ciRunning() && ForceColorInCI {
		// We force here a color profile
		lipgloss.SetColorProfile(termenv.TrueColor)
	}

	l := chlog.NewWithOptions(
		os.Stderr, chlog.Options{
			ReportCaller:    false,
			ReportTimestamp: true,
			TimeFormat:      time.TimeOnly,
		})

	err := setLevel(l, level)
	if err != nil {
		return logger{}, err
	}

	styles := getStyles()
	l.SetStyles(styles)

	return logger{l: l}, nil
}

func getStyles() *chlog.Styles {
	styles := chlog.DefaultStyles()

	styles.Levels[TraceLevel] = lipgloss.NewStyle().
		SetString("TRACE").
		Padding(0, 1, 0, 1).
		Foreground(lipgloss.Color("#a4a4a4")).Bold(true)

	styles.Levels[chlog.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Padding(0, 1, 0, 1).
		Foreground(lipgloss.Color("#00e6ff")).Bold(true)

	styles.Levels[chlog.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Padding(0, 1, 0, 1).
		Foreground(lipgloss.Color("#00c900")).Bold(true)

	styles.Levels[chlog.WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("#ff7400")).
		Foreground(lipgloss.Color("0")).Bold(true)

	styles.Levels[chlog.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("#ff0000")).
		Foreground(lipgloss.Color("0")).Bold(true)

	styles.Prefix = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#007399", Dark: "#44b5c3"}).
		Italic(true)

	styles.Caller = lipgloss.NewStyle().Italic(true)

	styles.Key = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#007399", Dark: "#44b5c3"}).
		Bold(true)

	styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	styles.Keys["error"] = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	styles.Values["err"] = lipgloss.NewStyle().Bold(true)
	styles.Values["error"] = lipgloss.NewStyle().Bold(true)

	return styles
}

func SetLevel(level string) error {
	return setLevel(globalLogger.l, level)
}

// SetLevel sets the log level on the logger.
func setLevel(logger *chlog.Logger, level string) (err error) {
	var l chlog.Level
	if level == "trace" {
		l = TraceLevel
	} else {
		l, err = chlog.ParseLevel(level)
	}

	if err != nil {
		return errors.AddContext(err, "Could not parse log level.")
	}

	if l <= chlog.DebugLevel {
		logger.SetReportCaller(true)
	}

	logger.SetLevel(l)

	return nil
}

// IsDebug checks if debug is enabled on the log.
func IsDebug() bool {
	return globalLogger.l.GetLevel() == chlog.DebugLevel
}

// Tracef will log a trace info with formatting.
func Tracef(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Tracef(msg, args...)
}

// Trace will log a trace info.
func Trace(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Trace(msg, args...)
}

// Debugf will log a debug info with formatting.
func Debugf(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Debugf(msg, args...)
}

// Debug will log a debug info.
func Debug(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Debug(msg, args...)
}

// Infof will log an info with formatting.
func Infof(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Infof(msg, args...)
}

// Info will log an info.
func Info(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Info(msg, args...)
}

// Warnf will log a warning info with formatting.
func Warnf(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Warnf(msg, args...)
}

// Warn will log an warning info.
func Warn(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Warn(msg, args...)
}

// WarnEf will log a warning for an error `err` with formatting.
func WarnEf(err error, msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.WarnEf(err, msg, args...)
}

// WarnE will log a warning for an error `err`.
func WarnE(err error, msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.WarnE(err, msg, args...)
}

// Errorf will log an error with formatting.
func Errorf(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Errorf(msg, args...)
}

// Error will log an error.
func Error(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Error(msg, args...)
}

// ErrorEf will log an error for `err` with formatting.
func ErrorEf(err error, msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.ErrorEf(err, msg, args...)
}

// ErrorE will log an error for `err`.
func ErrorE(err error, msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.ErrorE(err, msg, args...)
}

// Panicf will log and panic with formatting.
func Panicf(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Panicf(msg, args...)
}

// Panic will log and panic.
func Panic(msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.Panic(msg, args...)
}

// PanicEf will log and panic with formatting.
func PanicEf(err error, msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.PanicEf(err, msg, args...)
}

// PanicE will log and panic if `err` is not `nil`.
func PanicE(err error, msg string, args ...any) {
	globalLogger.l.Helper()
	globalLogger.PanicE(err, msg, args...)
}
