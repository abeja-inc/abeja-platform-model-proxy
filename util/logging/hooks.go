package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type LogHook struct {
	formatter logrus.Formatter
	writer    io.Writer
	levels    []logrus.Level
}

func (hook LogHook) Fire(entry *logrus.Entry) error {
	formatted, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.writer.Write(formatted)
	return err
}

func (hook LogHook) Levels() []logrus.Level {
	return hook.levels
}

func NewLogHook4Stdout(formatter logrus.Formatter, level logrus.Level) *LogHook {
	return &LogHook{
		formatter: formatter,
		writer:    os.Stdout,
		levels:    getAllowedLevels(level),
	}
}

func NewLogHook4File(path string, flag int, mode os.FileMode, formatter logrus.Formatter, level logrus.Level) (*LogHook, error) {
	logFile, err := os.OpenFile(path, flag, mode)
	if err != nil {
		return nil, err
	}
	return &LogHook{
		formatter: formatter,
		writer:    logFile,
		levels:    getAllowedLevels(level),
	}, nil
}

func getAllowedLevels(level logrus.Level) []logrus.Level {
	return logrus.AllLevels[0 : level+1]
}
