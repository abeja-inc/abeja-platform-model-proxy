package subprocess

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
)

// runtimeMap represents mapping of programming-language and executable-file-path.
var runtimeMap = map[string][]string{
	// TODO Replacing with the correct runtime is necessary
	// "python27": []string{"python", "abeja_runtime_python2"},
	"python36": []string{"abeja-runtime-python"},
}

// languageMap represents mapping of programming-language and executable-file-path.
var languageMap = map[string]string{
	// TODO Replacing with the correct runtime is necessary
	"python36": "python3",
}

// allowedExitStatusMap represents mapping of Non-zero but allowable exit-statuses per runtime.
var allowedExitStatusMap = map[string][]int{
	"python36": []int{120},
}

// Runtime represents process information(exec.Cmd) of runtime-process
// and status of runtime-process.
type Runtime struct {
	Cmd         *exec.Cmd
	Status      RuntimeStatus
	RuntimeType string
}

// RuntimeStatus represents status of runtime.
type RuntimeStatus int

// Statuses for RuntimeStatus.
const (
	RuntimeStatusPreparing RuntimeStatus = iota
	RuntimeStatusRunning
	RuntimeStatusExitedWithSuccess
	RuntimeStatusExitedWithFailure
)

// CreateServiceRuntime starts runtime(subprocess).
func CreateServiceRuntime(
	conf *config.Configuration,
	udsFilePath string,
	trainingResultDir string) (*Runtime, error) {

	command, ok := runtimeMap[conf.Runtime]
	if !ok {
		return nil, errors.Errorf("unsupported runtime language: %s", conf.Runtime)
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("ABEJA_IPC_PATH=%s", udsFilePath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ABEJA_TRAINING_RESULT_DIR=%s", trainingResultDir))

	runtime := &Runtime{
		Cmd:         cmd,
		Status:      RuntimeStatusPreparing,
		RuntimeType: conf.Runtime,
	}
	return runtime, nil
}

// CreateTrainRuntime starts runtime(subprocess).
func CreateTrainRuntime(
	conf *config.Configuration,
	runtimeBasePath string,
	trainingResultDir string) (*Runtime, error) {
	command, ok := languageMap[conf.Runtime]
	if !ok {
		return nil, errors.Errorf("unsupported runtime language: %s", conf.Runtime)
	}
	cmd := exec.Command(command, runtimeBasePath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("ABEJA_TRAINING_RESULT_DIR=%s", trainingResultDir))

	// replace ABEJA_PLATFORM_USER_ID because compensate 'user-'
	authInfo := conf.GetAuthInfo()
	cmd.Env = append(cmd.Env, fmt.Sprintf("ABEJA_PLATFORM_USER_ID=%s", authInfo.UserID))

	runtime := &Runtime{
		Cmd:         cmd,
		Status:      RuntimeStatusPreparing,
		RuntimeType: conf.Runtime,
	}
	return runtime, nil
}

// CreateOneshotRuntime creates runtime(subprocess).
func CreateOneshotRuntime(
	conf *config.Configuration,
	udsFilePath string,
	trainingResultDir string) (*Runtime, error) {

	command, ok := runtimeMap[conf.Runtime]
	if !ok {
		return nil, errors.Errorf("unsupported runtime language: %s", conf.Runtime)
	}

	cmd := exec.Command(command[0], command[1:]...)

	cmd.Env = append(os.Environ(), fmt.Sprintf("ABEJA_IPC_PATH=%s", udsFilePath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("ABEJA_TRAINING_RESULT_DIR=%s", trainingResultDir))

	runtime := &Runtime{
		Cmd:         cmd,
		Status:      RuntimeStatusPreparing,
		RuntimeType: conf.Runtime,
	}
	return runtime, nil
}

// Stop stops subprocess asynchronousely.
// (send signal(SIGINT) to subprocess)
func (r *Runtime) Stop(ctx context.Context) {
	if err := r.Cmd.Process.Signal(syscall.SIGINT); err != nil {
		log.Warning(ctx, "Error when sending signal to runtime:", err)
	}
}

// Kill kills subprocess.
// (send signal(SIGKILL) to subprocess)
func (r *Runtime) Kill(ctx context.Context) {
	if err := r.Cmd.Process.Signal(syscall.SIGKILL); err != nil {
		log.Warning(ctx, "Error when sending signal to runtime:", err)
	}
}

// IsReady returns result of `Is subprocess ready ?`.
func (r *Runtime) IsReady() bool {
	return r.Status == RuntimeStatusRunning
}

// IsExited returns result of `Is subprocess exited ?`.
func (r *Runtime) IsExited(ctx context.Context) bool {
	status := r.getStatus(ctx)
	if status == RuntimeStatusPreparing || status == RuntimeStatusRunning {
		return false
	}
	return true
}

// Start starts runtime.
func (r *Runtime) Start(errOnSub chan error) error {
	if err := r.Cmd.Start(); err != nil {
		return err
	}
	go func() {
		defer close(errOnSub)
		if err := r.Cmd.Wait(); err != nil {
			switch err := err.(type) {
			case *exec.ExitError:
				allowedExitStatuses := allowedExitStatusMap[r.RuntimeType]
				exitCode := err.ExitCode()
				if !contains(allowedExitStatuses, exitCode) {
					errOnSub <- err
				}
			default:
				errOnSub <- err
			}
		}
	}()
	return nil
}

// WaitUntilStarted waits to bootstraping of subprocess.
// NOTE: Although there is a dependence on the order of function calls,
// it is not very good...
func (r *Runtime) WaitUntilStarted(ctx context.Context, socketPath string) error {
	for {
		if _, err := os.Stat(socketPath); err == nil {
			log.Debug(ctx, "runtime started")
			r.Status = RuntimeStatusRunning
			return nil
		}
		if r.IsExited(ctx) {
			log.Warning(ctx, "runtime stopped unexpectedly")
			return errors.Errorf("runtime stopped unexpectedly: %s", r.Cmd.ProcessState.String())
		}
		log.Debug(ctx, "runtime bootstrapping yet...")
		time.Sleep(1 * time.Second)
	}
}

// RuntimeStatus gets status of subprocess.
func (r *Runtime) getStatus(ctx context.Context) RuntimeStatus {
	if r.Cmd.ProcessState == nil || !r.Cmd.ProcessState.Exited() {
		return r.Status
	}
	exitCode := r.Cmd.ProcessState.ExitCode()
	allowedExitStatuses := allowedExitStatusMap[r.RuntimeType]
	log.Infof(ctx, "runtime finished with exit-code: %d", exitCode)
	if r.Cmd.ProcessState.Success() || contains(allowedExitStatuses, exitCode) {
		r.Status = RuntimeStatusExitedWithSuccess
		return RuntimeStatusExitedWithSuccess
	}
	r.Status = RuntimeStatusExitedWithFailure
	return RuntimeStatusExitedWithFailure
}

// Shutdown waits stop subprocess.
func (r *Runtime) Shutdown(ctx context.Context, waitMax time.Duration) {
	ticker := *time.NewTicker(waitMax)
	r.Stop(ctx)
	for {
		select {
		case <-ticker.C:
			log.Debug(ctx, "ticker received. subprocess didn't stop.")
			ticker.Stop()
			if !r.IsExited(ctx) {
				log.Debug(ctx, "kill subprocess")
				r.Kill(ctx)
				return
			}
		default:
			if r.IsExited(ctx) {
				log.Debug(ctx, "subprocess stopped.")
				ticker.Stop()
				return
			}
		}
	}
}

func contains(s []int, e int) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
