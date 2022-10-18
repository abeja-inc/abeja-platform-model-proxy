package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/abeja-inc/platform-model-proxy/config"
)

func TestMain(m *testing.M) {

	os.Args = []string{"cmd"}
	code := m.Run()

	conf := config.NewConfiguration()
	tmpdir := conf.RequestedDataDir
	if err := os.RemoveAll(tmpdir); err != nil {
		fmt.Printf("failed to tmpdir: %s\n", tmpdir)
	}
	os.Exit(code)
}
