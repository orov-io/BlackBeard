package api

//Fields Type to pass when we want to call WithFields for structured logging
type Fields map[string]interface{}

//Logger is our contract for the logger
type Logger interface {
	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})

	Panicf(format string, args ...interface{})

	WithFields(keyValues Fields) Logger
}

type noLogger struct{}

func (l *noLogger) Debugf(format string, args ...interface{}) {
	return
}

func (l *noLogger) Infof(format string, args ...interface{}) {
	return
}

func (l *noLogger) Warnf(format string, args ...interface{}) {
	return
}

func (l *noLogger) Errorf(format string, args ...interface{}) {
	return
}

func (l *noLogger) Fatalf(format string, args ...interface{}) {
	return
}

func (l *noLogger) Panicf(format string, args ...interface{}) {
	return
}

func (l *noLogger) WithFields(keyValues Fields) Logger {
	return l
}
