package main

import (
	"context"
	"os"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/abeja-inc/platform-model-proxy/cmd"
	"github.com/abeja-inc/platform-model-proxy/config"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

func main() {
	procCtx := context.TODO()
	os.Exit(execute(procCtx))
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
