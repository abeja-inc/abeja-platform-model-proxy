package http

import (
	"strings"
	"testing"

	"github.com/abeja-inc/platform-model-proxy/util/auth"
)

func TestBuildURL(t *testing.T) {

	authInfo := auth.AuthInfo{AuthToken: "token"}
	client, _ := NewRetryHTTPClient("http://localhost", 1, 1, 1, authInfo, nil)

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
