package logutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/johnnylee/goutil/fileutil"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetFilePath(pathElem ...string) (err error) {
	path := fileutil.ExpandPath(pathElem...)

	// Make sure the directory exists.
	logDir := filepath.Dir(path)
	if err = os.MkdirAll(logDir, 0777); err != nil {
		return err
	}

	// Set the output.
	log.SetOutput(&lumberjack.Logger{
		Filename: path,
		MaxSize:  100,
	})

	return nil
}

type Logger struct {
	prefix string
}

func New(prefix string) Logger {
	log.SetFlags(log.Ldate | log.Ltime)
	return Logger{fmt.Sprintf("[%v] ", prefix)}
}

func (l Logger) Msg(format string, v ...interface{}) {
	log.Printf(l.prefix+format+"\n", v...)
}

func (l Logger) Err(err error, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if err != nil {
		log.Printf(l.prefix+"ERROR: %v:\n%v\n", msg, err)
	} else {
		log.Printf(l.prefix+"ERROR: %v\n", msg)
	}
}
