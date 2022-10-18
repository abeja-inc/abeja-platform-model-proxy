package proxy

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/netutil"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/subprocess"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

// HTTPServer represents the wrapper of net/http/Server.
type HTTPServer struct {
	Server            *http.Server
	HealthCheckServer *http.Server
	req               chan entity.ContentList
}

func deleteTempFiles(ctx context.Context, cl *entity.ContentList, resBody *os.File) {
	if cl != nil {
		contents := cl.Contents
		for _, c := range contents {
			cleanutil.Remove(ctx, *c.Path)
		}
	}
	if resBody != nil {
		cleanutil.Close(ctx, resBody, resBody.Name())
		cleanutil.Remove(ctx, resBody.Name())
	}
}

// CreateHTTPServer return HTTPServer.
func CreateHTTPServer(
	runtime *subprocess.Runtime,
	request chan entity.ContentList,
	response chan entity.Response,
	conf *config.Configuration) (*HTTPServer, error) {

	muxOptions := config.GetHTTPTraceOptions()
	serviceHandler := httptrace.NewServeMux(muxOptions...)
	healthCheckHandler := httptrace.NewServeMux(muxOptions...)

	// add HandlerFunc for health-check
	healthCheckHandler.HandleFunc("/health_check", getHealthCheckHandleFunc(runtime))
	serviceHandler.HandleFunc("/health_check", getHealthCheckHandleFunc(runtime))
	// add HandlerFunc for user request
	serviceHandler.HandleFunc(
		"/",
		getRequestHandleFunc(runtime, request, response, conf))

	serviceServer := &http.Server{
		Addr:           conf.GetListenAddress(),
		Handler:        serviceHandler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	healthCheckServer := &http.Server{
		Addr:           conf.GetHealthCheckAddress(),
		Handler:        healthCheckHandler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	httpServer := &HTTPServer{
		Server:            serviceServer,
		HealthCheckServer: healthCheckServer,
		req:               request,
	}
	return httpServer, nil
}

// ListenAndServe start serving http-request/response.
// Because the DL framework(s) are often incompatible with multithreading,
// This server has only one thread for waiting request.
func (hs *HTTPServer) ListenAndServe(ctx context.Context, errOnBoot chan int) {
	go func() {
		log.Debugf(ctx, "start listen health check with address: %s.", hs.HealthCheckServer.Addr)
		if err := hs.HealthCheckServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Errorf(
					ctx,
					"error occurred when health check request listening: "+log.ErrorFormat,
					err)
				close(errOnBoot)
			}
			log.Debugf(ctx, "finish to listen health check: "+log.ErrorFormat, err)
		}
	}()

	// Since the DL framework often does not support multithreading,
	// limit the number of simultaneous connections
	log.Debugf(ctx, "start listen with address: %s.", hs.Server.Addr)
	listener, err := net.Listen("tcp", hs.Server.Addr)
	if err != nil {
		log.Errorf(ctx, "failed to listen tcp port: "+log.ErrorFormat, err)
		close(errOnBoot)
		return
	}
	limitedListener := netutil.LimitListener(listener, 1)

	if err := hs.Server.Serve(limitedListener); err != nil {
		if err != http.ErrServerClosed {
			close(errOnBoot)
			log.Errorf(ctx, "error occurred when service request listening: "+log.ErrorFormat, err)
		}
		log.Debugf(ctx, "finish to listen: "+log.ErrorFormat, err)
	}
}

// Shutdown does graceful-shutdown.
func (hs *HTTPServer) Shutdown(procCtx context.Context, timeout time.Duration) error {
	close(hs.req)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go func() {
		if err := hs.HealthCheckServer.Shutdown(ctx); err != nil {
			log.Warningf(procCtx, "healthcheck server shutdown error: "+log.ErrorFormat, err)
		}
	}()
	return hs.Server.Shutdown(ctx)
}
