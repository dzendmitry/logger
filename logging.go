package logger

import (
	"os"
	"sync"
)

var lMap map[string]LogWriter = make(map[string]LogWriter)
var mutex sync.Mutex

func InitFileLogger(prefix, path string) ILogger {
	mutex.Lock()
	lw, ok := lMap[LOGGER_FILE]
	if !ok {
		if path == "" {
			path = DEFAULT_FILE_LOG_WRITER_PATH
		}
		lMap[LOGGER_FILE] = NewFileLogWriter(path, "")
		lw = lMap[LOGGER_FILE]
	}
	mutex.Unlock()
	return NewLogger(prefix, lw)
}

func InitConsoleLogger(prefix string) ILogger {
	mutex.Lock()
	lw, ok := lMap[LOGGER_CONSOLE]
	if !ok {
		lMap[LOGGER_CONSOLE] = NewFileStreamLogWriter(os.Stdout)
		lw = lMap[LOGGER_CONSOLE]
	}
	mutex.Unlock()
	return NewLogger(prefix, lw)
}