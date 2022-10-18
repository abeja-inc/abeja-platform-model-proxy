package config

import (
	"net"
	"os"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const envKeyDatadogEnv = "DATADOG_ENV"
const envKeyDatadogServiceName = "DATADOG_SERVICE_NAME"
const envKeyDatadogTraceAgentHostname = "DATADOG_TRACE_AGENT_HOSTNAME"
const envKeyDatadogTraceAgentPort = "DATADOG_TRACE_AGENT_PORT"

const defaultDatadogTraceAgentPort = "8126"

func GetTraceOptions() []tracer.StartOption {
	var options []tracer.StartOption
	ddHost, ok := os.LookupEnv(envKeyDatadogTraceAgentHostname)
	if !ok {
		return options
	}

	var ddPort string
	port, ok := os.LookupEnv(envKeyDatadogTraceAgentPort)
	if ok {
		ddPort = port
	} else {
		ddPort = defaultDatadogTraceAgentPort
	}
	addr := net.JoinHostPort(
		ddHost,
		ddPort,
	)
	options = append(options, tracer.WithAgentAddr(addr))
	options = append(options, tracer.WithAnalytics(true))

	if serviceName, ok := os.LookupEnv(envKeyDatadogServiceName); ok {
		options = append(options, tracer.WithServiceName(serviceName))
	}

	if datadogEnv, ok := os.LookupEnv(envKeyDatadogEnv); ok {
		options = append(options, tracer.WithGlobalTag("env", datadogEnv))
	}
	return options
}

func GetHTTPTraceOptions() []httptrace.MuxOption {
	var muxOptions []httptrace.MuxOption
	if serviceName, ok := os.LookupEnv("DATADOG_SERVICE_NAME"); ok {
		muxOptions = append(muxOptions, httptrace.WithServiceName(serviceName))
	}
	return muxOptions
}
