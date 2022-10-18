package util

import (
	"io/ioutil"
	"path/filepath"

	errors "golang.org/x/xerrors"
)

// unixSocketFileName specifies file name of unix domain socket for communiccation to runtime.
const unixSocketFileName = "samp_v2.sock"

func MakeUDSFilePath() (string, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", errors.Errorf(": %w", err)
	}
	filePath := filepath.Join(dir, unixSocketFileName)
	return filePath, nil
}
