package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/abeja-inc/platform-model-proxy/util/auth"
)

func TestBuildURL(t *testing.T) {

	authInfo := auth.AuthInfo{AuthToken: "token"}
	client, _ := NewRetryHTTPClient("http://localhost", 1, 1, authInfo, nil)

	cases := []struct {
		name           string
		path           string
		param          map[string]interface{}
		expected       string
		expectedParams []string
	}{
		{
			name:           "path with first slash",
			path:           "/foo/bar",
			param:          map[string]interface{}{},
			expected:       "http://localhost/foo/bar",
			expectedParams: []string{},
		}, {
			name:           "path without first slash",
			path:           "foo/bar",
			param:          nil,
			expected:       "http://localhost/foo/bar",
			expectedParams: []string{},
		}, {
			name: "with param",
			path: "/foo/bar",
			param: map[string]interface{}{
				"hoge": "fuga",
				"baz":  "qux",
			},
			expected: "http://localhost/foo/bar?",
			expectedParams: []string{
				"hoge=fuga",
				"baz=qux",
			},
		}, {
			name: "no path",
			path: "",
			param: map[string]interface{}{
				"hoge": "fuga",
				"baz":  "qux",
			},
			expected: "http://localhost?",
			expectedParams: []string{
				"hoge=fuga",
				"baz=qux",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := client.BuildURL(c.path, c.param)
			if !strings.HasPrefix(actual, c.expected) {
				t.Errorf("result should be start with %s, but %s", c.expected, actual)
			}
			if len(c.param) > 0 {
				for _, s := range c.expectedParams {
					if !strings.Contains(actual, s) {
						t.Errorf("result should contains %s, but not contains", s)
					}
				}
			}
		})
	}
}

func TestGatewayTimeoutRetry(t *testing.T) {
	// リクエスト回数のカウンタ
	count := 0

	// サーバー作成
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
		count++
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// クライアント作成
	authInfo := auth.AuthInfo{AuthToken: "token"}
	client, _ := NewRetryHTTPClient(server.URL, 10, 4, authInfo, nil)

	// アクセスしてみる
	tStart := time.Now()
	res, err := client.GetThrough(server.URL)
	tElapsed := time.Since(tStart)

	if tElapsed.Seconds() < 14 || 15 < tElapsed.Seconds() {
		t.Errorf("should be exponential backoff. want=14s, result=%fs", tElapsed.Seconds())
	}
	if count != 4 {
		t.Errorf("wrong retry time. want=4, result=%d", count)
	}
	if res == nil || res.StatusCode != 504 {
		t.Errorf("response shoud be 504")
	}
	if err != nil {
		t.Error("shouldn't return error")
	}
}
