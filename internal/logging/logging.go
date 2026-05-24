package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const logFileName = "cria.log"

var (
	initOnce sync.Once
	logFile  *os.File
	logPath  string
)

type safeWriter struct {
	file *os.File
}

func (w *safeWriter) Write(p []byte) (n int, err error) {
	if w.file != nil {
		_, _ = w.file.Write(p)
	}
	_, _ = os.Stdout.Write(p)

	return len(p), nil
}

func Init() {
	initOnce.Do(func() {
		logPath = filepath.Join(os.TempDir(), logFileName)
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
			log.Printf("logging: failed to open %s: %v", logPath, err)
			return
		}
		logFile = f

		log.SetOutput(&safeWriter{file: f})
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
		log.Printf("=== cria log opened at %s ===", logPath)
	})
}

func Close() {
	if logFile != nil {
		log.Printf("=== cria log closing ===")
		_ = logFile.Sync()
		_ = logFile.Close()
		logFile = nil
	}
}

func Path() string {
	return logPath
}

func Infof(format string, args ...interface{}) {
	log.Output(2, "[INFO] "+fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	log.Output(2, "[ERROR] "+fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}) {
	log.Output(2, "[DEBUG] "+fmt.Sprintf(format, args...))
}

func Userf(format string, args ...interface{}) {
	log.Output(2, "[USER] "+fmt.Sprintf(format, args...))
}

func Statef(format string, args ...interface{}) {
	log.Output(2, "[STATE] "+fmt.Sprintf(format, args...))
}
