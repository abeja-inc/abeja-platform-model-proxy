package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/entity"
)

func TestTransportOneshotMessage_DialError(t *testing.T) {
	errOnBoot := make(chan int)
	notifyFromMain := make(chan int)
	notifyToMain := make(chan int)
	defer close(notifyFromMain)

	ticker := *time.NewTicker(2 * time.Second)
	conf := &config.Configuration{}
	scopeChan := make(chan context.Context, 10)
	defer close(scopeChan)
	go TransportOneshotMessage(
		context.TODO(),
		conf,
		"invalid_socket_file",
		errOnBoot,
		notifyFromMain,
		notifyToMain,
		nil)
	for {
		select {
		case <-ticker.C:
			t.Error("timeout on DialError")
			ticker.Stop()
			close(errOnBoot)
			return
		case <-notifyToMain:
		case <-errOnBoot:
			ticker.Stop()
			return
		}
	}
}

func TestHandleResult_NoBody(t *testing.T) {
	conf := &config.Configuration{
		APIURL: "http://localhost",
		RunID:  "dummy",
	}
	datalakeChannelID := "dummy"

	cases := []struct {
		name         string
		statusCode   int
		expectNotify int
	}{
		{
			name:         "ok",
			statusCode:   200,
			expectNotify: 0,
		},
		{
			name:         "ng",
			statusCode:   500,
			expectNotify: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := entity.Response{
				StatusCode: &c.statusCode,
			}
			notifyToMain := make(chan int, 1)
			defer close(notifyToMain)

			handleResult(
				context.TODO(),
				conf,
				res,
				datalakeChannelID,
				notifyToMain,
				nil,
			)

			var actual int
			select {
			case v := <-notifyToMain:
				actual = v
			default:
				actual = 0
			}
			if actual != c.expectNotify {
				t.Errorf(
					"it should receive %d from notifyToMain, but received %d",
					c.expectNotify, actual)
			}
		})
	}
}
