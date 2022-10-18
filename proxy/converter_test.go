package proxy

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/abeja-inc/platform-model-proxy/config"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	httputil "github.com/abeja-inc/platform-model-proxy/util/http"
)

func TestFromInput(t *testing.T) {

	cases := []struct {
		name                string
		conf                *config.Configuration
		httpClient          *http.Client
		expectHasError      bool
		errMsg              string
		expectContentLength int
		expectContentType   string
		expectMetadata      map[string]interface{}
		expectBody          string
	}{
		{
			name: "normal datalake file input",
			conf: &config.Configuration{
				APIURL:            "http://localhost",
				PlatformAuthToken: "valid_token",
				Input:             "{\"$datalake:1\": \"1111111111111/file\"}",
			},
			httpClient: httputil.GetMockHTTPClient(t, []httputil.ResponseMock{
				{
					Path:           "/channels/1111111111111/file",
					NeedAuth:       true,
					AuthToken:      "valid_token",
					StatusCode:     http.StatusOK,
					ResponseBody:   "{\"download_url\": \"https://localhost/download\", \"content_type\": \"application/json\"}",
					ResponseHeader: make(http.Header),
				},
				{
					Path:           "/download",
					NeedAuth:       false,
					AuthToken:      "",
					StatusCode:     http.StatusOK,
					ResponseBody:   "{\"foo\": \"bar\"}",
					ResponseHeader: make(http.Header),
				},
			}),
			expectHasError:      false,
			errMsg:              "",
			expectContentLength: 1,
			expectContentType:   "application/json",
			expectMetadata:      nil,
			expectBody:          "{\"foo\": \"bar\"}",
		}, {
			name: "normal datalake file input w/ metadata",
			conf: &config.Configuration{
				APIURL:            "http://localhost",
				PlatformAuthToken: "valid_token",
				Input:             "{\"$datalake:1\": \"1111111111111/file\"}",
			},
			httpClient: httputil.GetMockHTTPClient(t, []httputil.ResponseMock{
				{
					Path:       "/channels/1111111111111/file",
					NeedAuth:   true,
					AuthToken:  "valid_token",
					StatusCode: http.StatusOK,
					ResponseBody: `{
					                 "download_url": "https://localhost/download",
					                 "content_type": "application/json",
									 "metadata": {
									   "x-abeja-meta-filename": "test.csv",
                                       "x-abeja-sys-meta-validation-status": "FAILURE",
                                       "x-abeja-sys-meta-validation-schema-version": "2",
                                       "x-abeja-sys-meta-validation-schema-id": "1466430869682",
                                       "x-abeja-sys-meta-validation-error": [
                                         {
                                           "message": "duplicate key value violates unique constraint \"sample_schema_pkey\"",
                                           "error": "invalid_record"
                                         }
                                       ]
									 }
								   }`,
					ResponseHeader: make(http.Header),
				},
				{
					Path:           "/download",
					NeedAuth:       false,
					AuthToken:      "",
					StatusCode:     http.StatusOK,
					ResponseBody:   "{\"foo\": \"bar\"}",
					ResponseHeader: make(http.Header),
				},
			}),
			expectHasError:      false,
			errMsg:              "",
			expectContentLength: 1,
			expectContentType:   "application/json",
			expectMetadata: map[string]interface{}{
				"x-abeja-meta-filename":                      "test.csv",
				"x-abeja-sys-meta-validation-status":         "FAILURE",
				"x-abeja-sys-meta-validation-schema-version": "2",
				"x-abeja-sys-meta-validation-schema-id":      "1466430869682",
				"x-abeja-sys-meta-validation-error": []map[string]interface{}{
					{
						"message": "duplicate key value violates unique constraint \"sample_schema_pkey\"",
						"error":   "invalid_record",
					},
				},
			},
			expectBody: "{\"foo\": \"bar\"}",
		}, {
			name: "normal json input",
			conf: &config.Configuration{
				Input: "{\"foo\": \"bar\"}",
			},
			httpClient:          nil,
			expectHasError:      false,
			errMsg:              "",
			expectContentLength: 1,
			expectContentType:   "application/json",
			expectMetadata:      nil,
			expectBody:          "{\"foo\": \"bar\"}",
		}, {
			name:                "no input",
			conf:                &config.Configuration{},
			httpClient:          nil,
			expectHasError:      false,
			errMsg:              "",
			expectContentLength: 0,
			expectContentType:   "",
			expectMetadata:      nil,
			expectBody:          "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := FromInput(context.TODO(), c.conf, c.httpClient)
			if err != nil {
				if c.expectHasError {
					if err.Error() != c.errMsg {
						t.Fatalf("Error message should be [%s], but [%s]", c.errMsg, err.Error())
					}
					return
				} else {
					t.Fatal("Unexpected Error occurred: ", err)
				}
			}
			if c.expectHasError {
				t.Fatal("Error should be raised")
			}

			defer func() {
				for _, content := range actual.Contents {
					if content.Path != nil {
						cleanutil.Remove(context.TODO(), *content.Path)
					}
				}
			}()

			if actual.ContentType != c.expectContentType {
				t.Errorf("content-type should be %s, but %s", c.expectContentType, actual.ContentType)
			}
			if len(actual.Contents) != c.expectContentLength {
				t.Errorf("length of content should be 1, but %d", len(actual.Contents))
			}
			if c.expectContentLength > 0 {
				content := actual.Contents[0]
				if content.Path == nil {
					t.Errorf("content.Path should not be nil")
				}
				actualBody, _ := ioutil.ReadFile(*content.Path)
				if string(actualBody) != c.expectBody {
					t.Errorf("body should be [%s], but [%s]", c.expectBody, string(actualBody))
				}
				if c.expectMetadata != nil {
					// Not even reflect.DeepEqual was able to compare these variables...
					expectMetadata, err := json.Marshal(c.expectMetadata)
					if err != nil {
						t.Fatalf("failed to marshal c.expectMetadata")
					}
					actualMetadata, err := json.Marshal(content.Metadata)
					if err != nil {
						t.Fatalf("failed to marshal content.Metadata")
					}
					if string(expectMetadata) != string(actualMetadata) {
						t.Errorf(
							"metadata should be [%+v], but [%+v]",
							string(expectMetadata), string(actualMetadata))
					}
				}
			}
		})
	}
}

func TestFromOutput(t *testing.T) {
	cases := []struct {
		name            string
		conf            *config.Configuration
		expectHasError  bool
		errMsg          string
		expectChannelID string
	}{
		{
			name: "normal datalake channel",
			conf: &config.Configuration{
				Output: "{\"$datalake:1\": \"1111111111111\"}",
			},
			expectHasError:  false,
			errMsg:          "",
			expectChannelID: "1111111111111",
		}, {
			name:            "no output",
			conf:            &config.Configuration{},
			expectHasError:  false,
			errMsg:          "",
			expectChannelID: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := FromOutput(c.conf)
			if err != nil {
				if c.expectHasError {
					if err.Error() != c.errMsg {
						t.Fatalf("Error message should be [%s], but [%s]", c.errMsg, err.Error())
					}
					return
				} else {
					t.Fatal("Unexpected Error occurred: ", err)
				}
			}
			if c.expectHasError {
				t.Fatal("Error should be raised")
			}

			if actual != c.expectChannelID {
				t.Errorf("channelID should be %s, but %s", c.expectChannelID, actual)
			}
		})
	}
}
