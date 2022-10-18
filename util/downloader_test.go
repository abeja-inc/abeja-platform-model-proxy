package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/abeja-inc/platform-model-proxy/util/auth"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	httputil "github.com/abeja-inc/platform-model-proxy/util/http"
)

const validToken = "valid_token"
const invalidToken = "invalid_token"
const validDownloadURL = "valid_download_url"
const invalidDownloadURL = "invalid_download_url"
const modelCode = "def handler():\n    return \"hello world\"\n"

type SourceResJSON struct {
	DownloadURL string `json:"download_uri"`
}

func (s *SourceResJSON) GetDownloadURL() string {
	return s.DownloadURL
}

func (_ *SourceResJSON) GetContentType() string {
	return ""
}

type TrainingJobResJSON struct {
	Artifacts struct {
		Complete struct {
			DownloadURL string `json:"uri"`
		} `json:"complete"`
	} `json:"artifacts"`
}

func (t *TrainingJobResJSON) GetDownloadURL() string {
	return t.Artifacts.Complete.DownloadURL
}

func (_ *TrainingJobResJSON) GetContentType() string {
	return ""
}

type FileInfoJSON struct {
	DownloadURL string `json:"download_uri"`
	ContentType string `json:"content_type"`
}

func (f *FileInfoJSON) GetDownloadURL() string {
	return f.DownloadURL
}

func (f *FileInfoJSON) GetContentType() string {
	return f.ContentType
}

func TestDownload(t *testing.T) {
	cases := []struct {
		name               string
		token              string
		client             *http.Client
		expectHasError     bool
		expectErrorMessage string
		decoderRes         DecoderRes
	}{
		{
			name:               "normal case",
			token:              validToken,
			client:             clientWithSourceResJSON(t, 200, validDownloadURL),
			expectHasError:     false,
			expectErrorMessage: "",
			decoderRes:         new(SourceResJSON),
		},
		{
			name:               "invalid token",
			token:              invalidToken,
			client:             clientWithSourceResJSON(t, 200, validDownloadURL),
			expectHasError:     true,
			expectErrorMessage: "failed to request to source: response error with StatusCode: 401",
			decoderRes:         new(SourceResJSON),
		},
		{
			name:               "invalid download url",
			token:              validToken,
			client:             clientWithSourceResJSON(t, 200, invalidDownloadURL),
			expectHasError:     true,
			expectErrorMessage: "failed to download from invalid_download_url with status 404, body = []",
			decoderRes:         new(SourceResJSON),
		},
		{
			name:               "normal case",
			token:              validToken,
			client:             clientWithTrainingJobResJSON(t, 200, validDownloadURL),
			expectHasError:     false,
			expectErrorMessage: "",
			decoderRes:         new(TrainingJobResJSON),
		},
		{
			name:               "invalid token",
			token:              invalidToken,
			client:             clientWithTrainingJobResJSON(t, 200, validDownloadURL),
			expectHasError:     true,
			expectErrorMessage: "failed to request to source: response error with StatusCode: 401",
			decoderRes:         new(TrainingJobResJSON),
		},
		{
			name:               "invalid download url",
			token:              validToken,
			client:             clientWithTrainingJobResJSON(t, 200, invalidDownloadURL),
			expectHasError:     true,
			expectErrorMessage: "failed to download from invalid_download_url with status 404, body = []",
			decoderRes:         new(TrainingJobResJSON),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			httpClient := c.client
			authInfo := auth.AuthInfo{
				AuthToken: c.token,
			}
			downloader, err := NewDownloader("http://localhost", authInfo, httpClient)
			if err != nil {
				t.Fatal("failed to NewDownloader: ", err)
			}
			tempfile, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatal("failed to create tempfile:", err)
			}
			filePath := tempfile.Name()
			if err := tempfile.Close(); err != nil {
				t.Fatal("Error when closing file:", err)
			}
			defer cleanutil.Remove(context.TODO(), filePath)
			if _, err = downloader.Download("source", filePath, c.decoderRes); c.expectHasError == true {
				if err == nil {
					t.Fatal("Download() should be raise error")
				}
				if err.Error() != c.expectErrorMessage {
					t.Fatalf("error message should be [%s], but [%s]", c.expectErrorMessage, err.Error())
				}
				return
			}
			actual, err := ioutil.ReadFile(filePath)
			if err != nil {
				t.Fatal("failed to read download file: ", err)
			}
			if string(actual) != modelCode {
				t.Fatalf("[%s] should be equals [%s]", string(actual), modelCode)
			}
		})
	}
}

