package logutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/johnnylee/goutil/fileutil"
)

func SetFilePath(pathElem ...string) (err error) {
	path := fileutil.ExpandPath(pathElem...)

	// Make sure the directory exists.
	logDir := filepath.Dir(path)
	if err = os.MkdirAll(logDir, 0777); err != nil {
		return err
	}

	// Open the log file.
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Set the output.
	log.SetOutput(f)

	return nil
}

type Logger struct {
	prefix string
}

func New(prefix string) Logger {
	log.SetFlags(log.Ldate | log.Ltime)
	return Logger{fmt.Sprintf("[%v]", prefix)}
}

func (l Logger) Msg(format string, v ...interface{}) {
	log.Printf(l.prefix+" "+format, v...)
}

func (l Logger) Err(err error, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if err != nil {
		log.Printf("%v ERROR Msg: %v\n  Err: %v", l.prefix, msg, err)
	} else {
		log.Printf("%v ERROR Msg: %v", l.prefix, msg)
	}
}
