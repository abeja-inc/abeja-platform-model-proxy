package training

import (
	"bytes"
	"context"
	"strings"
	"testing"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
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
				AbejaApiUrl:                  "",
				AbejaOrganizationID:          "",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "",
				TrainingJobDefinitionVersion: 1,
				AbejaRuntime:                 "golang",
				AbejaTrainingResultDir:       "",
				AbejaUserModelRoot:           "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: require flag(s) abeja_organization_id, training_job_definition_name not set",
		}, {
			name: "env full",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                  "https://api.dev.abeja.io",
				AbejaOrganizationID:          "1111111111111",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: 1,
				AbejaRuntime:                 "golang",
				AbejaTrainingResultDir:       "result_dir",
				AbejaUserModelRoot:           "tmp",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                  "https://api.dev.abeja.io",
				AbejaOrganizationID:          "1111111111111",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: 1,
				AbejaRuntime:                 "golang",
				AbejaTrainingResultDir:       "result_dir",
				AbejaUserModelRoot:           "tmp",
			},
			errMsg: "",
		}, {
			name: "auth token error",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
				TrainingJobDefinitionName:        "sample",
				TrainingJobDefinitionVersion:     1,
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "",
				AbejaUserModelRoot:               "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: : platform_auth_token or (abeja_platform_user_id and abeja_platform_personal_access_token) need but not set.",
		}, {
			name: "basic auth",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "dummy",
				AbejaPlatformUserID:              "foo",
				AbejaPlatformPersonalAccessToken: "foobar",
				TrainingJobDefinitionName:        "sample",
				TrainingJobDefinitionVersion:     1,
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "result_dir",
				AbejaUserModelRoot:               "tmp",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "dummy",
				AbejaPlatformUserID:              "foo",
				AbejaPlatformPersonalAccessToken: "foobar",
				TrainingJobDefinitionName:        "sample",
				TrainingJobDefinitionVersion:     1,
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "result_dir",
				AbejaUserModelRoot:               "tmp",
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
						t.Fatalf("error message should be start with [%s], but [%s]", c.errMsg, get)
					}
					return
				} else {
					t.Fatalf("unexpected error occurred: %s", err.Error())
				}
			} else {
				if c.hasError {
					t.Fatalf("err should be occurred. but it doesn't.")
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
				if confDefault.TrainingJobDefinitionName != c.expects.TrainingJobDefinitionName {
					t.Errorf("TrainingJobDefinitionName should be %s, but %s", c.expects.TrainingJobDefinitionName, confDefault.TrainingJobDefinitionName)
				}
				if confDefault.TrainingJobDefinitionVersion != c.expects.TrainingJobDefinitionVersion {
					t.Errorf("TrainingJobDefinitionVersion should be %d, but %d", c.expects.TrainingJobDefinitionVersion, confDefault.TrainingJobDefinitionVersion)
				}
				if confDefault.Runtime != c.expects.AbejaRuntime {
					t.Errorf("Runtime should be %s, but %s", c.expects.AbejaRuntime, confDefault.Runtime)
				}
				if confDefault.TrainingResultDir != c.expects.AbejaTrainingResultDir {
					t.Errorf("TrainingResultDir should be %s, but %s", c.expects.AbejaTrainingResultDir, confDefault.TrainingResultDir)
				}
				if confDefault.UserModelRoot != c.expects.AbejaUserModelRoot {
					t.Errorf("UserModelRoot should be %s, but %s", c.expects.AbejaUserModelRoot, confDefault.UserModelRoot)
				}
				if confDefault.PlatformUserID != c.expects.AbejaPlatformUserID {
					t.Errorf("PlatformUserID should be %s, but %s", c.expects.AbejaPlatformUserID, confDefault.PlatformUserID)
				}
				if confDefault.PlatformPersonalAccessToken != c.expects.AbejaPlatformPersonalAccessToken {
					t.Errorf("PlatformPersonalAccessToken should be %s, but %s", c.expects.AbejaPlatformPersonalAccessToken, confDefault.PlatformPersonalAccessToken)
				}
			}
		})
	}
}
