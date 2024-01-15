package training

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rakyll/statik/fs"
	errors "golang.org/x/xerrors"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	_ "github.com/abeja-inc/abeja-platform-model-proxy/runtime"
	"github.com/abeja-inc/abeja-platform-model-proxy/subprocess"
	cleanutil "github.com/abeja-inc/abeja-platform-model-proxy/util/clean"
	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
)

var runtimeMap = map[string][]string{
	// TODO Replacing with the correct runtime is necessary
	// "python27": []string{"python", "abeja_runtime_python2"},
	"python36": {"py36.py"},
}

func createRuntimeBase(ctx context.Context, command string) (string, error) {
	statikFS, err := fs.New()
	if err != nil {
		return "", errors.Errorf("failed to open the runtime file system: %w", err)
	}
	f, err := statikFS.Open("/" + command)
	if err != nil {
		return "", errors.Errorf("failed to open the py36.py: %w", err)
	}
	defer f.Close()

	tempfile, err := ioutil.TempFile("", command)
	if err != nil {
		return "", errors.Errorf("failed to create temporary file: %w", err)
	}
	defer cleanutil.Close(ctx, tempfile, tempfile.Name())

	if _, err := io.Copy(tempfile, f); err != nil {
		return "", errors.Errorf("failed to write runtime base: %w", err)
	}
	return tempfile.Name(), nil
}

func handleSignal(ctx context.Context, runtime *subprocess.Runtime, subChan chan error) int {
	var sigReceived bool = false
	var subReceived bool = false
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT)
	defer close(sigChan)

	log.Debug(ctx, "waiting signal...")
	func() {
		for {
			select {
			case sig := <-sigChan:
				log.Infof(ctx, "signal[%s] received.", sig.String())
				sigReceived = true
				return
			case err, received := <-subChan:
				if received {
					log.Warning(ctx, "Error when waiting finish runtime:", err)
					subReceived = true
				}
				return
			default:
				if runtime.IsExited(ctx) {
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	if !runtime.IsExited(ctx) {
		runtime.Shutdown(ctx, 25*time.Second)
	}

	if sigReceived {
		log.Warning(ctx, "training-process finished with signal")
		return 1
	}
	if subReceived {
		return 1
	}
	if runtime.Status == subprocess.RuntimeStatusExitedWithSuccess {
		return 0
	}
	return 1
}

func Train(ctx context.Context, conf *config.Configuration) error {
	command, ok := runtimeMap[conf.Runtime]
	if !ok {
		return errors.Errorf("unsupported runtime language: %s", conf.Runtime)
	}

	workingDir, err := conf.GetWorkingDir()
	if err != nil {
		return errors.Errorf("failed to get working direcoty path: %w", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		return errors.Errorf("failed to move working direcoty path: %s: %w", workingDir, err)
	}

	runtimeBasePath, err := createRuntimeBase(ctx, command[0])
	if err != nil {
		return errors.Errorf("failed to craete Runtime base: %w", err)
	}
	defer cleanutil.Remove(ctx, runtimeBasePath)

	trainingResultDir, err := conf.GetTrainingResultDir()
	if err != nil {
		return errors.Errorf("failed to get path for training-result: %w", err)
	}
	if err := cmdutil.CreateDirIfNotExist(trainingResultDir); err != nil {
		return errors.Errorf("failed to create directory for training-result: %w", err)
	}

	// for trap error of Cmd.Wait()
	subChan := make(chan error)
	runtime, err := subprocess.CreateTrainRuntime(conf, runtimeBasePath, trainingResultDir)
	if err != nil {
		return errors.Errorf("failed to create runtime: %w", err)
	}

	// subprocess logger
	scopeChan := make(chan context.Context)
	runtimeLogger := subprocess.NewRuntimeLogger(ctx, runtime.Cmd, scopeChan)

	if err = runtime.Start(subChan); err != nil {
		return errors.Errorf("failed to start runtime: %w", err)
	}

	runtimeLogger.Run()
	defer runtimeLogger.Flush(3) // wait 3 seconds for flush all logs.
	close(scopeChan)

	exitStatus := handleSignal(ctx, runtime, subChan)
	if exitStatus > 0 {
		return errors.New("failed to training-process")
	}
	return nil
}
