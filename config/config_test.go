package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	pathutil "github.com/abeja-inc/abeja-platform-model-proxy/util/path"
)

func TestGetListenAddress(t *testing.T) {
	conf := NewConfiguration()
	conf.Port = 5000
	actual := conf.GetListenAddress()
	if actual != ":5000" {
		t.Errorf("listen-address should be %s, but %s", ":5000", actual)
	}
}

func TestGetWorkingDir(t *testing.T) {
	tempDir, rmFunc := testTempDir(t)
	defer rmFunc()
	defer testChdir(t, tempDir)()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name          string
		userModelRoot string
		expect        string
	}{
		{
			name:          "default",
			userModelRoot: "",
			expect:        pwd,
		}, {
			name:          "relative path",
			userModelRoot: "foo",
			expect:        filepath.Join(pwd, "foo"),
		}, {
			name:          "absolute path",
			userModelRoot: "/tmp",
			expect:        "/tmp",
		}, {
			name:          "current",
			userModelRoot: ".",
			expect:        pwd,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := NewConfiguration()
			conf.UserModelRoot = c.userModelRoot
			actual, err := conf.GetWorkingDir()
			if err != nil {
				t.Error("unexpected error occurred:", err)
			}
			if actual != c.expect {
				t.Errorf("working dir should be %s, but %s", c.expect, actual)
			}
		})
	}
}

func TestGetTrainingResultDir(t *testing.T) {
	tempDir, rmFunc := testTempDir(t)
	defer rmFunc()
	defer testChdir(t, tempDir)()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name              string
		trainingResultDir string
		expect            string
	}{
		{
			name:              "default",
			trainingResultDir: "",
			expect:            filepath.Join(pwd, pathutil.DefaultTrainingResultDir),
		}, {
			name:              "relative path",
			trainingResultDir: "foo",
			expect:            filepath.Join(pwd, "foo"),
		}, {
			name:              "absolute path",
			trainingResultDir: "/tmp",
			expect:            "/tmp",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := NewConfiguration()
			conf.TrainingResultDir = c.trainingResultDir
			actual, err := conf.GetTrainingResultDir()
			if err != nil {
				t.Error("unexpected error occurred:", err)
			}
			if actual != c.expect {
				t.Errorf("training-result dir should be %s, but %s", c.expect, actual)
			}
		})
	}
}

// helper

func testTempDir(t *testing.T) (string, func()) {
	t.Helper()
	tempPath, err := ioutil.TempDir("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	return tempPath, func() {
		if err := os.RemoveAll(tempPath); err != nil {
			t.Errorf("failed to remove temporary directory: %s, error: %s", tempPath, err.Error())
		}
	}
}

func testChdir(t *testing.T, tempPath string) func() {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempPath); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Chdir(current); err != nil {
			t.Fatalf("failed to pop to %s, error: %s", current, err.Error())
		}
	}
}
