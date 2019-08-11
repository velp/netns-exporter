package log

// Debug logs a message at Debug level.
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Debugf logs a formatted message at Debug level.
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Info logs a message at Info level.
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof logs a formatted message at Info level.
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Warn logs a message at Warn level.
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf logs a formatted message at Warn level.
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs a message at Error level.
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf logs a formatted message at Error level.
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a message at Fatal level.
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted message at Fatal level.
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}
