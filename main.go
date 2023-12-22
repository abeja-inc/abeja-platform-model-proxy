package main

import (
	"context"
	"os"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/abeja-inc/platform-model-proxy/cmd"
	"github.com/abeja-inc/platform-model-proxy/config"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
	"github.com/abeja-inc/platform-model-proxy/version"
)

func main() {
	procCtx := context.TODO()
	log.Infof(procCtx, "abeja-runner[%s]: Hello!", version.Version)
	code := execute(procCtx)
	log.Infof(procCtx, "abeja-runner[%s]: Bye!", version.Version)
	os.Exit(code)
}

func execute(procCtx context.Context) int {
	if setupDD(procCtx) {
		defer tracer.Stop()
	}
	return cmd.Execute(procCtx)
}

func setupDD(procCtx context.Context) bool {
	options := config.GetTraceOptions()
	if len(options) == 0 {
		// datadog is not set because the datadog-agent host is not set
		log.Debug(procCtx, "datadog is not set up.")
		return false
	}
	tracer.Start(options...)
	log.Debug(procCtx, "datadog is set up.")
	return true
}
