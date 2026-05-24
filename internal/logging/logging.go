// Package logging sets up a process-wide logger that writes to both
// stdout and a log file under the OS temp directory.
//
// The built wails binary has no attached console, so plain fmt.Println
// output is invisible. This package routes everything through the
// standard `log` package so it lands in a file we can tail.
//
// Usage:
//
//	import "cria/internal/logging"
//
//	func main() {
//	    logging.Init()
//	    logging.Infof("starting up")
//	}
package logging

import (
	"fmt"
	"io"
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

// Init opens the log file (creating it if missing, appending if not)
// and reroutes the standard logger so log.Printf calls go to both
// stdout and the file. Safe to call multiple times; only the first
// call has effect.
func Init() {
	initOnce.Do(func() {
		logPath = filepath.Join(os.TempDir(), logFileName)
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// Couldn't open the file — keep stdout-only and warn.
			log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
			log.Printf("logging: failed to open %s: %v", logPath, err)
			return
		}
		logFile = f

		log.SetOutput(io.MultiWriter(os.Stdout, f))
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
		log.Printf("=== cria log opened at %s ===", logPath)
	})
}

// Close flushes and closes the log file. Call from shutdown hooks.
func Close() {
	if logFile != nil {
		log.Printf("=== cria log closing ===")
		_ = logFile.Sync()
		_ = logFile.Close()
		logFile = nil
	}
}

// Path returns the absolute path of the log file. Useful for surfacing
// the location to the user.
func Path() string {
	return logPath
}

// Infof logs an informational message. Prefix makes lines easy to grep.
func Infof(format string, args ...interface{}) {
	log.Output(2, "[INFO] "+fmt.Sprintf(format, args...))
}

// Errorf logs an error message.
func Errorf(format string, args ...interface{}) {
	log.Output(2, "[ERROR] "+fmt.Sprintf(format, args...))
}

// Debugf logs a verbose debug line.
func Debugf(format string, args ...interface{}) {
	log.Output(2, "[DEBUG] "+fmt.Sprintf(format, args...))
}
