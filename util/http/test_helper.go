package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

type MockRoundTripFunc func(req *http.Request) *http.Response

func (f MockRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn MockRoundTripFunc) *http.Client {
	return &http.Client{
		Transport: MockRoundTripFunc(fn),
	}
}

type ResponseMock struct {
	Path           string /* request path */
	NeedAuth       bool   /* path need auth ? */
	AuthToken      string
	StatusCode     int
	ResponseBody   string
	ResponseHeader http.Header
}

func GetMockHTTPClient(t *testing.T, resps []ResponseMock) *http.Client {
	t.Helper()

	return NewTestClient(func(req *http.Request) *http.Response {
		if len(resps) < 1 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(`{"message": "no response mock"}`))),
				Header:     make(http.Header),
			}
		}

		var resp *ResponseMock
		for index, r := range resps {
			if req.URL.Path == r.Path {
				resp = &r
				defer func() {
					// remove used mock object
					upd := []ResponseMock{}
					for i := range resps {
						if i != index {
							upd = append(upd, resps[i])
						}
					}
					resps = upd
				}()
				break
			}
		}
		if resp != nil {
			if resp.NeedAuth {
				// need auth
				authToken := req.Header.Get("Authorization")
				if authToken == "" || authToken != fmt.Sprintf("Bearer %s", resp.AuthToken) {
					return &http.Response{
						StatusCode: http.StatusUnauthorized,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(""))),
						Header:     make(http.Header),
					}
				}
			}
			return &http.Response{
				StatusCode: resp.StatusCode,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(resp.ResponseBody))),
				Header:     resp.ResponseHeader,
			}
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(""))),
			Header:     make(http.Header),
		}
	})
}
