package service

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
)

func TestSetupDefaultConfiguration(t *testing.T) {

	cases := []struct {
		name          string
		optionEnv     cmdutil.AllOptions
		optionCmdLine cmdutil.AllOptions
		hasError      bool
		expects       cmdutil.AllOptions
		errMsg        string
	}{
		{
			name: "missing multi required",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
				AbejaRuntime:                "golang",
				Port:                        8080,
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
				AbejaRuntime:                "",
				Port:                        0,
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: require flag(s) abeja_organization_id, abeja_model_version_id not set",
		}, {
			name: "env and cmdline full",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "download",
				AbejaTrainingModelDownload:  "download",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "5555555555555",
				AbejaRuntime:                "",
				Port:                        8080,
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "tmp",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "4444444444444",
				TrainingJobDefinitionName:   "",
				AbejaRuntime:                "golang",
				Port:                        0,
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "tmp",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "download",
				AbejaTrainingModelDownload:  "download",
				TrainingJobID:               "4444444444444",
				TrainingJobDefinitionName:   "5555555555555",
				AbejaRuntime:                "golang",
				Port:                        8080,
			},
			errMsg: "",
		}, {
			name: "port number too small",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "download",
				AbejaTrainingModelDownload:  "download",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "5555555555555",
				AbejaRuntime:                "",
				Port:                        0,
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "tmp",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "4444444444444",
				TrainingJobDefinitionName:   "",
				AbejaRuntime:                "golang",
				Port:                        80,
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: port [80] must be greater than 1023 and less than 65536",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmdutil.CleanUp(t)
			confDefault = config.NewConfiguration()
			cmdutil.SetOptionsToEnv(c.optionEnv)
			cmdutil.SetOptionsToCmdline("", c.optionCmdLine)
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
			if confDefault.ModelID != c.expects.AbejaModelID {
				t.Errorf("ModelID should be %s, but %s", c.expects.AbejaModelID, confDefault.ModelID)
			}
			if confDefault.ModelVersionID != c.expects.AbejaModelVersionID {
				t.Errorf("ModelVersionID should be %s, but %s", c.expects.AbejaModelVersionID, confDefault.ModelVersionID)
			}
			if confDefault.UserModelRoot != c.expects.AbejaUserModelRoot {
				t.Errorf("UserModelRoot should be %s, but %s", c.expects.AbejaUserModelRoot, confDefault.UserModelRoot)
			}
			if confDefault.PlatformAuthToken != c.expects.PlatformAuthToken {
				t.Errorf("PlatformAuthToken should be %s, but %s", c.expects.PlatformAuthToken, confDefault.PlatformAuthToken)
			}
			if confDefault.DeploymentCodeDownload != c.expects.AbejaDeploymentCodeDownload {
				t.Errorf("DeploymentCodeDownload should be %s, but %s", c.expects.AbejaDeploymentCodeDownload, confDefault.DeploymentCodeDownload)
			}
			if confDefault.TrainingModelDownload != c.expects.AbejaTrainingModelDownload {
				t.Errorf("TrainingModelDownload should be %s, but %s", c.expects.AbejaTrainingModelDownload, confDefault.TrainingModelDownload)
			}
			if confDefault.TrainingJobID != c.expects.TrainingJobID {
				t.Errorf("TrainingJobID should be %s, but %s", c.expects.TrainingJobID, confDefault.TrainingJobID)
			}
			if confDefault.TrainingJobDefinitionName != c.expects.TrainingJobDefinitionName {
				t.Errorf("TrainingJobDefinitionName should be %s, but %s", c.expects.TrainingJobDefinitionName, confDefault.TrainingJobDefinitionName)
			}
			if confDefault.Runtime != c.expects.AbejaRuntime {
				t.Errorf("AbejaRuntime should be %s, but %s", c.expects.AbejaRuntime, confDefault.Runtime)
			}
			if confDefault.Port != c.expects.Port {
				t.Errorf("Port should be %d, but %d", c.expects.Port, confDefault.Port)
			}
		})
	}
}
