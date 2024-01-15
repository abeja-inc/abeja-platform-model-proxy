package tensorboard

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
)

func TestSetupDefaultConfiguration(t *testing.T) {

	cases := []struct {
		name      string
		optionEnv cmdutil.AllOptions
		hasError  bool
		expects   cmdutil.AllOptions
		errMsg    string
	}{
		{
			name: "missing multi required",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:               "",
				AbejaOrganizationID:       "",
				PlatformAuthToken:         "",
				TrainingJobDefinitionName: "xxx",
				TrainingJobIDS:            "12345",
				TensorboardID:             "10000",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: require flag(s) abeja_organization_id, platform_auth_token not set",
		}, {
			name: "env and cmdline full",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:               "https://api.dev.abeja.io",
				AbejaOrganizationID:       "1111111111111",
				PlatformAuthToken:         "aaaaaaaaaa",
				TrainingJobIDS:            "1,2,3,4",
				TrainingJobDefinitionName: "5555555555555",
				TensorboardID:             "1230000000000",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:               "https://api.dev.abeja.io",
				AbejaOrganizationID:       "1111111111111",
				PlatformAuthToken:         "aaaaaaaaaa",
				TrainingJobIDS:            "1,2,3,4",
				TrainingJobDefinitionName: "5555555555555",
				TensorboardID:             "1230000000000",
			},
			errMsg: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmdutil.CleanUp(t)
			confDefault = config.NewConfiguration()
			cmdutil.SetOptionsToEnv(c.optionEnv)
			cmdRoot := newCmdRoot(context.TODO())
			orgRunE := cmdRoot.RunE
			cmdRoot.RunE = cmdutil.DummyRunEFunc
			defer func() {
				cmdRoot.RunE = orgRunE
			}()
			buf := new(bytes.Buffer)
			cmdRoot.SetOutput(buf)

			err := cmdRoot.Execute()
			if err != nil {
				if c.hasError {
					get := buf.String()
					if !strings.HasPrefix(get, c.errMsg) {
						t.Fatalf("error message should be start with [%s], but [%s]\ncmdline args: [%v]", c.errMsg, get, os.Args)
					}
					return
				} else {
					t.Fatalf("unexpected error occurred: %s", err.Error())
				}
			}

			if confDefault.APIURL != c.expects.AbejaApiUrl {
				t.Errorf("AbejaApiUrl should be %s, but %s", c.expects.AbejaApiUrl, confDefault.APIURL)
			}
			if confDefault.OrganizationID != c.expects.AbejaOrganizationID {
				t.Errorf("OrganizationID should be %s, but %s", c.expects.AbejaOrganizationID, confDefault.OrganizationID)
			}
			if confDefault.PlatformAuthToken != c.expects.PlatformAuthToken {
				t.Errorf("PlatformAuthToken should be %s, but %s", c.expects.PlatformAuthToken, confDefault.PlatformAuthToken)
			}
			if confDefault.TrainingJobIDS != c.expects.TrainingJobIDS {
				t.Errorf("TrainingJobIDS should be %s, but %s", c.expects.TrainingJobIDS, confDefault.TrainingJobIDS)
			}
			if confDefault.TrainingJobDefinitionName != c.expects.TrainingJobDefinitionName {
				t.Errorf("TrainingJobDefinitionName should be %s, but %s", c.expects.TrainingJobDefinitionName, confDefault.TrainingJobDefinitionName)
			}
			if confDefault.TensorboardID != c.expects.TensorboardID {
				t.Errorf("TensorboardID should be %s, but %s", c.expects.TensorboardID, confDefault.TensorboardID)
			}
		})
	}
}
