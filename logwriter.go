package logger

import (
	"sync"
	"os"
	"fmt"
	"time"
	"bytes"
)

const (
	LOG_DATETIME_FORMAT = "15:04:05.999 2006-01-02"
)

type LogWriter interface {
	Write(level int, prefix string, v ...interface{})
	Writef(level int, prefix, format string, v ...interface{})
	SetMaxLevel(level int)
	Close() error
}

type BaseLogWriter struct {
	mutex sync.Mutex
	maxLevel int
}

func (flw *BaseLogWriter) SetMaxLevel(level int) {
	flw.mutex.Lock()
	defer flw.mutex.Unlock()
	flw.maxLevel = level
}

const (
	DEFAULT_FILE_LOG_WRITER_PATH = "."
	DEFAULT_FILE_SIZE_LIMIT = 100 * 1024 * 1024
	DEFAULT_FILE_BUF_SIZE = 4096
)

var (
	EOL = []byte("\n")
)

type FileLogWriter struct {
	BaseLogWriter
	path string
	prefix string
	file *os.File
	opened bool
	lastTime time.Time
	curSize int
	sizeLimit int
	rbw *RingBufWriter
	buf []byte
}

func NewFileLogWriter(path, prefix string) LogWriter {
	return &FileLogWriter{
		path: path,
		prefix: prefix,
		sizeLimit: DEFAULT_FILE_SIZE_LIMIT,
		buf: make([]byte, 0, DEFAULT_FILE_BUF_SIZE),
	}
}

func (flw *FileLogWriter) Write(level int, prefix string, v ...interface{}) {
	flw.mutex.Lock()
	defer flw.mutex.Unlock()
	if level < flw.maxLevel {
		return
	}
	t := time.Now()
	dayDiffers := t.Day() != flw.lastTime.Day() || t.Month() != flw.lastTime.Month() || t.Year() != flw.lastTime.Year()
	if !flw.opened || flw.curSize > flw.sizeLimit || dayDiffers {
		if !flw.reopenFile(t) {
			return
		}
	}
	b := bytes.NewBuffer(flw.buf)
	fmt.Fprintf(b, "%s %1d %s ", t.Local().Format(LOG_DATETIME_FORMAT), level, prefix)
	prefixBytes := b.Bytes()
	for i := range v {
		flw.curSize += flw.rbw.WriteLine(prefixBytes, []byte(v[i].(string)), EOL)
	}
}

func (flw *FileLogWriter) Writef(level int, prefix, format string, v ...interface{}) {
	flw.Write(level, prefix, fmt.Sprintf(format, v...))
}

func (flw *FileLogWriter) reopenFile(t time.Time) bool {
	flw.opened = false
	var err error
	if flw.file != nil {
		flw.unlockedClose()
	}
	var fileStamp string
	if t.Hour() == 0 && t.Minute() == 0 {
		fileStamp = fmt.Sprintf("%4d-%02d-%02d", t.Year(), t.Month(), t.Day())
	} else {
		fileStamp = fmt.Sprintf("%4d-%02d-%02d_%02d-%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	}
	flw.file, err = os.OpenFile(fmt.Sprintf("%s%c%s%s.log", flw.path, os.PathSeparator, flw.prefix, fileStamp),
		os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		return false
	}
	flw.rbw = NewRingBufWriter(flw.file)
	flw.lastTime = t
	flw.opened = true
	flw.curSize = 0
	return true
}

func (flw *FileLogWriter) Close() error {
	flw.mutex.Lock()
	defer flw.mutex.Unlock()
	return flw.unlockedClose()
}

func (flw *FileLogWriter) unlockedClose() error {
	if flw.rbw != nil {
		flw.rbw.Close()
		flw.rbw = nil
	}
	if flw.file != nil {
		err := flw.file.Close()
		flw.file = nil
		return err
	}
	return nil
}

type FileStreamLogWriter struct {
	BaseLogWriter
	file *os.File
}

func NewFileStreamLogWriter(fstream *os.File) LogWriter {
	return &FileStreamLogWriter{
		file: fstream,
	}
}

func (clw *FileStreamLogWriter) Write(level int, prefix string, v ...interface{}) {
	clw.mutex.Lock()
	defer clw.mutex.Unlock()
	if level < clw.maxLevel {
		return
	}
	fmt.Fprintf(clw.file, "%s %s ", time.Now().Local().Format(LOG_DATETIME_FORMAT), prefix)
	fmt.Fprintln(clw.file, v...)
}

func (clw *FileStreamLogWriter) Writef(level int, prefix, format string, v ...interface{}) {
	clw.Write(level, prefix, fmt.Sprintf(format, v...))
}

func (clw *FileStreamLogWriter) Close() error {
	clw.mutex.Lock()
	defer clw.mutex.Unlock()
	return clw.file.Close()
}