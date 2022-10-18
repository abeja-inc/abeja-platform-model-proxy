package training

import (
	"bytes"
	"strings"
	"testing"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
)

func TestSetupDownloadConfiguration(t *testing.T) {

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
				AbejaOrganizationID:          "1111111111111",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: 1,
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                  config.DefaultAbejaAPIURL,
				AbejaOrganizationID:          "1111111111111",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: 1,
				AbejaUserModelRoot:           "",
			},
			errMsg: "",
		}, {
			name: "env full",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                  "https://api.dev.abeja.io",
				AbejaOrganizationID:          "1111111111111",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: 1,
				AbejaUserModelRoot:           "tmp",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                  "https://api.dev.abeja.io",
				AbejaOrganizationID:          "1111111111111",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: 1,
				AbejaUserModelRoot:           "tmp",
			},
			errMsg: "",
		}, {
			name: "missing multiple required field",
			optionEnv: cmdutil.AllOptions{
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: 1,
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: require flag(s) abeja_organization_id not set",
		}, {
			name: "invalid traininig job definition version",
			optionEnv: cmdutil.AllOptions{
				AbejaOrganizationID:          "1111111111111",
				PlatformAuthToken:            "dummy",
				TrainingJobDefinitionName:    "sample",
				TrainingJobDefinitionVersion: -1,
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: : training_job_definition_version [-1] must be greater than 0",
		}, {
			name: "basic auth",
			optionEnv: cmdutil.AllOptions{
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "foo",
				AbejaPlatformPersonalAccessToken: "foobar",
				TrainingJobDefinitionName:        "sample",
				TrainingJobDefinitionVersion:     1,
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                      config.DefaultAbejaAPIURL,
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "foo",
				AbejaPlatformPersonalAccessToken: "foobar",
				TrainingJobDefinitionName:        "sample",
				TrainingJobDefinitionVersion:     1,
				AbejaUserModelRoot:               "",
			},
			errMsg: "",
		}, {
			name: "auth token error",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
				TrainingJobDefinitionName:        "sample",
				TrainingJobDefinitionVersion:     1,
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
			confDownload = config.NewConfiguration()
			cmdutil.SetOptionsToEnv(c.optionEnv)
			cmdDownload := newCmdDownload()
			orgRunE := cmdDownload.RunE
			cmdDownload.RunE = cmdutil.DummyRunEFunc
			defer func() {
				cmdDownload.RunE = orgRunE
			}()
			buf := new(bytes.Buffer)
			cmdDownload.SetOutput(buf)

			err := cmdDownload.Execute()
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
				if confDownload.APIURL != c.expects.AbejaApiUrl {
					t.Errorf("AbejaApiUrl should be %s, but %s", c.expects.AbejaApiUrl, confDownload.APIURL)
				}
				if confDownload.OrganizationID != c.expects.AbejaOrganizationID {
					t.Errorf("OrganizationID should be %s, but %s", c.expects.AbejaOrganizationID, confDownload.OrganizationID)
				}
				if confDownload.PlatformAuthToken != c.expects.PlatformAuthToken {
					t.Errorf("PlatformAuthToken should be %s, but %s", c.expects.PlatformAuthToken, confDownload.PlatformAuthToken)
				}
				if confDownload.TrainingJobDefinitionName != c.expects.TrainingJobDefinitionName {
					t.Errorf("TrainingJobDefinitionName should be %s, but %s", c.expects.TrainingJobDefinitionName, confDownload.TrainingJobDefinitionName)
				}
				if confDownload.TrainingJobDefinitionVersion != c.expects.TrainingJobDefinitionVersion {
					t.Errorf("TrainingJobDefinitionVersion should be %d, but %d", c.expects.TrainingJobDefinitionVersion, confDownload.TrainingJobDefinitionVersion)
				}
				if confDownload.UserModelRoot != c.expects.AbejaUserModelRoot {
					t.Errorf("UserModelRoot should be %s, but %s", c.expects.AbejaUserModelRoot, confDownload.UserModelRoot)
				}
				if confDownload.PlatformUserID != c.expects.AbejaPlatformUserID {
					t.Errorf("PlatformUserID should be %s, but %s", c.expects.AbejaPlatformUserID, confDownload.PlatformUserID)
				}
				if confDownload.PlatformPersonalAccessToken != c.expects.AbejaPlatformPersonalAccessToken {
					t.Errorf("PlatformPersonalAccessToken should be %s, but %s", c.expects.AbejaPlatformPersonalAccessToken, confDownload.PlatformPersonalAccessToken)
				}
			}
		})
	}
}
