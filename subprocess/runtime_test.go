package subprocess

import (
	"testing"

	"github.com/abeja-inc/platform-model-proxy/config"
)

func TestServiceRuntimeLanguageError(t *testing.T) {
	conf := &config.Configuration{}
	conf.Runtime = "invalid_language"
	_, err := CreateServiceRuntime(conf, "/path/to/uds.sock", "/path/to/tr")
	if err == nil {
		t.Errorf("`CreateRuntime(invalid_language)` should be return err")
	}
}

func TestTrainingRuntimeLanguageError(t *testing.T) {
	conf := &config.Configuration{
		Runtime: "invalid_language",
	}
	_, err := CreateTrainRuntime(conf, "/path/to/runtime/base", "/path/to/training/result")
	if err == nil {
		t.Errorf("`CreateRuntime(invalid_language)` should be return err")
	}
}
