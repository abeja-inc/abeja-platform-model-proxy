package service

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	errors "golang.org/x/xerrors"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/preprocess"
	"github.com/abeja-inc/platform-model-proxy/proxy"
	"github.com/abeja-inc/platform-model-proxy/subprocess"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

var (
	runtime    *subprocess.Runtime
	httpServer *proxy.HTTPServer
)

func shutdownHTTPServer(ctx context.Context) {
	if httpServer != nil {
		if err := httpServer.Shutdown(ctx, 28*time.Second); err != nil {
			log.Errorf(ctx, "failed to shutdown httpserver: "+log.ErrorFormat, err)
		}
	}
}

func shutdownRuntime(ctx context.Context) {
	if runtime != nil {
		runtime.Shutdown(ctx, 25*time.Second)
	}
}

func shutdownServices(ctx context.Context, skipRuntime bool) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		shutdownHTTPServer(ctx)
	}()

	if skipRuntime {
		wg.Done()
	} else {
		go func() {
			defer wg.Done()
			shutdownRuntime(ctx)
		}()
	}

	wg.Wait()
}

func shutdownOnError(ctx context.Context, errOnBoot chan int, err error) {
	defer close(errOnBoot)
	log.Errorf(ctx, "unexpected error occurred: "+log.ErrorFormat, err)
}

func download(ctx context.Context, conf *config.Configuration) error {
	preprocessor, err := preprocess.NewPreprocessor(ctx, conf)
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	if err = preprocessor.Prepare(ctx); err != nil {
		return errors.Errorf(": %w", err)
	}
	return nil
}

// handleSignal traps signal(SIGINT/SIGTERM) and does shutdown-graceful runtime/web-server.
func handleSignal(
	ctx context.Context,
	dataDir string,
	errOnBoot chan int,
	exitStatus chan int,
	errOnSub chan error,
	notifyFromMain chan int,
	notifyToMain chan int) {

	var status int // exit status

	log.Debug(ctx, "handleSignal: create channel for signal")
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	log.Debug(ctx, "handleSignal: channel for signal created")

	log.Debug(ctx, "waiting signal...")
	skipRuntime := false
	func() {
		for {
			select {
			case <-errOnBoot:
				log.Warning(ctx, "failed to Bootstrapping.")
				status = 1
				return
			case sig := <-gracefulStop:
				log.Infof(ctx, "signal[%s] received.", sig.String())
				close(errOnBoot)
				return
			case err, received := <-errOnSub:
				if received {
					log.Warning(ctx, "Error when waiting finish runtime:", err)
					status = 1
				}
				return
			default:
				if runtime != nil {
					if runtime.IsExited(ctx) {
						skipRuntime = true
						return
					}
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// wait finishing of subprocess & web-server
	shutdownServices(ctx, skipRuntime)
	close(notifyFromMain)
	<-notifyToMain

	cleanutil.RemoveAll(ctx, dataDir)

	log.Debug(ctx, "runtime finished")

	// exit with subprocess status
	if status == 0 && runtime != nil {
		if runtime.Status == subprocess.RuntimeStatusExitedWithSuccess {
			status = 0
		} else {
			status = 1
		}
	}

	exitStatus <- status
}

func run(ctx context.Context, conf *config.Configuration, execDownload bool) error {

	workingDir, err := conf.GetWorkingDir()
	if err != nil {
		log.Fatalf(ctx, "failed to get working direcoty path: "+log.ErrorFormat, err)
		return errors.Errorf(": %w", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		log.Fatalf(
			ctx,
			"failed to move working direcoty path: %s, error: "+log.ErrorFormat,
			workingDir, err)
		return errors.Errorf(": %w", err)
	}

	udsFilePath, err := cmdutil.MakeUDSFilePath()
	if err != nil {
		log.Fatalf(
			ctx,
			"failed to build path to socket file for communication to runtime: "+log.ErrorFormat,
			err)
		return errors.Errorf(": %w", err)
	}
	defer cleanutil.RemoveAll(ctx, filepath.Dir(udsFilePath))

	trainingResultDir, err := conf.GetTrainingResultDir()
	if err != nil {
		log.Fatalf(ctx, "failed to get path for training-result: "+log.ErrorFormat, err)
		return errors.Errorf(": %w", err)
	}
	runtime, err = subprocess.CreateServiceRuntime(conf, udsFilePath, trainingResultDir)
	if err != nil {
		log.Fatalf(ctx, "failed to CreateServiceRuntime: "+log.ErrorFormat, err)
		return errors.Errorf(": %w", err)
	}

	// trap signals
	errOnBoot := make(chan int)
	exitStatus := make(chan int)
	// trap error of Cmd.Wait
	errOnSub := make(chan error)
	// for async error
	notifyFromMain := make(chan int)
	notifyToMain := make(chan int)
	// defer close(errOnBoot) // <- close clearly in shutdown process
	defer close(exitStatus)
	log.Debug(ctx, "call handleSignal")
	go handleSignal(
		ctx, conf.RequestedDataDir, errOnBoot, exitStatus, errOnSub, notifyFromMain, notifyToMain)

	// prepare & start web server
	request := make(chan entity.ContentList, 10000)
	response := make(chan entity.Response)
	// defer close(request) // <- close clearly in shutdown process
	defer close(response)

	// Need to finish downloading model/codes before serving healthcheck endpoint
	if execDownload {
		err := download(ctx, conf)
		if err != nil {
			shutdownOnError(ctx, errOnBoot, err)
			return errors.Errorf(": %w", err)
		}
	}

	httpServer, err = proxy.CreateHTTPServer(runtime, request, response, conf)
	if err != nil {
		shutdownOnError(ctx, errOnBoot, err)
		return errors.Errorf(": %w", err)
	}
	go httpServer.ListenAndServe(ctx, errOnBoot)

	// subprocess logger
	scopeChan := make(chan context.Context)
	defer close(scopeChan)
	runtimeLogger := subprocess.NewRuntimeLogger(ctx, runtime.Cmd, scopeChan)

	// start runtime
	if err = runtime.Start(errOnSub); err != nil {
		shutdownOnError(ctx, errOnBoot, err)
		return errors.Errorf(": %w", err)
	}

	runtimeLogger.Run()
	defer runtimeLogger.Flush(3) // wait 3 seconds for flush all logs.

	if err = runtime.WaitUntilStarted(ctx, udsFilePath); err != nil {
		shutdownOnError(ctx, errOnBoot, err)
		return errors.Errorf(": %w", err)
	}

	// connect to runtime after runtime started.
	go proxy.TransportMessages(ctx, conf, udsFilePath, request, response, errOnBoot, notifyFromMain, notifyToMain, scopeChan, nil)

	handledStatus := <-exitStatus
	if handledStatus > 0 {
		return errors.New("failed to finalize")
	}
	return nil
}
