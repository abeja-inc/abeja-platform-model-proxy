package util

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var confEnvs = [...]string{
	"abeja_api_url",
	"abeja_organization_id",
	"abeja_model_id",
	"abeja_model_version",
	"abeja_model_version_id",
	"abeja_deployment_id",
	"abeja_deployment_code_download",
	"abeja_service_id",
	"abeja_training_model_download",
	"training_job_definition_name",
	"training_job_definition_version",
	"training_job_id",
	"abeja_run_id",
	"platform_auth_token",
	"abeja_platform_user_id",
	"abeja_platform_personal_access_token",
	"abeja_runtime",
	"abeja_user_model_root",
	"abeja_training_result_dir",
	"datasets",
	"input",
	"output",
	"port",
}

func CleanUp(t *testing.T) {
	for _, env := range confEnvs {
		if err := os.Unsetenv(strings.ToUpper(env)); err != nil {
			t.Fatal("Error when unsetenv:", err)
		}
	}
	os.Args = []string{"cmd"}
}

type AllOptions struct {
	AbejaApiUrl                      string
	AbejaOrganizationID              string
	AbejaModelID                     string
	AbejaModelVersion                string
	AbejaModelVersionID              string
	AbejaDeploymentID                string
	AbejaDeploymentCodeDownload      string
	AbejaServiceID                   string
	AbejaTrainingModelDownload       string
	TrainingJobDefinitionName        string
	TrainingJobDefinitionVersion     int
	TrainingJobID                    string
	TrainingJobIDS                   string
	TensorboardID                    string
	AbejaRunID                       string
	PlatformAuthToken                string
	AbejaPlatformUserID              string
	AbejaPlatformPersonalAccessToken string
	AbejaRuntime                     string
	AbejaUserModelRoot               string
	AbejaTrainingResultDir           string
	Datasets                         string
	Input                            string
	Output                           string
	Port                             int
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func SetOptionsToEnv(options AllOptions) {
	v := reflect.Indirect(reflect.ValueOf(options))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		g := f.Interface()
		key := toSnakeCase(t.Field(i).Name)
		var val string
		if value, ok := g.(int); ok {
			val = strconv.Itoa(value)
		} else {
			val = f.String()
		}
		if val != "" && val != "0" {
			os.Setenv(strings.ToUpper(key), val)
		}
	}
}

func SetOptionsToCmdline(subCmd string, options AllOptions) {
	cmd := []string{"cmd"}
	if subCmd != "" {
		cmd = append(cmd, subCmd)
	}
	v := reflect.Indirect(reflect.ValueOf(options))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		g := f.Interface()
		key := toSnakeCase(t.Field(i).Name)
		var val string
		if value, ok := g.(int); ok {
			val = strconv.Itoa(value)
		} else {
			val = f.String()
		}
		if val != "" && val != "0" {
			option := fmt.Sprintf("--%s=%s", key, val)
			cmd = append(cmd, option)
		}
	}
	os.Args = cmd
}

func DummyRunEFunc(cmd *cobra.Command, args []string) error {
	return nil
}
