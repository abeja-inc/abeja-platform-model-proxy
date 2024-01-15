//go:build extra
// +build extra

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abeja-inc/abeja-platform-model-proxy/util/auth"
)

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

	// 検証
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
