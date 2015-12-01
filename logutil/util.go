package logutil

import (
	"fmt"
	"github.com/johnnylee/goutil/fileutil"
	"log"
	"os"
	"path/filepath"
)

type UtilLogger struct {
	logger *log.Logger
}

func NewLogger(prefix string, pathElem ...string) (UtilLogger, error) {
	logger := UtilLogger{}

	// Expand the path.
	path, err := fileutil.ExpandPath(pathElem...)
	if err != nil {
		return logger, err
	}

	// Make sure the directory exists.
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0777); err != nil {
		return logger, err
	}

	// Open the file.
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return logger, err
	}

	// Create the logger.
	logger.logger = log.New(f, "["+prefix+"] ", log.Ldate|log.Ltime)
	return logger, nil
}

func (l UtilLogger) Msg(format string, v ...interface{}) {
	l.logger.Printf(format+"\n", v...)
}

func (l UtilLogger) Err(err error, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("ERROR: %v:\n    %v\n", msg, err)
}
