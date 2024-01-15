package training

import (
	"bytes"
	"strings"
	"testing"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	pathutil "github.com/abeja-inc/abeja-platform-model-proxy/util/path"
)

func TestSetupTrainConfiguration(t *testing.T) {

	cases := []struct {
		name      string
		optionEnv cmdutil.AllOptions
		hasError  bool
		expects   cmdutil.AllOptions
		errMsg    string
	}{
		{
			name: "normal",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:            "",
				AbejaOrganizationID:    "1111111111111",
				PlatformAuthToken:      "token",
				AbejaRuntime:           "",
				AbejaTrainingResultDir: "",
				AbejaUserModelRoot:     "",
				Datasets:               "{}",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:            config.DefaultAbejaAPIURL,
				AbejaOrganizationID:    "1111111111111",
				PlatformAuthToken:      "token",
				AbejaRuntime:           config.DefaultRuntime,
				AbejaTrainingResultDir: pathutil.DefaultTrainingResultDir,
				AbejaUserModelRoot:     "",
			},
			errMsg: "",
		}, {
			name: "env full",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:            "https://api.dev.abeja.io",
				AbejaOrganizationID:    "1111111111111",
				PlatformAuthToken:      "token",
				AbejaRuntime:           "golang",
				AbejaTrainingResultDir: "result_dir",
				AbejaUserModelRoot:     "tmp",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:            "https://api.dev.abeja.io",
				AbejaOrganizationID:    "1111111111111",
				PlatformAuthToken:      "token",
				AbejaRuntime:           "golang",
				AbejaTrainingResultDir: "result_dir",
				AbejaUserModelRoot:     "tmp",
			},
			errMsg: "",
		}, {
			name: "authToken",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "result_dir",
				AbejaUserModelRoot:               "tmp",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "result_dir",
				AbejaUserModelRoot:               "tmp",
			},
			errMsg: "",
		}, {
			name: "basic auth",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "foo",
				AbejaPlatformPersonalAccessToken: "foobar",
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "result_dir",
				AbejaUserModelRoot:               "tmp",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "foo",
				AbejaPlatformPersonalAccessToken: "foobar",
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "result_dir",
				AbejaUserModelRoot:               "tmp",
			},
			errMsg: "",
		}, {
			name: "auth error",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
				AbejaRuntime:                     "golang",
				AbejaTrainingResultDir:           "result_dir",
				AbejaUserModelRoot:               "tmp",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: : platform_auth_token or (abeja_platform_user_id and abeja_platform_personal_access_token) need but not set.",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmdutil.CleanUp(t)
			confTrain = config.NewConfiguration()
			cmdutil.SetOptionsToEnv(c.optionEnv)
			cmdTrain := newCmdTrain()
			orgRunE := cmdTrain.RunE
			cmdTrain.RunE = cmdutil.DummyRunEFunc
			defer func() {
				cmdTrain.RunE = orgRunE
			}()
			buf := new(bytes.Buffer)
			cmdTrain.SetOutput(buf)

			err := cmdTrain.Execute()
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
				if confTrain.APIURL != c.expects.AbejaApiUrl {
					t.Errorf("APIURL should be %s, but %s", c.expects.AbejaApiUrl, confTrain.APIURL)
				}
				if confTrain.OrganizationID != c.expects.AbejaOrganizationID {
					t.Errorf("OrganizationID should be %s, but %s", c.expects.AbejaOrganizationID, confTrain.OrganizationID)
				}
				if confTrain.PlatformAuthToken != c.expects.PlatformAuthToken {
					t.Errorf("PlatformAuthToken should be %s, but %s", c.expects.PlatformAuthToken, confTrain.PlatformAuthToken)
				}
				if confTrain.Runtime != c.expects.AbejaRuntime {
					t.Errorf("Runtime should be %s, but %s", c.expects.AbejaRuntime, confTrain.Runtime)
				}
				if confTrain.TrainingResultDir != c.expects.AbejaTrainingResultDir {
					t.Errorf("TrainingResultDir should be %s, but %s", c.expects.AbejaTrainingResultDir, confTrain.TrainingResultDir)
				}
				if confTrain.UserModelRoot != c.expects.AbejaUserModelRoot {
					t.Errorf("UserModelRoot should be %s, but %s", c.expects.AbejaUserModelRoot, confTrain.UserModelRoot)
				}
				if confTrain.PlatformUserID != c.expects.AbejaPlatformUserID {
					t.Errorf("PlatformUserID should be %s, but %s", c.expects.AbejaPlatformUserID, confTrain.PlatformUserID)
				}
				if confTrain.PlatformPersonalAccessToken != c.expects.AbejaPlatformPersonalAccessToken {
					t.Errorf("PlatformPersonalAccessToken should be %s, but %s", c.expects.AbejaPlatformPersonalAccessToken, confTrain.PlatformPersonalAccessToken)
				}
			}
		})
	}
}
