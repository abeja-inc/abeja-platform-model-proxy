package tensorboard

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
)

func TestRemoveDuplicates(t *testing.T) {
	testCases := []struct {
		name     string
		given    []string
		expected []string
	}{
		{
			name:     "remove duplicates",
			given:    []string{"a", "b", "c", "b", "a"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "no need to remove duplicates",
			given:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := removeDuplicates(testCase.given)
			if !reflect.DeepEqual(actual, testCase.expected) {
				t.Errorf("expected %v, but got %v", testCase.expected, actual)
			}
		})
	}
}

func TestProcessDownloadAndUnarchive(t *testing.T) {
	organizationId := "1000000000000"
	trainingJobId := "1500000000000"
	tensorboardId := "1800000000000"

	conf := &config.Configuration{
		OrganizationID: organizationId,
		TensorboardID:  tensorboardId,
		TrainingJobIDS: trainingJobId,
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	}()
	conf.MountTargetDir = tmpDir

	// suppose file exists in /mnt/tensorboards/<tensorboard_id>/training_jobs/<training_job> directory.
	targetDir := filepath.Join(conf.MountTargetDir, "tensorboards", tensorboardId, "training_jobs", trainingJobId)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(targetDir, "dummyFile"), []byte("existing!!\n"), 0644); err != nil {
		t.Fatal(err)
	}

	downloader := MockDownloader{}
	fakeUnarchive := func(filePath, destination string) error { return nil }

	err = downloadAndUnarchive(context.TODO(), conf, downloader, fakeUnarchive, trainingJobId)
	if err != nil {
		t.Errorf("expected nil, but got %v", err)
	}
}
