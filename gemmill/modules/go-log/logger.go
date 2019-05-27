package log

// Logger is a common logger used univeral
type Logger interface {
	Debug(msg string, keyvals ...interface{}) error
	Info(msg string, keyvals ...interface{}) error
	Error(msg string, keyvals ...interface{}) error

	With(keyvals ...interface{}) Logger
}
