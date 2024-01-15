package service

import (
	"bytes"
	"strings"
	"testing"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
)

func TestSetupRunConfiguration(t *testing.T) {

	cases := []struct {
		name          string
		optionEnv     cmdutil.AllOptions
		optionCmdLine cmdutil.AllOptions
		hasError      bool
		expects       cmdutil.AllOptions
		errMsg        string
	}{
		{
			name:          "all default",
			optionEnv:     cmdutil.AllOptions{},
			optionCmdLine: cmdutil.AllOptions{},
			hasError:      false,
			expects: cmdutil.AllOptions{
				AbejaRuntime: config.DefaultRuntime,
				Port:         config.DefaultHTTPListenPort,
			},
			errMsg: "",
		}, {
			name: "env full",
			optionEnv: cmdutil.AllOptions{
				AbejaRuntime: "golang",
				Port:         8080,
			},
			optionCmdLine: cmdutil.AllOptions{},
			hasError:      false,
			expects: cmdutil.AllOptions{
				AbejaRuntime: "golang",
				Port:         8080,
			},
			errMsg: "",
		}, {
			name:      "cmdline full",
			optionEnv: cmdutil.AllOptions{},
			optionCmdLine: cmdutil.AllOptions{
				AbejaRuntime: "golang",
				Port:         8080,
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaRuntime: "golang",
				Port:         8080,
			},
			errMsg: "",
		}, {
			name: "cmdline takes precedence",
			optionEnv: cmdutil.AllOptions{
				AbejaRuntime: "golang",
				Port:         8080,
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaRuntime: "python27",
				Port:         8081,
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaRuntime: "python27",
				Port:         8081,
			},
			errMsg: "",
		}, {
			name:      "port number too small",
			optionEnv: cmdutil.AllOptions{},
			optionCmdLine: cmdutil.AllOptions{
				Port: 80,
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: port [80] must be greater than 1023 and less than 65536",
		}, {
			name: "port number too large",
			optionEnv: cmdutil.AllOptions{
				Port: 65536,
			},
			optionCmdLine: cmdutil.AllOptions{},
			hasError:      true,
			expects:       cmdutil.AllOptions{},
			errMsg:        "Error: port [65536] must be greater than 1023 and less than 65536",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmdutil.CleanUp(t)
			confRun = config.NewConfiguration()
			cmdutil.SetOptionsToEnv(c.optionEnv)
			cmdutil.SetOptionsToCmdline("run", c.optionCmdLine)
			cmdRun := newCmdRun()
			orgRunE := cmdRun.RunE
			cmdRun.RunE = cmdutil.DummyRunEFunc
			defer func() {
				cmdRun.RunE = orgRunE
			}()
			buf := new(bytes.Buffer)
			cmdRun.SetOutput(buf)

			err := cmdRun.Execute()
			if err != nil {
				if c.hasError {
					get := buf.String()
					if !strings.HasPrefix(get, c.errMsg) {
						t.Fatalf("error message should be start with [%s], but [%s]", c.errMsg, get)
					}
					return
				} else {
					t.Fatalf("unexpected error occurred: %s", err.Error())
				}
			}

			if confRun.Runtime != c.expects.AbejaRuntime {
				t.Errorf("AbejaRuntime should be %s, but %s", c.expects.AbejaRuntime, confRun.Runtime)
			}
			if confRun.Port != c.expects.Port {
				t.Errorf("Port should be %d, but %d", c.expects.Port, confRun.Port)
			}
		})
	}
}
