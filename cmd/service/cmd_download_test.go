package service

import (
	"bytes"
	"testing"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
)

func TestSetupDownloadConfiguration(t *testing.T) {

	cases := []struct {
		name          string
		optionEnv     cmdutil.AllOptions
		optionCmdLine cmdutil.AllOptions
		hasError      bool
		expects       cmdutil.AllOptions
		errMsg        string
	}{
		{
			name: "env missing required",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: require flag(s) abeja_organization_id not set",
		}, {
			name: "cmdline missing required",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: require flag(s) abeja_model_id not set",
		}, {
			name: "env minimal",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			errMsg: "",
		}, {
			name: "cmdline minimal",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			errMsg: "",
		}, {
			name: "cmdline takes precedence",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "aaaaaaaaaa",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "bbbbbbbbbb",
				AbejaDeploymentCodeDownload: "download",
				AbejaTrainingModelDownload:  "download",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaModelID:                "2222222222222",
				AbejaModelVersionID:         "3333333333333",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "bbbbbbbbbb",
				AbejaDeploymentCodeDownload: "download",
				AbejaTrainingModelDownload:  "download",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			errMsg: "",
		}, {
			name: "env full",
			optionEnv: cmdutil.AllOptions{
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
			},
			optionCmdLine: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
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
			},
			errMsg: "",
		}, {
			name: "cmdline full",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                 "",
				AbejaOrganizationID:         "",
				AbejaModelID:                "",
				AbejaModelVersionID:         "",
				AbejaUserModelRoot:          "",
				PlatformAuthToken:           "",
				AbejaDeploymentCodeDownload: "",
				AbejaTrainingModelDownload:  "",
				TrainingJobID:               "",
				TrainingJobDefinitionName:   "",
			},
			optionCmdLine: cmdutil.AllOptions{
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
			},
			errMsg: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmdutil.CleanUp(t)
			confDownload = config.NewConfiguration()
			cmdutil.SetOptionsToEnv(c.optionEnv)
			cmdutil.SetOptionsToCmdline("download", c.optionCmdLine)
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
					return
				} else {
					t.Fatalf("unexpected error occurred: %s", err.Error())
				}
			}

			if confDownload.APIURL != c.expects.AbejaApiUrl {
				t.Errorf("AbejaApiUrl should be %s, but %s", c.expects.AbejaApiUrl, confDownload.APIURL)
			}
			if confDownload.OrganizationID != c.expects.AbejaOrganizationID {
				t.Errorf("OrganizationID should be %s, but %s", c.expects.AbejaOrganizationID, confDownload.OrganizationID)
			}
			if confDownload.ModelID != c.expects.AbejaModelID {
				t.Errorf("ModelID should be %s, but %s", c.expects.AbejaModelID, confDownload.ModelID)
			}
			if confDownload.ModelVersionID != c.expects.AbejaModelVersionID {
				t.Errorf("ModelVersionID should be %s, but %s", c.expects.AbejaModelVersionID, confDownload.ModelVersionID)
			}
			if confDownload.UserModelRoot != c.expects.AbejaUserModelRoot {
				t.Errorf("UserModelRoot should be %s, but %s", c.expects.AbejaUserModelRoot, confDownload.UserModelRoot)
			}
			if confDownload.PlatformAuthToken != c.expects.PlatformAuthToken {
				t.Errorf("PlatformAuthToken should be %s, but %s", c.expects.PlatformAuthToken, confDownload.PlatformAuthToken)
			}
			if confDownload.DeploymentCodeDownload != c.expects.AbejaDeploymentCodeDownload {
				t.Errorf("DeploymentCodeDownload should be %s, but %s", c.expects.AbejaDeploymentCodeDownload, confDownload.DeploymentCodeDownload)
			}
			if confDownload.TrainingModelDownload != c.expects.AbejaTrainingModelDownload {
				t.Errorf("TrainingModelDownload should be %s, but %s", c.expects.AbejaTrainingModelDownload, confDownload.TrainingModelDownload)
			}
			if confDownload.TrainingJobID != c.expects.TrainingJobID {
				t.Errorf("TrainingJobID should be %s, but %s", c.expects.TrainingJobID, confDownload.TrainingJobID)
			}
			if confDownload.TrainingJobDefinitionName != c.expects.TrainingJobDefinitionName {
				t.Errorf("TrainingJobDefinitionName should be %s, but %s", c.expects.TrainingJobDefinitionName, confDownload.TrainingJobDefinitionName)
			}
		})
	}
}
