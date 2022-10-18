package convert

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	errors "golang.org/x/xerrors"
)

// ToFileFromBody returns file-path that stored `body`.
func ToFileFromBody(body string, ext string, dataDir string) (string, error) {
	in := bytes.NewBufferString(body)
	return ToFileFromReader(in, ext, dataDir)
}

// ToFileFromReader return file-path that stored `in`.
func ToFileFromReader(in io.Reader, name string, dataDir string) (string, error) {
	fileName := buildFileName(name)
	filePath := filepath.Join(dataDir, fileName)

	fp, err := createFile(filePath)
	if err != nil {
		return "", errors.Errorf(": %w", err)
	}
	defer func() {
		if ferr := fp.Close(); ferr != nil {
			err = errors.Errorf(": %w", ferr)
		}
	}()

	_, err = io.Copy(fp, in)
	return filePath, err
}

// FromFile returns `os.File` from `filePath`.
func FromFile(filePath string) (*os.File, error) {
	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}
	return os.Open(filePath)
}

func buildFileName(name string) string {
	t := time.Now()
	nano := t.Nanosecond()
	mili := nano / 1000
	tstr := time.Now().Format("20060102150405") + strconv.Itoa(mili)
	var sb strings.Builder
	sb.Grow(len(tstr) + len(name) + 1)
	sb.WriteString(tstr)
	if !strings.HasPrefix(name, ".") {
		sb.WriteString("_")
	}
	sb.WriteString(name)
	return sb.String()
}

func createFile(filePath string) (*os.File, error) {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return nil, errors.Errorf(": %w", err)
		}
	}
	fp, err := os.Create(filePath)
	if err != nil {
		return nil, errors.Errorf(": %w", err)
	}
	return fp, nil
}