func TestDownloadWithContentType(t *testing.T) {
	cases := []struct {
		name               string
		token              string
		client             *http.Client
		expectHasError     bool
		expectErrorMessage string
		expectContentType  string
		decoderRes         DecoderRes
	}{
		{
			name:               "normal case",
			token:              validToken,
			client:             clientWithFileInfoJSON(t, 200, validDownloadURL, "application/json"),
			expectHasError:     false,
			expectErrorMessage: "",
			expectContentType:  "application/json",
			decoderRes:         new(FileInfoJSON),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			httpClient := c.client
			authInfo := auth.AuthInfo{
				AuthToken: c.token,
			}
			downloader, err := NewDownloader("http://localhost", authInfo, httpClient)
			if err != nil {
				t.Fatal("failed to NewDownloader: ", err)
			}
			tempfile, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatal("failed to create tempfile:", err)
			}
			filePath := tempfile.Name()
			if err := tempfile.Close(); err != nil {
				t.Fatal("Error when closing file:", err)
			}
			defer cleanutil.Remove(context.TODO(), filePath)
			contentType, err := downloader.Download("source", filePath, c.decoderRes)
			if c.expectHasError == true {
				if err == nil {
					t.Fatal("Download() should be raise error")
					return
				}
				if err.Error() != c.expectErrorMessage {
					t.Fatalf("error message should be [%s], but [%s]", c.expectErrorMessage, err.Error())
				}
				return
			}
			actual, err := ioutil.ReadFile(filePath)
			if err != nil {
				t.Fatal("failed to read download file: ", err)
			}
			if string(actual) != modelCode {
				t.Fatalf("[%s] should be equals [%s]", string(actual), modelCode)
			}
			if contentType != c.expectContentType {
				t.Errorf("contentType should be [%s], but [%s]", c.expectContentType, contentType)
			}
		})
	}
}

func TestDownloadWithTemporary5XXError(t *testing.T) {
	const datalakeGetFileInfoPath = "/channels/1111111111111/20200101T000000-12345678-90ab-cdef-1234-567890abcdef"
	httpClient := httputil.GetMockHTTPClient(t, []httputil.ResponseMock{
		{
			Path:           datalakeGetFileInfoPath,
			NeedAuth:       true,
			AuthToken:      validToken,
			StatusCode:     http.StatusInternalServerError,
			ResponseBody:   "",
			ResponseHeader: make(http.Header),
		},
		{
			// check retry
			Path:           datalakeGetFileInfoPath,
			NeedAuth:       true,
			AuthToken:      validToken,
			StatusCode:     http.StatusOK,
			ResponseBody:   `{"download_uri": "https://localhost/download", "content_type": "plain/text"}`,
			ResponseHeader: make(http.Header),
		},
		{
			Path:           "/download",
			NeedAuth:       false,
			AuthToken:      "",
			StatusCode:     http.StatusOK,
			ResponseBody:   modelCode,
			ResponseHeader: make(http.Header),
		},
	})
	authInfo := auth.AuthInfo{
		AuthToken: validToken,
	}
	downloader, err := NewDownloader("http://localhost", authInfo, httpClient)
	if err != nil {
		t.Fatal("failed to NewDownloader: ", err)
	}
	tempfile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("failed to create tempfile:", err)
	}
	filePath := tempfile.Name()
	if err := tempfile.Close(); err != nil {
		t.Fatal("Error when closing file:", err)
	}
	defer cleanutil.Remove(context.TODO(), filePath)

	decoderRes := new(FileInfoJSON)
	contentType, err := downloader.Download(datalakeGetFileInfoPath, filePath, decoderRes)
	if err != nil {
		t.Fatal("failed to download:", err)
	}
	if contentType != "plain/text" {
		t.Errorf("contentType should be %s, but %s", "plain/text", contentType)
	}
	actual, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal("failed to read download file: ", err)
	}
	if string(actual) != modelCode {
		t.Fatalf("[%s] should be equals [%s]", string(actual), modelCode)
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func clientWithSourceResJSON(t *testing.T, status int, downloadURL string) *http.Client {
	t.Helper()

	// response for get-source request
	body := SourceResJSON{
		DownloadURL: downloadURL,
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return client(t, status, b)
}

func clientWithTrainingJobResJSON(t *testing.T, status int, downloadURL string) *http.Client {
	t.Helper()

	// response for get-source request
	body := new(TrainingJobResJSON)
	body.Artifacts.Complete.DownloadURL = downloadURL
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return client(t, status, b)
}

func clientWithFileInfoJSON(t *testing.T, status int, downloadURL, contentType string) *http.Client {
	t.Helper()

	// response for get-source request
	body := FileInfoJSON{
		DownloadURL: downloadURL,
		ContentType: contentType,
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return client(t, status, b)
}

func client(t *testing.T, status int, body []byte) *http.Client {
	t.Helper()

	return NewTestClient(func(req *http.Request) *http.Response {
		// response for presigned-url
		if req.URL.Path == validDownloadURL {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(modelCode))),
				Header:     make(http.Header),
			}
		}
		if req.URL.Path == invalidDownloadURL {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(""))),
				Header:     make(http.Header),
			}
		}

		// authorization check
		authToken := req.Header.Get("Authorization")
		if authToken != "" {
			if req.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", validToken) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(""))),
					Header:     make(http.Header),
				}
			}
		}

		headers := make(http.Header)
		headers.Set("Content-Type", "application/json")
		return &http.Response{
			StatusCode: status,
			Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
			Header:     headers,
		}
	})
}
