package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/abeja-inc/abeja-platform-model-proxy/util/auth"
	pathutil "github.com/abeja-inc/abeja-platform-model-proxy/util/path"
)

const DefaultAbejaAPIURL = "https://api.abeja.io"
const DefaultHTTPListenPort = 5000
const DefaultHealthCheckListenPort = 5001
const DefaultRuntime = "python36"

const DefaultMountTargetDir = "/mnt"

var requestedDataDir string

func init() {
	tempDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		requestedDataDir = "/tmp"
	} else {
		requestedDataDir = tempDir
	}
}

// Configuration represents options specified by command-line parameters or environment variables.
type Configuration struct {
	APIURL                       string
	DeploymentID                 string
	ModelID                      string
	ModelVersion                 string
	ModelVersionID               string
	OrganizationID               string
	ServiceID                    string
	DeploymentCodeDownload       string
	TrainingModelDownload        string
	UserModelRoot                string
	PlatformAuthToken            string
	PlatformUserID               string
	PlatformPersonalAccessToken  string
	TrainingJobID                string
	TrainingJobIDS               string
	TrainingJobDefinitionName    string
	TrainingJobDefinitionVersion int
	TensorboardID                string
	MountTargetDir               string
	RunID                        string
	Runtime                      string
	RequestedDataDir             string
	Port                         int
	HealthCheckPort              int
	TrainingResultDir            string
	Input                        string
	Output                       string
}

func NewConfiguration() Configuration {
	conf := Configuration{}
	conf.RequestedDataDir = requestedDataDir
	return conf
}

// GetListenAddress returns the address and port number on which the web server listens.
func (config *Configuration) GetListenAddress() string {
	return fmt.Sprintf(":%d", config.Port)
}

func (config *Configuration) GetHealthCheckAddress() string {
	return fmt.Sprintf(":%d", config.HealthCheckPort)
}

func (config *Configuration) GetWorkingDir() (string, error) {
	return pathutil.GetWorkingDir(config.UserModelRoot)
}

func (config *Configuration) GetTrainingResultDir() (string, error) {
	return pathutil.GetTrainingResultDir(config.TrainingResultDir, config.UserModelRoot)
}

func (config *Configuration) GetAuthInfo() auth.AuthInfo {
	var userID string
	if config.PlatformUserID != "" && !strings.HasPrefix(config.PlatformUserID, "user-") {
		userID = "user-" + config.PlatformUserID
	} else {
		userID = config.PlatformUserID
	}
	return auth.AuthInfo{
		AuthToken:     config.PlatformAuthToken,
		UserID:        userID,
		PersonalToken: config.PlatformPersonalAccessToken,
	}
}

func (config *Configuration) String() string {
	var ret bytes.Buffer
	ret.WriteString("{")
	v := reflect.Indirect(reflect.ValueOf(config))
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		if i > 0 {
			ret.WriteString(", ")
		}
		field := t.Field(i).Name
		var value string
		f := v.Field(i)
		iface := f.Interface()
		if val, ok := iface.(int); ok {
			value = strconv.Itoa(val)
		} else {
			value = f.String()
		}
		if field == "PlatformAuthToken" || field == "PlatformPersonalAccessToken" {
			value = "xxxxxxxxxx"
		}
		ret.WriteString(fmt.Sprintf("%s: %s", field, value))
	}
	ret.WriteString("}")
	return ret.String()
}
