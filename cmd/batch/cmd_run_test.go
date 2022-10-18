package batch

import (
	"bytes"
	"context"
	"strings"
	"testing"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
)

func TestSetupRunConfiguration(t *testing.T) {

	cases := []struct {
		name      string
		optionEnv cmdutil.AllOptions
		hasError  bool
		expects   cmdutil.AllOptions
		errMsg    string
	}{
		{
			name: "missing auth when datalake input",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
				Input:                            "$datalake:1:5555555555555",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: platform_auth_token or (abeja_platform_user_id and abeja_platform_personal_access_token) need but not set.",
		}, {
			name: "normal",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
				Input:                            "$datalake:1:5555555555555",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:         "https://api.dev.abeja.io",
				AbejaOrganizationID: "1111111111111",
				PlatformAuthToken:   "token",
				Input:               "$datalake:1:5555555555555",
			},
			errMsg: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmdutil.CleanUp(t)
			confRun = config.NewConfiguration()
			cmdutil.SetOptionsToEnv(c.optionEnv)
			procCtx = context.TODO()
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
					t.Fatal("unexpected error occurred:", err)
				}
			} else {
				if c.hasError {
					t.Fatal("err should be occurred. but it doesn't")
				}
				if confRun.APIURL != c.expects.AbejaApiUrl {
					t.Errorf("AbejaApiUrl should be %s, but %s", c.expects.AbejaApiUrl, confRun.APIURL)
				}
				if confRun.OrganizationID != c.expects.AbejaOrganizationID {
					t.Errorf("OrganizationID should be %s, but %s", c.expects.AbejaOrganizationID, confRun.OrganizationID)
				}
				if confRun.ModelID != c.expects.AbejaModelID {
					t.Errorf("ModelID should be %s, but %s", c.expects.AbejaModelID, confRun.ModelID)
				}
				if confRun.ModelVersionID != c.expects.AbejaModelVersionID {
					t.Errorf("ModelVersionID should be %s, but %s", c.expects.AbejaModelVersionID, confRun.ModelVersionID)
				}
				if confRun.PlatformAuthToken != c.expects.PlatformAuthToken {
					t.Errorf("PlatformAuthToken should be %s, but %s", c.expects.PlatformAuthToken, confRun.PlatformAuthToken)
				}
				if confRun.PlatformUserID != c.expects.AbejaPlatformUserID {
					t.Errorf("PlatformUserID should be %s, but %s", c.expects.AbejaPlatformUserID, confRun.PlatformUserID)
				}
				if confRun.PlatformPersonalAccessToken != c.expects.AbejaPlatformPersonalAccessToken {
					t.Errorf("PlatformPersonalAccessToken should be %s, but %s", c.expects.AbejaPlatformPersonalAccessToken, confRun.PlatformPersonalAccessToken)
				}
				if confRun.Input != c.expects.Input {
					t.Errorf("Input should be %s, but %s", c.expects.Input, confRun.Input)
				}
			}
		})
	}
}
