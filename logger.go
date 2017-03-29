package logger

const (
	LVL_DEBUG = 60
	LVL_INFO  = 70
	LVL_WARN  = 80
	LVL_FATAL = 90
	LVL_PANIC = 100

	LOGGER_FILE    = "file"
	LOGGER_CONSOLE = "console"
)

type ILogger interface {
	Debug(lines ...interface{})
	Debugf(format string, v ...interface{})

	Info(lines ...interface{})
	Infof(format string, v ...interface{})

	Warn(lines ...interface{})
	Warnf(format string, v ...interface{})

	Fatal(lines ...interface{})
	Fatalf(format string, v ...interface{})

	Panic(lines ...interface{})
	Panicf(format string, v ...interface{})

	SetMaxLevel(level int)
	Close() error
}

type Logger struct {
	prefix string
	logWriter LogWriter
}

func NewLogger(prefix string, lw LogWriter) ILogger {
	return &Logger {
		prefix: prefix,
		logWriter: lw,
	}
}

func (l *Logger) Debug(v ...interface{}) {
	l.logWriter.Write(LVL_DEBUG, l.prefix, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logWriter.Writef(LVL_DEBUG, l.prefix, format, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.logWriter.Write(LVL_INFO, l.prefix, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.logWriter.Writef(LVL_INFO, l.prefix, format, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.logWriter.Write(LVL_WARN, l.prefix, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logWriter.Writef(LVL_WARN, l.prefix, format, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.logWriter.Write(LVL_FATAL, l.prefix, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logWriter.Writef(LVL_FATAL, l.prefix, format, v...)
}

func (l *Logger) Panic(v ...interface{}) {
	l.logWriter.Write(LVL_PANIC, l.prefix, v...)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	l.logWriter.Writef(LVL_PANIC, l.prefix, format, v...)
}

func (l *Logger) SetMaxLevel(level int) {
	l.logWriter.SetMaxLevel(level)
}

func (l *Logger) Close() error {
	return l.logWriter.Close()
}