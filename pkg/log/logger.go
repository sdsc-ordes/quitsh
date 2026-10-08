package log

import (
	"fmt"
	"strings"

	chlog "charm.land/log/v2"
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

// Interface implementation guard.
var _ ILog = (*logger)(nil)

// Trace implements [ILog].
func (l logger) Trace(msg string, args ...any) {
	l.l.Helper()
	if l.l.GetLevel() <= TraceLevel {
		l.l.Debug(msg, args...)
	}
}

// Tracef implements [ILog].
func (l logger) Tracef(msg string, args ...any) {
	l.l.Helper()
	l.Trace(fmt.Sprintf(msg, args...))
}

// Debug implements [ILog].
func (l logger) Debug(msg string, args ...any) {
	l.l.Helper()
	l.l.Debug(msg, args...)
}

// Debugf implements [ILog].
func (l logger) Debugf(msg string, args ...any) {
	l.l.Helper()
	l.l.Debugf(msg, args...)
}

// Info implements [ILog].
func (l logger) Info(msg string, args ...any) {
	l.l.Helper()
	l.l.Info(msg, args...)
}

// Infof implements [ILog].
func (l logger) Infof(msg string, args ...any) {
	l.l.Helper()
	l.l.Infof(msg, args...)
}

// Warn implements [ILog].
func (l logger) Warn(msg string, args ...any) {
	l.l.Helper()
	l.l.Warn(msg, args...)
}

// Warnf implements [ILog].
func (l logger) Warnf(msg string, args ...any) {
	l.l.Helper()
	l.l.Warnf(msg, args...)
}

// WarnE implements [ILog].
func (l logger) WarnE(err error, msg string, args ...any) {
	if err == nil {
		return
	}
	l.l.Helper()
	//FIXME: I want here to print a nice string but
	//       until resolved: https://github.com/charmbracelet/log/issues/187
	l.Warn("Error Summary:", "error", strings.ReplaceAll(err.Error(), "\t", "  "))
	l.Warn(msg, args...)
}
func (l logger) WarnEf(err error, msg string, args ...any) {
	l.l.Helper()
	l.WarnE(err, fmt.Sprintf(msg, args...))
}

// Error implements [ILog].
func (l logger) Error(msg string, args ...any) {
	l.l.Helper()
	l.l.Error(msg, args...)
}

// Errorf implements [ILog].
func (l logger) Errorf(msg string, args ...any) {
	l.l.Helper()
	l.l.Errorf(msg, args...)
}

// ErrorE implements [ILog].
func (l logger) ErrorE(err error, msg string, args ...any) {
	if err == nil {
		return
	}
	globalLogger.l.Helper()
	//FIXME: I want here to print a nice string but
	//       until resolved: https://github.com/charmbracelet/log/issues/187
	l.Error("Error Summary:", "error", strings.ReplaceAll(err.Error(), "\t", "  "))
	l.Error(msg, args...)
}

// ErrorEf implements [ILog].
func (l logger) ErrorEf(err error, msg string, args ...any) {
	l.l.Helper()
	l.ErrorE(err, fmt.Sprintf(msg, args...))
}

// Panic implements [ILog].
func (l logger) Panic(msg string, args ...any) {
	l.l.Helper()
	l.Error(msg, args...)
	panic(msg)
}

// Panicf implements [ILog].
func (l logger) Panicf(msg string, args ...any) {
	l.l.Helper()
	l.Panic(fmt.Sprintf(msg, args...))
}

// PanicE implements [ILog].
func (l logger) PanicE(err error, msg string, args ...any) {
	l.l.Helper()
	l.ErrorE(err, msg, args...)
	if err != nil {
		panic(err)
	}
}

// PanicEf implements [ILog].
func (l logger) PanicEf(err error, msg string, args ...any) {
	l.l.Helper()
	l.PanicE(err, fmt.Sprintf(msg, args...))
}
