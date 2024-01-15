package clean

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
)

func Close(ctx context.Context, v interface{}, target string) {
	if d, ok := v.(io.Closer); ok {
		if err := d.Close(); err != nil {
			funcName, line := getCallerInfo()
			log.Warning(ctx, fmt.Sprintf("funcName: %s, line: %d, target: %s. Error when closing:", funcName, line, target), err)
		}
	}
}

func Remove(ctx context.Context, path string) {
	if err := os.Remove(path); err != nil {
		funcName, line := getCallerInfo()
		log.Warning(ctx, fmt.Sprintf("funcName: %s, line: %d, path: %s. Error when removing:", funcName, line, path), err)
	}
}

func RemoveAll(ctx context.Context, path string) {
	if err := os.RemoveAll(path); err != nil {
		funcName, line := getCallerInfo()
		log.Warning(ctx, fmt.Sprintf("funcName: %s, line: %d, path: %s. Error when removing:", funcName, line, path), err)
	}
}

func getCallerInfo() (string, int) {
	pt, _, line, ok := runtime.Caller(2)
	if !ok {
		return "Unknown", 0
	}
	funcName := runtime.FuncForPC(pt).Name()
	return funcName, line
}
