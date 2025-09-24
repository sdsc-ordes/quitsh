package log

import (
	"fmt"

	chlog "github.com/charmbracelet/log"
)

type logger struct {
	l *chlog.Logger
}

// NewLogger creates a threadsafe logger with structured key/value pairs added.
func NewLogger(prefix string) ILog {
	return &logger{
		l: globalLogger.l.WithPrefix(prefix),
	}
}

func (l logger) Trace(msg string, args ...any) {
	l.l.Helper()
	if l.l.GetLevel() <= TraceLevel {
		l.l.Debug(msg, args...)
	}
}
func (l logger) Tracef(msg string, args ...any) {
	l.l.Helper()
	l.Trace(fmt.Sprintf(msg, args...))
}

func (l logger) Debug(msg string, args ...any) {
	l.l.Helper()
	l.l.Debug(msg, args...)
}
func (l logger) Debugf(msg string, args ...any) {
	l.l.Helper()
	l.l.Debugf(msg, args...)
}

func (l logger) Info(msg string, args ...any) {
	l.l.Helper()
	l.l.Info(msg, args...)
}
func (l logger) Infof(msg string, args ...any) {
	l.l.Helper()
	l.l.Infof(msg, args...)
}

func (l logger) Warn(msg string, args ...any) {
	l.l.Helper()
	l.l.Warn(msg, args...)
}
func (l logger) Warnf(msg string, args ...any) {
	l.l.Helper()
	l.l.Warnf(msg, args...)
}
func (l logger) WarnE(err error, msg string, args ...any) {
	l.l.Helper()
	a := make([]any, 0, 2+len(args)) //nolint: mnd
	a = append(a, "error", err.Error())
	a = append(a, args...)
	l.Warn(msg, a...)
}
func (l logger) WarnEf(err error, msg string, args ...any) {
	l.l.Helper()
	l.WarnE(err, fmt.Sprintf(msg, args...))
}

func (l logger) Error(msg string, args ...any) {
	l.l.Helper()
	l.l.Error(msg, args...)
}
func (l logger) Errorf(msg string, args ...any) {
	l.l.Helper()
	l.l.Errorf(msg, args...)
}
func (l logger) ErrorE(err error, msg string, args ...any) {
	globalLogger.l.Helper()
	a := make([]any, 0, 2+len(args)) //nolint: mnd
	a = append(a, args...)
	a = append(a, "error", err.Error())
	l.Error(msg, a...)
}
func (l logger) ErrorEf(err error, msg string, args ...any) {
	l.l.Helper()
	l.ErrorE(err, fmt.Sprintf(msg, args...))
}

func (l logger) Panic(msg string, args ...any) {
	l.l.Helper()
	l.Error(msg, args...)
	panic(msg)
}
func (l logger) Panicf(msg string, args ...any) {
	l.l.Helper()
	Panic(fmt.Sprintf(msg, args...))
}
func (l logger) PanicE(err error, msg string, args ...any) {
	if err != nil {
		l.l.Helper()
		l.ErrorE(err, msg, args...)
		panic(err)
	}
}
func (l logger) PanicEf(err error, msg string, args ...any) {
	l.l.Helper()
	l.PanicE(err, fmt.Sprintf(msg, args...))
}
