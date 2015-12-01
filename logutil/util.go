package logutil

import (
	"fmt"
	"github.com/johnnylee/goutil/fileutil"
	"log"
	"os"
	"path/filepath"
)

type UtilLogger struct {
	msgLog *log.Logger
	errLog *log.Logger
}

func New(prefix string, pathElem ...string) (UtilLogger, error) {
	l := UtilLogger{
		log.New(os.Stderr, "", log.Ldate|log.Ltime),
		log.New(os.Stderr, "", log.Ldate|log.Ltime),
	}

	l.SetPrefix(prefix)
	err := l.UseFiles(pathElem...)
	return l, err
}

func (l UtilLogger) SetPrefix(prefix string) {
	l.msgLog.SetPrefix(prefix)
	l.errLog.SetPrefix(prefix)
}

func (l UtilLogger) UseFiles(pathElem ...string) (err error) {
	pathPrefix := fileutil.ExpandPath(pathElem...)

	// Make sure the directory exists.
	logDir := filepath.Dir(pathPrefix)
	if err = os.MkdirAll(logDir, 0777); err != nil {
		return err
	}

	// Output file paths.
	msgPath := pathPrefix + ".err.log"
	errPath := pathPrefix + ".msg.log"

	// Open the files.
	fMsg, err := os.OpenFile(msgPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fErr, err := os.OpenFile(errPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Set output.
	l.msgLog.SetOutput(fMsg)
	l.errLog.SetOutput(fErr)
	return nil
}

func (l UtilLogger) LogMsg(format string, v ...interface{}) {
	l.msgLog.Printf(format, v...)
}

func (l UtilLogger) LogErr(err error, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.errLog.Printf("ERROR: %v:\n    %v\n", msg, err)
}
