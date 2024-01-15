package util

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	cleanutil "github.com/abeja-inc/abeja-platform-model-proxy/util/clean"
)

func TestUnarchive(t *testing.T) {
	dir, err := ioutil.TempDir("", "testUnarchive")
	if err != nil {
		t.Fatal("failed to making temporary directory for test: ", err)
	}
	defer cleanutil.RemoveAll(context.TODO(), dir)

	if err = Unarchive("../test_resources/test_archive.tgz", dir); err != nil {
		t.Fatal("failed to unarchive: ", err)
	}
	filePath := filepath.Join(dir, "main.py")
	if _, err = os.Stat(filePath); err != nil {
		t.Fatal("file that unarchived not found: ", err)
	}

	expect := "def handler():\n    return \"hello world\"\n"
	actual, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal("failed to read unarchived file: ", err)
	}
	if string(actual) != expect {
		t.Fatalf("[%s] should be equals [%s]", string(actual), expect)
	}
}
