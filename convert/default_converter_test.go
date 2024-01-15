package convert

import (
	"bytes"
	"context"
	_ "encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/entity"
	cleanutil "github.com/abeja-inc/abeja-platform-model-proxy/util/clean"
	"github.com/abeja-inc/abeja-platform-model-proxy/version"
)

func TestIsTarget_Default(t *testing.T) {
	conv := defaultConverter{}
	cases := []struct {
		name        string
		contentType string
		method      string
		isTarget    bool
	}{
		{name: "GET", contentType: "", method: "GET", isTarget: true},
		{name: "GET-json", contentType: "application/json", method: "GET", isTarget: true}, // when method is GET, content-type is ignored
		{name: "POST-json", contentType: "application/json", method: "POST", isTarget: true},
		{name: "POST-text/plain", contentType: "text/plain", method: "POST", isTarget: true},
		{name: "POST-text/csv", contentType: "text/csv", method: "POST", isTarget: true},
		{name: "POST-application/x-www-form-urlencoded", contentType: "application/x-www-form-urlencoded", method: "POST", isTarget: true},
		{name: "POST-image/jpeg", contentType: "image/jpeg", method: "POST", isTarget: true},    // not text data
		{name: "PATCH-image/jpeg", contentType: "image/jpeg", method: "PATCH", isTarget: false}, // only GET/POST/PUT are allowed
		{name: "POST-multipart", contentType: "multipart/form-data", method: "POST", isTarget: false},
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

func TestToContent_Default(t *testing.T) {
	conv := defaultConverter{}
	conf := config.NewConfiguration()
	cases := []struct {
		name           string
		reqContentType string
		resContentType string
		method         string
		body           string
		binary         string
		ext            string
	}{
		{
			name:           "GET",
			reqContentType: "",
			resContentType: "application/x-www-form-urlencoded",
			method:         "GET",
			body:           "foo=bar&baz=qux",
			binary:         "",
			ext:            ".txt",
		}, { // when method is GET, content-type is ignored
			name:           "GET-json",
			reqContentType: "application/json",
			resContentType: "application/x-www-form-urlencoded",
			method:         "GET",
			body:           "foo=bar&baz=qux",
			binary:         "",
			ext:            ".txt",
		}, {
			name:           "POST-json",
			reqContentType: "application/json",
			resContentType: "application/json",
			method:         "POST",
			body:           "{\"foo\":\"bar\", \"baz\":\"qux\"}",
			binary:         "",
			ext:            ".json",
		}, {
			name:           "POST-text/plain",
			reqContentType: "text/plain",
			resContentType: "text/plain",
			method:         "POST",
			body:           "foo = bar, baz.\nqux...",
			binary:         "",
			ext:            ".txt",
		}, {
			name:           "POST-text/csv",
			reqContentType: "text/csv",
			resContentType: "text/csv",
			method:         "POST",
			body:           "foo,bar,baz\nqux,quux,corge",
			binary:         "",
			ext:            ".csv",
		}, {
			name:           "POST-application/x-www-form-urlencoded",
			reqContentType: "application/x-www-form-urlencoded",
			resContentType: "application/x-www-form-urlencoded",
			method:         "POST",
			body:           "foo=bar&baz=qux",
			binary:         "",
			ext:            ".txt",
		}, {
			name:           "POST-image/jpeg",
			reqContentType: "image/jpeg",
			resContentType: "image/jpeg",
			method:         "POST",
			body:           "",
			binary:         "../test_resources/cat.jpg",
			ext:            ".jpg",
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
			req.Header.Add("Content-Type", c.reqContentType)

			cl, err := conv.ToContent(context.TODO(), req, &conf)
			if err != nil {
				t.Fatal("failed to ToContent: ", err)
			}
			if cl.Method != c.method {
				t.Errorf("ContentList.Method should be %s, but %s", c.method, cl.Method)
			}
			if cl.ContentType != c.resContentType {
				t.Errorf("ContentList.ContentType should be %s, but %s", c.resContentType, cl.ContentType)
			}
			if len(cl.Contents) != 1 {
				t.Errorf("ContentList.Content's len should be 1, but %d", len(cl.Contents))
			}
			content := cl.Contents[0]
			if content.ContentType != (*string)(nil) {
				t.Errorf("Content.ContentType should be %s, but %s", c.resContentType, *(content.ContentType))
			}
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
			// data, _ := json.Marshal(cl)
			// t.Logf("data = %s\n", string(data))
		})
	}
}

func TestFromResponseWithEmptyResponse_Default(t *testing.T) {
	conv := defaultConverter{}
	res := entity.Response{}
	status, headers, body, err := conv.FromResponse(context.TODO(), res)
	if err != nil {
		t.Errorf("err should be nil, but %s", err.Error())
	}
	if status != http.StatusOK {
		t.Errorf("status_code should be 200, but %d", status)
	}
	if len(headers) != 4 {
		t.Errorf("headers length should be 4, but %d", len(headers))
	}
	aval, ok := headers[KeyContentType]
	if ok != true {
		t.Errorf("headers should contain key: %s", KeyContentType)
	} else {
		if aval != "application/json" {
			t.Errorf("content_type should be `application/json`, but %s", headers[KeyContentType])
		}
	}
	aval, ok = headers[KeyContentLength]
	if ok != true {
		t.Errorf("headers should contain key: %s", KeyContentLength)
	} else {
		if aval != "0" {
			t.Errorf("content_length should be `0`, but %s", headers[KeyContentLength])
		}
	}
	connection, ok := headers[KeyConnection]
	if ok != true {
		t.Errorf("headers should contain key: %s", KeyConnection)
	} else {
		if connection != "close" {
			t.Errorf("connection should be `close`, but %s", connection)
		}
	}
	aval, ok = headers[KeyAbejaProxyVersion]
	if ok != true {
		t.Errorf("headers should contain key: %s", KeyAbejaProxyVersion)
	} else {
		if aval != version.Version {
			t.Errorf("value of key[%s] should %s, but %s",
				KeyAbejaProxyVersion, version.Version, aval)
		}
	}
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		t.Error("failed to read response file: ", err)
	}
	if string(buf) != "" {
		t.Errorf("body should be empty, but %s", string(buf))
	}
}

