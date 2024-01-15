package convert

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/convert/testutils"
	"github.com/abeja-inc/abeja-platform-model-proxy/entity"
)

func TestToContentsAsSingleRequest(t *testing.T) {
	os.Args = []string{"cmd"}
	conf := config.NewConfiguration()

	cases := []struct {
		name             string
		reqContentType   string
		resContentType   string
		method           string
		body             string
		binary           string
		ext              string
		additionalHeader entity.Header
	}{
		{
			name:           "GET",
			reqContentType: "",
			resContentType: "application/x-www-form-urlencoded",
			method:         "GET",
			body:           "foo=bar&baz=qux",
			binary:         "",
			ext:            ".txt",
			additionalHeader: entity.Header{
				Key:    "X-Abeja-Model-Version",
				Values: []string{"1.0"},
			},
		}, { // when method is GET, content-type is ignored
			name:           "GET-json",
			reqContentType: "application/json",
			resContentType: "application/x-www-form-urlencoded",
			method:         "GET",
			body:           "foo=bar&baz=qux",
			binary:         "",
			ext:            ".txt",
			additionalHeader: entity.Header{
				Key:    "Set-Cookie",
				Values: []string{"cookie1", "cookie2"},
			},
		}, {
			name:             "POST-json",
			reqContentType:   "application/json",
			resContentType:   "application/json",
			method:           "POST",
			body:             "{\"foo\":\"bar\", \"baz\":\"qux\"}",
			binary:           "",
			ext:              ".json",
			additionalHeader: entity.Header{},
		}, {
			name:             "POST-text/plain",
			reqContentType:   "text/plain",
			resContentType:   "text/plain",
			method:           "POST",
			body:             "foo = bar, baz.\nqux...",
			binary:           "",
			ext:              ".txt",
			additionalHeader: entity.Header{},
		}, {
			name:             "POST-text/csv",
			reqContentType:   "text/csv",
			resContentType:   "text/csv",
			method:           "POST",
			body:             "foo,bar,baz\nqux,quux,corge",
			binary:           "",
			ext:              ".csv",
			additionalHeader: entity.Header{},
		}, {
			name:             "POST-application/x-www-form-urlencoded",
			reqContentType:   "application/x-www-form-urlencoded",
			resContentType:   "application/x-www-form-urlencoded",
			method:           "POST",
			body:             "foo=bar&baz=qux",
			binary:           "",
			ext:              ".txt",
			additionalHeader: entity.Header{},
		}, {
			name:             "POST-image/jpeg",
			reqContentType:   "image/jpeg",
			resContentType:   "image/jpeg",
			method:           "POST",
			body:             "",
			binary:           "../test_resources/cat.jpg",
			ext:              ".jpg",
			additionalHeader: entity.Header{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var body io.Reader
			url := "http://example.com"
			if c.method == "POST" {
				if c.binary == "" {
					body = bytes.NewBufferString(c.body)
				} else {
					data, err := ioutil.ReadFile(c.binary)
					if err != nil {
						t.Fatalf("failed to read from file %s: %s", c.binary, err.Error())
					}
					body = bytes.NewBuffer(data)
				}
			} else if c.method == "GET" {
				url = url + "?" + c.body
			}
			req := httptest.NewRequest(c.method, url, body)

			var expectHeaders []entity.Header
			expectHeaders = append(expectHeaders, entity.Header{
				Key:    "content-type",
				Values: []string{c.reqContentType},
			})
			req.Header.Add("Content-Type", c.reqContentType)
			if len(c.additionalHeader.Values) > 0 {
				for i := 0; i < len(c.additionalHeader.Values); i++ {
					req.Header.Add(c.additionalHeader.Key, c.additionalHeader.Values[i])
				}
				expectHeaders = append(expectHeaders, c.additionalHeader)
			}

			ml, err := ToContents(context.TODO(), req, &conf)
			if err != nil {
				t.Fatal("failed to ToContents: ", err)
			}
			if ml.Method != c.method {
				t.Errorf("ContentList.Method should be %s, but %s", c.method, ml.Method)
			}
			if ml.ContentType != c.resContentType {
				t.Errorf("ContentList.ContentType should be %s, but %s", c.resContentType, ml.ContentType)
			}
			if len(ml.Contents) != 1 {
				t.Errorf("ContentList.Content's len should be 1, but %d", len(ml.Contents))
			}
			content := ml.Contents[0]
			if filepath.Ext(*(content.Path)) != c.ext {
				t.Errorf("Content Path should has ext %s, but %s ", c.ext, *(content.Path))
			}
			contentActual, err := ioutil.ReadFile(*content.Path)
			if err != nil {
				t.Fatal("Content file read error: ", err)
			}
			if c.binary != "" {
				data, _ := ioutil.ReadFile(c.binary)
				if !bytes.Equal(contentActual, data) {
					t.Errorf("Content.Body should be %s, but different", c.binary)
				}
			} else {
				if string(contentActual) != c.body {
					t.Errorf("Content.Body should be %s, but %s", c.body, string(contentActual))
				}
			}
			if len(expectHeaders) != len(ml.Headers) {
				t.Errorf("headers length should be %d, but %d", len(expectHeaders), len(ml.Headers))
			}
			for i := 0; i < len(expectHeaders); i++ {
				if strings.ToLower(expectHeaders[i].Key) != ml.Headers[i].Key {
					t.Errorf("Header[%d]'s key should be %s, but %s",
						i, expectHeaders[i].Key, ml.Headers[i].Key)
				}
				for j := 0; j < len(expectHeaders[i].Values); j++ {
					if expectHeaders[i].Values[j] != ml.Headers[i].Values[j] {
						t.Errorf("Header[%d].Values[%d] should be %s, but %s",
							i, j, expectHeaders[i].Values[j], ml.Headers[i].Values[j])
					}
				}
			}
		})
	}
}

func TestToContentsAsMultipartRequest(t *testing.T) {
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

	cl, err := ToContents(context.TODO(), req, &conf)
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
