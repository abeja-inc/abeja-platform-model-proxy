package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/sirupsen/logrus"

	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

const maxLogSize = 1024 * 250

var (
	procCtx context.Context
	reqCtx  context.Context
)

type RuntimeLogger struct {
	stdout *bufio.Reader
	stderr *bufio.Reader
	ch     chan context.Context
	wg     sync.WaitGroup
}

func NewRuntimeLogger(ctx context.Context, cmd *exec.Cmd, scopeChan chan context.Context) *RuntimeLogger {
	var stdoutReader, stderrReader *bufio.Reader
	procCtx = ctx
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Warning(ctx, "failed to get stdout of subprocess: ", err)
	} else {
		stdoutReader = bufio.NewReader(stdout)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Warning(ctx, "failed to get stderr of subprocess: ", err)
	} else {
		stderrReader = bufio.NewReader(stderr)
	}

	return &RuntimeLogger{
		stdout: stdoutReader,
		stderr: stderrReader,
		ch:     scopeChan,
	}
}

func (rl *RuntimeLogger) Flush(timeout time.Duration) {
	if !waitWithTimeout(&rl.wg, timeout*time.Second) {
		fmt.Println(
			"Timed out waiting for log flushing. " +
				"There may be some logs left that have not been written out.")
	}
}

func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return true // completed normally
	case <-time.After(timeout):
		return false // timed out
	}
}

func (rl *RuntimeLogger) Run() {

	go func() {
		for {
			if ctx, ok := <-rl.ch; !ok {
				return
			} else {
				reqCtx = ctx
			}
		}
	}()

	rl.wg.Add(2)
	go func() {
		proxySubprocessLogs(rl.stdout, logrus.InfoLevel)
		rl.wg.Done()
	}()
	go func() {
		proxySubprocessLogs(rl.stderr, logrus.WarnLevel)
		rl.wg.Done()
	}()
}

func proxySubprocessLogs(reader *bufio.Reader, defaultLogLevel logrus.Level) {
	if reader == nil {
		return
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && !isEOForPathError(err) {
			outputLog(fmt.Sprintf("failed to scanning user output: %T", err), logrus.WarnLevel)
			continue
		}
		// The last \n is included, so remove it.
		line = strings.TrimSpace(line)
		if len(line) > maxLogSize {
			prefix := string([]rune(line)[:64])
			outputLog(
				fmt.Sprintf("user output is too long. Maximum of 250kB per line. [%s...]", prefix),
				logrus.WarnLevel)
			continue
		}

		parseAndOutputLog(line, defaultLogLevel)

		if err != nil && isEOForPathError(err) {
			break
		}
	}
}

func isEOForPathError(err error) bool {
	if err == io.EOF {
		return true
	}
	if _, ok := err.(*os.PathError); ok {
		return true
	}
	return false
}

func parseAndOutputLog(text string, defaultLogLevel logrus.Level) {
	if strings.TrimSpace(text) == "" {
		// empty line
		return
	}
	jsonObj, err := simplejson.NewJson([]byte(text))
	if err != nil {
		// output of subprocess is plain text
		outputLog(text, defaultLogLevel)
		return
	}

	escapedJson, err := json.Marshal(text)
	if err != nil {
		outputLog(text, defaultLogLevel)
		return
	}

	levelStr, err := jsonObj.Get("log_level").String()
	if err != nil {
		// no log_level field in json
		outputLog(string(escapedJson), defaultLogLevel)
		return
	}
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		// unknown level
		outputLog(string(escapedJson), defaultLogLevel)
		return
	}

	outputLog(string(escapedJson), level)
}

func outputLog(text string, level logrus.Level) {
	if reqCtx != nil {
		log.Log(reqCtx, level, text)
	} else {
		log.Log(procCtx, level, text)
	}
}
