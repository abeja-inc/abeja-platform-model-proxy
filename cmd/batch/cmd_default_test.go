package batch

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
			name: "missing required organization id",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "",
				AbejaDeploymentCodeDownload:      "",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				AbejaTrainingModelDownload:       "",
				TrainingJobDefinitionName:        "",
				TrainingJobID:                    "",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: require flag(s) abeja_organization_id not set",
		}, {
			name: "missing auth",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				AbejaDeploymentCodeDownload:      "",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				AbejaTrainingModelDownload:       "",
				TrainingJobDefinitionName:        "",
				TrainingJobID:                    "",
				PlatformAuthToken:                "",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: platform_auth_token or (abeja_platform_user_id and abeja_platform_personal_access_token) need but not set.",
		}, {
			name: "missing deployment code version",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				AbejaDeploymentCodeDownload:      "",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				AbejaTrainingModelDownload:       "",
				TrainingJobDefinitionName:        "",
				TrainingJobID:                    "",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: abeja_deployment_code_download or (abeja_organization_id and abeja_model_id and abeja_model_version_id) need but not set.",
		}, {
			name: "missing trainingJobDefinitionName for Training Model",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				AbejaDeploymentCodeDownload:      "organizations/1111111111111/deployments/2222222222222/versions/ver-3333333333333/download",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				AbejaTrainingModelDownload:       "",
				TrainingJobDefinitionName:        "",
				TrainingJobID:                    "job-4444444444444",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
			},
			hasError: true,
			expects:  cmdutil.AllOptions{},
			errMsg:   "Error: When using the training result, please specify abeja_training_model_download or (abeja_organization_id, training_job_definition_name and training_job_id).",
		}, {
			name: "normal w/o model",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				AbejaDeploymentCodeDownload:      "organizations/1111111111111/deployments/2222222222222/versions/ver-3333333333333/download",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				AbejaTrainingModelDownload:       "",
				TrainingJobDefinitionName:        "",
				TrainingJobID:                    "",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaDeploymentCodeDownload: "organizations/1111111111111/deployments/2222222222222/versions/ver-3333333333333/download",
				PlatformAuthToken:           "token",
			},
			errMsg: "",
		}, {
			name: "normal w/ model",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				AbejaDeploymentCodeDownload:      "organizations/1111111111111/deployments/2222222222222/versions/ver-3333333333333/download",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				AbejaTrainingModelDownload:       "",
				TrainingJobDefinitionName:        "training",
				TrainingJobID:                    "job-111",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaDeploymentCodeDownload: "organizations/1111111111111/deployments/2222222222222/versions/ver-3333333333333/download",
				TrainingJobDefinitionName:   "training",
				TrainingJobID:               "job-111",
				PlatformAuthToken:           "token",
			},
			errMsg: "",
		}, {
			name: "normal w/ model 2",
			optionEnv: cmdutil.AllOptions{
				AbejaApiUrl:                      "https://api.dev.abeja.io",
				AbejaOrganizationID:              "1111111111111",
				AbejaDeploymentCodeDownload:      "organizations/1111111111111/deployments/2222222222222/versions/ver-3333333333333/download",
				AbejaModelID:                     "",
				AbejaModelVersionID:              "",
				AbejaTrainingModelDownload:       "organizations/1111111111111/training/definitions/test_train/models/2222222222222/download",
				TrainingJobDefinitionName:        "",
				TrainingJobID:                    "",
				PlatformAuthToken:                "token",
				AbejaPlatformUserID:              "",
				AbejaPlatformPersonalAccessToken: "",
			},
			hasError: false,
			expects: cmdutil.AllOptions{
				AbejaApiUrl:                 "https://api.dev.abeja.io",
				AbejaOrganizationID:         "1111111111111",
				AbejaDeploymentCodeDownload: "organizations/1111111111111/deployments/2222222222222/versions/ver-3333333333333/download",
				AbejaTrainingModelDownload:  "organizations/1111111111111/training/definitions/test_train/models/2222222222222/download",
				PlatformAuthToken:           "token",
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
					t.Fatal("unexpected error occurred:", err)
				}
			} else {
				if c.hasError {
					t.Fatal("err should be occurred. but it doesn't")
				}
				if confDefault.APIURL != c.expects.AbejaApiUrl {
					t.Errorf("AbejaApiUrl should be %s, but %s", c.expects.AbejaApiUrl, confDefault.APIURL)
				}
				if confDefault.OrganizationID != c.expects.AbejaOrganizationID {
					t.Errorf("OrganizationID should be %s, but %s", c.expects.AbejaOrganizationID, confDefault.OrganizationID)
				}
				if confDefault.DeploymentCodeDownload != c.expects.AbejaDeploymentCodeDownload {
					t.Errorf("DeploymentCodeDownload should be %s, but %s", c.expects.AbejaDeploymentCodeDownload, confDefault.DeploymentCodeDownload)
				}
				if confDefault.ModelID != c.expects.AbejaModelID {
					t.Errorf("ModelID should be %s, but %s", c.expects.AbejaModelID, confDefault.ModelID)
				}
				if confDefault.ModelVersionID != c.expects.AbejaModelVersionID {
					t.Errorf("ModelVersionID should be %s, but %s", c.expects.AbejaModelVersionID, confDefault.ModelVersionID)
				}
				if confDefault.TrainingModelDownload != c.expects.AbejaTrainingModelDownload {
					t.Errorf("TrainingModelDownload should be %s, but %s", c.expects.AbejaTrainingModelDownload, confDefault.TrainingModelDownload)
				}
				if confDefault.TrainingJobDefinitionName != c.expects.TrainingJobDefinitionName {
					t.Errorf("TrainingJobDefinitionName should be %s, but %s", c.expects.TrainingJobDefinitionName, confDefault.TrainingJobDefinitionName)
				}
				if confDefault.TrainingJobID != c.expects.TrainingJobID {
					t.Errorf("TrainingJobID should be %s, but %s", c.expects.TrainingJobID, confDefault.TrainingJobID)
				}
				if confDefault.PlatformAuthToken != c.expects.PlatformAuthToken {
					t.Errorf("PlatformAuthToken should be %s, but %s", c.expects.PlatformAuthToken, confDefault.PlatformAuthToken)
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
