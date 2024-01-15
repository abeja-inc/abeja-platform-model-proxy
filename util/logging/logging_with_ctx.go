package logging

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/abeja-inc/abeja-platform-model-proxy/version"
)

const KeyRequestID = "request_id"
const KeyRequesterID = "requester_id"

var (
	serviceID      string
	runID          string
	jobID          string
	deploymentID   string
	organizationID string
)

func init() {

	if v, ok := os.LookupEnv("ABEJA_SERVICE_ID"); ok {
		serviceID = v
	}
	if v, ok := os.LookupEnv("ABEJA_TRAINING_JOB_ID"); ok {
		jobID = v
	}
	if v, ok := os.LookupEnv("ABEJA_RUN_ID"); ok {
		runID = v
	}
	if v, ok := os.LookupEnv("ABEJA_ORGANIZATION_ID"); ok {
		organizationID = v
	}
	if v, ok := os.LookupEnv("ABEJA_DEPLOYMENT_ID"); ok {
		deploymentID = v
	}
}

func Debug(ctx context.Context, v ...interface{}) {
	Log(ctx, log.DebugLevel, v...)
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	Logf(ctx, log.DebugLevel, format, v...)
}

func Info(ctx context.Context, v ...interface{}) {
	Log(ctx, log.InfoLevel, v...)
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	Logf(ctx, log.InfoLevel, format, v...)
}

func Warning(ctx context.Context, v ...interface{}) {
	Log(ctx, log.WarnLevel, v...)
}

func Warningf(ctx context.Context, format string, v ...interface{}) {
	Logf(ctx, log.WarnLevel, format, v...)
}

func Error(ctx context.Context, v ...interface{}) {
	Log(ctx, log.ErrorLevel, v...)
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	Logf(ctx, log.ErrorLevel, format, v...)
}

func Fatal(ctx context.Context, v ...interface{}) {
	Log(ctx, log.FatalLevel, v...)
}

func Fatalf(ctx context.Context, format string, v ...interface{}) {
	Logf(ctx, log.FatalLevel, format, v...)
}

func Log(ctx context.Context, level log.Level, v ...interface{}) {
	fields := log.Fields{
		"version": version.Version,
	}
	if ctx != nil {
		if v := ctx.Value(KeyRequestID); v != nil && v != "" {
			fields[KeyRequestID] = v
		}
		if v := ctx.Value(KeyRequesterID); v != nil && v != "" {
			fields[KeyRequesterID] = v
		}
	}
	if serviceID != "" {
		fields["service_id"] = serviceID
	}
	if runID != "" {
		fields["run_id"] = runID
	}
	if jobID != "" {
		fields["training_job_id"] = jobID
	}
	if organizationID != "" {
		fields["organization_id"] = organizationID
	}
	if deploymentID != "" {
		fields["deployment_id"] = deploymentID
	}
	log.WithFields(fields).Log(level, v...)
}

func Logf(ctx context.Context, level log.Level, format string, v ...interface{}) {
	Log(ctx, level, fmt.Sprintf(format, v...))
}

func AccessLog(ctx context.Context, status int, start time.Time, r *http.Request) {
	end := time.Now()
	delta := end.Sub(start)
	fields := logrus.Fields{}
	if ctx != nil {
		if v := ctx.Value(KeyRequestID); v != nil && v != "" {
			fields[KeyRequestID] = v
		}
		if v := ctx.Value(KeyRequesterID); v != nil && v != "" {
			fields[KeyRequesterID] = v
		}
	}
	if serviceID != "" {
		fields["service_id"] = serviceID
	}
	if organizationID != "" {
		fields["organization_id"] = organizationID
	}
	if deploymentID != "" {
		fields["deployment_id"] = deploymentID
	}
	fields["log_type"] = "access_log"
	fields["elapsed_microsecs"] = (delta / 1000)
	fields["http_method"] = r.Method
	fields["http_status"] = status
	fields["request_timestamp"] = start.Format("2006-01-02T15:04:05.000000-07:00")
	fields["response_timestamp"] = end.Format("2006-01-02T15:04:05.000000-07:00")
	log.WithFields(fields).Log(log.InfoLevel)
}
