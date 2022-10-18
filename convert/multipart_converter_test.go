package convert

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/convert/testutils"
)

func TestIsTarget_Multipart(t *testing.T) {
	conv := multipartConverter{}
	cases := []struct {
		name        string
		contentType string
		method      string
		isTarget    bool
	}{
		{
			name:        "GET",
			contentType: "",
			method:      "GET",
			isTarget:    false,
		}, {
			name:        "application/json",
			contentType: "application/json",
			method:      "POST",
			isTarget:    false,
		}, {
			name:        "application/x-www-form-urlencoded",
			contentType: "text/plain",
			method:      "POST",
			isTarget:    false,
		}, {
			name:        "multipart/form-data",
			contentType: "multipart/form-data",
			method:      "POST",
			isTarget:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(c.method, "http://example.com", nil)
			req.Header.Add("Content-Type", c.contentType)

			if conv.IsTarget(context.TODO(), req.Method, req.Header.Get("Content-Type")) != c.isTarget {
				t.Errorf("IsTarget failed, method = %s, content-type = %s, expect = %t",
					c.method, c.contentType, c.isTarget)
			}
		})
	}
}

func TestToContent_Multipart(t *testing.T) {
	conv := multipartConverter{}
	conf := config.NewConfiguration()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	testutils.AddFormFile("../test_resources/cat.jpg", "image/jpeg", writer, "file1", t)
	_ = writer.WriteField("foo", "bar")
	_ = writer.WriteField("baz", "qux")
	if err := writer.Close(); err != nil {
		t.Fatal("failed to close writer", err)
	}

	req := httptest.NewRequest("POST", "http://example.com", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	cl, err := conv.ToContent(context.TODO(), req, &conf)
	if err != nil {
		t.Fatal("failed to ToContent: ", err)
	}
	if cl.Method != "POST" {
		t.Errorf("ContentList.Method should be POST, but %s", cl.Method)
	}
	if strings.HasPrefix(cl.ContentType, "multipart/form-data") != true {
		t.Errorf("ContentList.ContentType should be multipart/form-data, but %s", cl.ContentType)
	}
	if len(cl.Contents) != 3 {
		t.Errorf("ContentList.Content's len should be 3, but %d", len(cl.Contents))
	}
	content0 := cl.Contents[0]
	if *(content0.ContentType) != "image/jpeg" {
		t.Errorf("Content.ContentType should be image/jpeg, but %s", *(content0.ContentType))
	}
	if *(content0.FileName) != "cat.jpg" {
		t.Errorf("Content.FileName should be cat.jpg, but %s", *(content0.FileName))
	}
	if *(content0.FormName) != "file1" {
		t.Errorf("Content.FormName should be file, but %s", *(content0.FormName))
	}
	content0Actual, err := ioutil.ReadFile(*content0.Path)
	if err != nil {
		t.Fatal("Content file read error:", err)
	}
	expectData, err := ioutil.ReadFile("../test_resources/cat.jpg")
	if err != nil {
		t.Fatal("Original file read error:", err)
	}
	if !bytes.Equal(expectData, content0Actual) {
		t.Error("Content.Body should be cat.jpg, but different")
	}
	content1 := cl.Contents[1]
	if content1.ContentType != (*string)(nil) {
		t.Errorf("Content.ContentType should be nil, but %s", *(content1.ContentType))
	}
	if content1.FileName != (*string)(nil) {
		t.Errorf("Content.FileName should be nil, but %s", *(content1.FileName))
	}
	if *(content1.FormName) != "foo" {
		t.Errorf("Content.FormName should be foo, but %s", *(content1.FormName))
	}
	content1Actual, err := ioutil.ReadFile(*content1.Path)
	if err != nil {
		t.Fatal("Content file read error:", err)
	}
	if "bar" != string(content1Actual) {
		t.Errorf("Content.Body should be bar, but %s", string(content1Actual))
	}
	content2 := cl.Contents[2]
	if content2.ContentType != (*string)(nil) {
		t.Errorf("Content.ContentType should be nil, but %s", *(content2.ContentType))
	}
	if content2.FileName != (*string)(nil) {
		t.Errorf("Content.FileName should be nil, but %s", *(content2.FileName))
	}
	if *(content2.FormName) != "baz" {
		t.Errorf("Content.FormName should be baz, but %s", *(content2.FormName))
	}
	content2Actual, err := ioutil.ReadFile(*content2.Path)
	if err != nil {
		t.Fatal("Content file read error:", err)
	}
	if "qux" != string(content2Actual) {
		t.Errorf("Content.Body should be qux, but %s", string(content2Actual))
	}
}
