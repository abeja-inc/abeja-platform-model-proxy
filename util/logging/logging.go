package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/evalphobia/logrus_sentry"
	log "github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"

	pathutil "github.com/abeja-inc/platform-model-proxy/util/path"
)

var ErrorFormat string

const logFileName = ".abeja_train.log"

func init() {

	if _, ok := os.LookupEnv("LOG_DETAIL"); ok {
		ErrorFormat = "%+v"
	} else {
		ErrorFormat = "%v"
	}

	jsonFormatter := &JSONFormatter{
		FieldMap: FieldMap{
			log.FieldKeyTime:  "timestamp",
			log.FieldKeyLevel: "log_level",
			log.FieldKeyMsg:   "message",
		},
		TimestampFormat: "2006-01-02T15:04:05.000-07:00",
	}
	log.SetOutput(ioutil.Discard)

	var logLevel log.Level
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		if level, err := log.ParseLevel(v); err != nil {
			logLevel = log.InfoLevel
		} else {
			logLevel = level
		}
	} else {
		logLevel = log.InfoLevel
	}

	stdoutHook := NewLogHook4Stdout(jsonFormatter, logLevel)
	log.AddHook(stdoutHook)

	if _, ok := os.LookupEnv("ABEJA_EXPORT_TRAIN_LOG"); ok {
		fileHook, err := getFileHook(logLevel)
		if err != nil {
			fmt.Fprintf(os.Stdout, "failed to open log file: %v", err)
		} else {
			log.AddHook(fileHook)
		}
	}

	sentryDsn := os.Getenv("SENTRY_DSN")
	if sentryDsn != "" {
		levels := []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
		}
		hook, err := logrus_sentry.NewAsyncSentryHook(sentryDsn, levels)
		if err != nil {
			log.WithFields(log.Fields{"err": err.Error()}).Error("failed to initialize sentry.")
			return
		}
		hook.StacktraceConfiguration.Enable = true
		hook.Timeout = 30 * time.Second
		log.AddHook(hook)
	}
}

func getFileHook(logLevel log.Level) (*LogHook, error) {
	trainingResultDir := os.Getenv("ABEJA_TRAINING_RESULT_DIR")
	userRoot := os.Getenv("ABEJA_USER_MODEL_ROOT")
	logDir, err := pathutil.GetTrainingResultDir(trainingResultDir, userRoot)
	if err != nil {
		return nil, errors.Errorf("failed to get logDir: %w", err)
	}
	logfile := filepath.Join(logDir, logFileName)
	fileFormatter := &SimpleFormatter{
		FieldMap: FieldMap{
			log.FieldKeyTime: "timestamp",
			log.FieldKeyMsg:  "message",
		},
		TimestampFormat: "2006-01-02T15:04:05.000-07:00",
	}
	return NewLogHook4File(
		logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666, fileFormatter, logLevel)
}
