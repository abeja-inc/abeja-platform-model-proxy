package batch

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	errors "golang.org/x/xerrors"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/preprocess"
	"github.com/abeja-inc/abeja-platform-model-proxy/proxy"
	"github.com/abeja-inc/abeja-platform-model-proxy/subprocess"
	cleanutil "github.com/abeja-inc/abeja-platform-model-proxy/util/clean"
	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
)

var runtime *subprocess.Runtime

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

func run(ctx context.Context, conf *config.Configuration) error {

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
	runtime, err = subprocess.CreateOneshotRuntime(conf, udsFilePath, trainingResultDir)
	if err != nil {
		log.Fatalf(ctx, "failed to CreateOneshotRuntime: "+log.ErrorFormat, err)
		return errors.Errorf(": %w", err)
	}

	// trap signals
	errOnBoot := make(chan int)
	// trap error of Cmd.Wait
	errOnSub := make(chan error)
	// for async error
	notifyFromMain := make(chan int)
	notifyToMain := make(chan int)

	// subprocess logger
	scopeChan := make(chan context.Context, 1)
	defer close(scopeChan)
	runtimeLogger := subprocess.NewRuntimeLogger(ctx, runtime.Cmd, scopeChan)

	if err := runtime.Start(errOnSub); err != nil {
		shutdownOnError(ctx, errOnBoot)
		return errors.Errorf(": %w", err)
	}

	runtimeLogger.Run()
	defer runtimeLogger.Flush(3) // wait 3 seconds for flush all logs.

	if err := runtime.WaitUntilStarted(ctx, udsFilePath); err != nil {
		shutdownOnError(ctx, errOnBoot)
		return errors.Errorf(": %w", err)
	}

	go proxy.TransportOneshotMessage(
		ctx, conf, udsFilePath, errOnBoot, notifyFromMain, notifyToMain, nil)

	exitStatus := handleSignal(
		ctx, conf.RequestedDataDir, errOnBoot, errOnSub, notifyFromMain, notifyToMain)
	if exitStatus > 0 {
		return errors.New("failed to batch-process")
	}
	return nil
}

func handleSignal(
	ctx context.Context,
	dataDir string,
	errOnBoot chan int,
	errOnSub chan error,
	notifyFromMain chan int,
	notifyToMain chan int) int {

	var status int // exit status

	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)

	log.Debug(ctx, "waiting signal...")
	skipRuntime := false
	func() {
		for {
			select {
			case <-errOnBoot:
				log.Warning(ctx, "failed to Bootstrapping.")
				status = 1
				return
			case sig := <-sigChan:
				log.Infof(ctx, "signal[%s] received.", sig.String())
				close(errOnBoot)
				return
			case err, received := <-errOnSub:
				if received {
					log.Warning(ctx, "Error when waiting finish runtime:", err)
					status = 1
				}
				return
			case val, received := <-notifyToMain:
				if received {
					status = val
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

	if !skipRuntime {
		shutdownRuntime(ctx)
	}
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
	return status
}

func shutdownOnError(ctx context.Context, errOnBoot chan int) {
	defer close(errOnBoot)
	shutdownRuntime(ctx)
}

func shutdownRuntime(ctx context.Context) {
	if runtime != nil {
		runtime.Shutdown(ctx, 25*time.Second)
	}
}
