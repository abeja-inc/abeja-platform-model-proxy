package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/convert"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/subprocess"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

func getHealthCheckHandleFunc(runtime *subprocess.Runtime) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		w.Header().Set("Content-Type", "application/json")
		if runtime.IsReady() {
			w.WriteHeader(http.StatusOK)
			status := []byte("{\"status\":\"ok\"}")
			if _, err := w.Write(status); err != nil {
				log.Warningf(ctx, "Error when writing response body: "+log.ErrorFormat, err)
			}
		} else {
			var status string
			if runtime.Status == subprocess.RuntimeStatusExitedWithSuccess {
				w.WriteHeader(http.StatusNotFound)
				status = "{\"status\":\"service not found\"}"
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				status = "{\"status\":\"service unavailable\"}"
			}
			if _, err := w.Write([]byte(status)); err != nil {
				log.Warningf(ctx, "Error when writing response body: "+log.ErrorFormat, err)
			}
		}
	}
}

func getRequestHandleFunc(
	runtime *subprocess.Runtime,
	request chan entity.ContentList,
	response chan entity.Response,
	conf *config.Configuration) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		accessLog := AccessLog{
			start: time.Now(),
		}
		defer func() {
			accessLog.log(ctx, r)
		}()
		if v := r.Header.Get("x-abeja-request-id"); v != "" {
			ctx = context.WithValue(ctx, log.KeyRequestID, v) //nolint // SA1029: should not use built-in type string as key for value; define your own type to avoid collisions
		}
		if v := r.Header.Get("x-abeja-requester-id"); v != "" {
			ctx = context.WithValue(ctx, log.KeyRequesterID, v) //nolint // SA1029: should not use built-in type string as key for value; define your own type to avoid collisions
		}

		if !runtime.IsReady() {
			// not ready
			outputErrorResponse(ctx, w, http.StatusServiceUnavailable, "service unavailable")
			accessLog.status = http.StatusServiceUnavailable
			return
		}

		cl, err := convert.ToContents(ctx, r, conf)
		if err != nil {
			// failed to parse request
			var statusCode = http.StatusServiceUnavailable
			if convertError, ok := err.(*convert.ConverterError); ok {
				statusCode = convertError.StatusCode
			}
			outputErrorResponse(ctx, w, statusCode, err.Error())
			accessLog.status = statusCode
			return
		}

		asyncRequestID := r.Header.Get("x-abeja-arms-async-request-id")
		if asyncRequestID != "" {
			// async request
			asyncToken := r.Header.Get("x-abeja-arms-async-request-token")
			cl.AsyncRequestID = asyncRequestID
			cl.AsyncARMSToken = asyncToken
			request <- *cl

			w.Header().Set(convert.KeyContentType, "application/json")
			// SAMPv2 limits the number of concurrent requests by LimitListener,
			// but it accepts requests from multiple clients at the same time.
			// Keepalive is disabled because it can't process requests
			// until the previous connection is closed, which causes a wait time.
			w.Header().Set(convert.KeyConnection, "close")
			w.WriteHeader(http.StatusAccepted)
			if _, err := w.Write([]byte("")); err != nil {
				log.Warningf(ctx, "Error when writing empty response body: "+log.ErrorFormat, err)
			}

			accessLog.status = http.StatusAccepted
			return
		}

		request <- *cl

		res := <-response
		status, headers, body, err := convert.FromResponse(ctx, res)
		if err != nil {
			var statusCode = http.StatusServiceUnavailable
			if convertError, ok := err.(*convert.ConverterError); ok {
				statusCode = convertError.StatusCode
			}
			outputErrorResponse(ctx, w, statusCode, err.Error())
			accessLog.status = statusCode
			return
		}

		for key, value := range headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(status)
		if _, err := io.Copy(w, body); err != nil {
			log.Warningf(ctx, "Error when writing response body: "+log.ErrorFormat, err)
		}
		// Even if an error occurs during the transmission of response,
		// record the response code to be returned
		accessLog.status = status

		// TODO record request/response
		deleteTempFiles(ctx, cl, body)
	}
}

func outputErrorResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	status := []byte(fmt.Sprintf("{\"status\":\"%s\"}", message))
	if _, err := w.Write(status); err != nil {
		log.Warningf(ctx, "Error when writing response body: "+log.ErrorFormat, err)
	}
}

type AccessLog struct {
	start  time.Time
	status int
}

func (a AccessLog) log(ctx context.Context, r *http.Request) {
	log.AccessLog(ctx, a.status, a.start, r)
}
