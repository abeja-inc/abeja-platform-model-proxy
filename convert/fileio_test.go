package convert

import (
	"io/ioutil"
	"testing"

	"github.com/abeja-inc/platform-model-proxy/config"
)

func TestToFileFromBody(t *testing.T) {
	conf := config.NewConfiguration()
	expect := "foo,bar"
	filePath, err := ToFileFromBody(expect, ".csv", conf.RequestedDataDir)
	if err != nil {
		t.Fatal("failed to ToFileFromBody: ", err)
	}
	actual, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal("failed to file of result of ToFileFromBody: ", err)
	}
	if expect != string(actual) {
		t.Fatalf("expect [%s], but actual is [%#v]", expect, actual)
	}
}