func TestFromResponseWithResponse_Default(t *testing.T) {
	conv := defaultConverter{}

	cases := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		metadata    map[string]string
	}{
		{
			name:        "json",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body:        "{\"foo\":\"bar\",\"hoge\":\"fuga\"}",
			metadata:    map[string]string{},
		},
		{
			name:        "csv",
			statusCode:  http.StatusOK,
			contentType: "text/csv",
			body:        "foo,bar\nbaz,qux",
			metadata:    map[string]string{"foo": "bar", "hoge": "fuga"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fp, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatal("failed to create temp file", err)
			}
			fpath := fp.Name()
			cleanutil.Close(context.TODO(), fp, fpath)
			if err := ioutil.WriteFile(fpath, []byte(c.body), 0644); err != nil {
				t.Fatal("Error when writing to file:", err)
			}
			res := entity.Response{
				ContentType: &c.contentType,
				Path:        &fpath,
				StatusCode:  &c.statusCode,
			}
			if len(c.metadata) > 0 {
				res.Metadata = &c.metadata
			}

			status, headers, body, err := conv.FromResponse(context.TODO(), res)
			if err != nil {
				t.Errorf("err should be nil, but %s", err.Error())
			}
			if status != c.statusCode {
				t.Errorf("status_code should be %d, but %d", c.statusCode, status)
			}
			contentType, ok := headers[KeyContentType]
			if ok != true {
				t.Errorf("headers should contain key: %s", KeyContentType)
			} else {
				if contentType != c.contentType {
					t.Errorf("content_type should be `%s`, but %s", c.contentType, contentType)
				}
			}
			contentLength, ok := headers[KeyContentLength]
			if ok != true {
				t.Errorf("headers should contain key: %s", KeyContentLength)
			} else {
				actualLength := strconv.Itoa(len(c.body))
				if contentLength != actualLength {
					t.Errorf("content-length should be %s, but %s", actualLength, contentLength)
				}
			}
			connection, ok := headers[KeyConnection]
			if ok != true {
				t.Errorf("headers should contain key: %s", KeyConnection)
			} else {
				if connection != "close" {
					t.Errorf("connection should be `close`, but %s", connection)
				}
			}
			if (len(c.metadata) + 4) != len(headers) {
				t.Errorf("headers length should be %d, but %d",
					(len(c.metadata) + 4), len(headers))
			}
			if len(c.metadata) > 0 {
				for ekey, eval := range c.metadata {
					aval, ok := headers[ekey]
					if ok != true {
						t.Errorf("headers should contain key: %s", ekey)
					} else {
						if eval != aval {
							t.Errorf("value of key[%s] should %s, but %s", ekey, eval, aval)
						}
					}
				}
			}
			aval, ok := headers[KeyAbejaProxyVersion]
			if ok != true {
				t.Errorf("headers should contain key: %s", KeyAbejaProxyVersion)
			} else {
				if aval != version.Version {
					t.Errorf("value of key[%s] should %s, but %s",
						KeyAbejaProxyVersion, version.Version, aval)
				}
			}
			buf, err := ioutil.ReadAll(body)
			if err != nil {
				t.Error("failed to read response file: ", err)
			}
			if string(buf) != c.body {
				t.Errorf("body should be [%s], but %s", c.body, string(buf))
			}
		})
	}
}
