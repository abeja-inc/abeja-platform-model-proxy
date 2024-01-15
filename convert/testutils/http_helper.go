package testutils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	cleanutil "github.com/abeja-inc/abeja-platform-model-proxy/util/clean"
)

// helper functions

// AddFormFile adds file to multipart/form-data Request.
func AddFormFile(
	path string, contentType string, writer *multipart.Writer, name string, t *testing.T) {
	t.Helper()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, filepath.Base(path)))
	h.Set("Content-Type", contentType)
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatal("failed to write.CreateFormFile", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("file %s open failed.", path)
	}
	defer cleanutil.Close(context.TODO(), file, path)

	if _, err = io.Copy(part, file); err != nil {
		t.Fatal("failed to write part", err)
	}
}
