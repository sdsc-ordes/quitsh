package log

type logger struct{}

// NewLoggerST creates a non-threadsafe logger.
func NewLoggerST() ILog {
	return &logger{}
}

func (l logger) Trace(msg string, args ...any) {
	Trace(msg, args...)
}
func (l logger) Debug(msg string, args ...any) {
	Debug(msg, args...)
}
func (l logger) Info(msg string, args ...any) {
	Info(msg, args...)
}
func (l logger) Warn(msg string, args ...any) {
	Warn(msg, args...)
}
func (l logger) WarnE(err error, msg string, args ...any) {
	WarnE(err, msg, args...)
}
func (l logger) Error(msg string, args ...any) {
	Error(msg, args...)
}
func (l logger) ErrorE(err error, msg string, args ...any) {
	ErrorE(err, msg, args...)
}
func (l logger) Panic(msg string, args ...any) {
	Panic(msg, args...)
}
func (l logger) PanicE(err error, msg string, args ...any) {
	PanicE(err, msg, args...)
}
