package proxy

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/entity"
	cleanutil "github.com/abeja-inc/abeja-platform-model-proxy/util/clean"
)

func RemoveUDSFile(path string, t *testing.T) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		if err = os.Remove(path); err != nil {
			t.Fatal("Error when removing existing file for unix domain socket:", err)
		}
	}
}

func TestTransportMessage_OK(t *testing.T) {
	errOnBoot := make(chan int)
	request := make(chan entity.ContentList)
	response := make(chan entity.Response)
	notifyFromMain := make(chan int)
	notifyToMain := make(chan int)
	scopeChan := make(chan context.Context, 10)
	defer close(errOnBoot)
	defer close(request)
	defer close(response)
	defer close(scopeChan)
	defer close(notifyFromMain)

	path := filepath.Join(os.TempDir(), "test_unixdomainsocket")
	RemoveUDSFile(path, t)
	listener, _ := net.Listen("unix", path)
	defer cleanutil.Close(context.TODO(), listener, path)

	contentPath := "/path/to/content"
	content := entity.Content{
		ContentType: nil,
		Path:        &contentPath,
		FileName:    nil,
		FormName:    nil,
	}
	reqHeader := entity.Header{
		Key:    "content-type",
		Values: []string{"application/json"},
	}
	reqCL := entity.ContentList{
		Method:      "POST",
		ContentType: "application/json",
		Headers:     []*entity.Header{&reqHeader},
		Contents:    []*entity.Content{&content},
	}

	respContentType := "image/jpeg"
	respMetadata := map[string]string{
		"key1":                  "value1",
		"key2":                  "value2",
		"X-Abeja-Model-Id":      "dummy-model-id",
		"X-Abeja-Model-Version": url.PathEscape("dummy−model-version"),
		"X-Abeja-Deployment-Id": "dummy-deployment-id",
		"X-Abeja-Service-Id":    "dummy-service-id",
	}
	respPath := "/path/to/resp/content"
	respStatus := http.StatusOK
	respFromRuntime := entity.Response{
		ContentType: &respContentType,
		Metadata:    &respMetadata,
		Path:        &respPath,
		ErrMsg:      nil,
		StatusCode:  &respStatus,
	}

	// mock for runtime
	go func() {
		fd, _ := listener.Accept()

		// receive & check request
		headBuf := make([]byte, 8)
		if _, err := io.ReadFull(fd, headBuf); err != nil {
			t.Error("Error when reading header:", err)
		}
		var header Header
		if err := binary.Read(bytes.NewReader(headBuf), binary.BigEndian, &header); err != nil {
			t.Error("Error when reading header:", err)
		}
		if !bytes.Equal([]byte{magic0, magic1, magic2}, header.Magic[:]) {
			t.Errorf("header.Magic should be `ABE9A`, but %v", header.Magic)
		}
		if version != header.Version {
			t.Errorf("header.Version should be %v, but %v", version, header.Version)
		}

		bodyBuf := make([]byte, header.Length)
		_, err := io.ReadFull(fd, bodyBuf)
		if err != nil {
			t.Error(err)
		}
		var body entity.ContentList
		if err := json.Unmarshal(bodyBuf, &body); err != nil {
			t.Error("Error when unmarshaling body:", err)
		}
		if "POST" != body.Method {
			t.Errorf("body.Method should be `POST`, but %s", body.Method)
		}
		if "application/json" != body.ContentType {
			t.Errorf("body.ContentType should be `application/json`, but %s", body.ContentType)
		}
		if 1 != len(body.Headers) {
			t.Errorf("len(body.Headers should be `1`, but %d", len(body.Headers))
		}
		if "content-type" != body.Headers[0].Key {
			t.Errorf("body.Headers[0].Key should be `content-type`, but %s", body.Headers[0].Key)
		}
		if 1 != len(body.Headers[0].Values) {
			t.Errorf("len(body.Headers[0].Values should be `1`, but %d", len(body.Headers[0].Values))
		}
		if "application/json" != body.Headers[0].Values[0] {
			t.Errorf("body.Headers[0].Values[0] should be `application/json`, but %s", body.Headers[0].Values[0])
		}
		if 1 != len(body.Contents) {
			t.Errorf("len(body.Contents should be `1`, but %d", len(body.Contents))
		} else {
			content := body.Contents[0]
			if content.ContentType != nil {
				t.Errorf("content.ContentType should be nil, but %s", *content.ContentType)
			}
			if content.Path == nil {
				t.Error("content.Path should not be nil")
			} else {
				if contentPath != *content.Path {
					t.Errorf("content.Path should be %s, but %s", contentPath, *content.Path)
				}
			}
			if content.FileName != nil {
				t.Errorf("content.FileName should be nil, but %s", *content.FileName)
			}
			if content.FormName != nil {
				t.Errorf("content.FormName should be nil, but %s", *content.FormName)
			}
		}

		// send response
		b, _ := json.Marshal(respFromRuntime)
		respHeader := Header{
			Magic:   [3]byte{magic0, magic1, magic2},
			Version: version,
			Length:  uint32(len(b)),
		}
		if err := binary.Write(fd, binary.BigEndian, respHeader); err != nil {
			t.Error("Error when writing response header:", err)
		}
		if _, err := fd.Write(b); err != nil {
			t.Error("Error when writing response body:", err)
		}

		cleanutil.Close(context.TODO(), fd, "Listener#Accept")
	}()

	conf := &config.Configuration{
		ModelID:      "dummy-model-id",
		ModelVersion: "dummy−model-version", // attention, using not hyphen
		DeploymentID: "dummy-deployment-id",
		ServiceID:    "dummy-service-id",
	}
	go TransportMessages(context.TODO(), conf, path, request, response, errOnBoot, notifyFromMain, notifyToMain, scopeChan, nil)

	var resp *entity.Response
	ticker := *time.NewTicker(2 * time.Second)
	request <- reqCL

B:
	for {
		select {
		case <-ticker.C:
			t.Error("timeout on ReadFromSocket")
			ticker.Stop()
			break B
		case r := <-response:
			resp = &r
			ticker.Stop()
			break B
		}
	}

	// check response from runtime
	if resp == nil {
		t.Fatal("resp should not be nil")
	}
	if resp.ContentType == nil {
		t.Error("ContentType should be not nil")
	} else {
		if "image/jpeg" != *resp.ContentType {
			t.Errorf(
				"ContentType should be `image/jpeg`, but %s", *resp.ContentType)
		}
	}
	if resp.Metadata == nil {
		t.Errorf("Metadata should not be nil")
	} else {
		if reflect.DeepEqual(respMetadata, *resp.Metadata) != true {
			t.Errorf("Metadata should be %+v, but %+v", respMetadata, *resp.Metadata)
		}
	}
	if resp.Path == nil {
		t.Errorf("Path should not be nil")
	} else {
		if respPath != *resp.Path {
			t.Errorf("Path should be %s, but %s", respPath, *resp.Path)
		}
	}
	if resp.ErrMsg != nil {
		t.Errorf("ErrMsg should be nil, but %s", *resp.ErrMsg)
	}
	if resp.StatusCode == nil {
		t.Error("StatusCode should be not nil")
	} else {
		if http.StatusOK != *resp.StatusCode {
			t.Errorf(
				"StatusCode should be %d, but %d",
				http.StatusOK, *resp.StatusCode)
		}
	}
}

func TestTransportMessage_DialError(t *testing.T) {
	errOnBoot := make(chan int)
	request := make(chan entity.ContentList)
	response := make(chan entity.Response)
	notifyFromMain := make(chan int)
	notifyToMain := make(chan int)
	defer close(request)
	defer close(response)
	defer close(notifyFromMain)

	ticker := *time.NewTicker(2 * time.Second)
	conf := &config.Configuration{}
	scopeChan := make(chan context.Context, 10)
	defer close(scopeChan)
	go TransportMessages(
		context.TODO(),
		conf,
		"invalid_socket_file",
		request,
		response,
		errOnBoot,
		notifyFromMain,
		notifyToMain,
		scopeChan,
		nil)
	for {
		select {
		case <-ticker.C:
			t.Error("timeout on DialError")
			ticker.Stop()
			close(errOnBoot)
			return
		case <-errOnBoot:
			ticker.Stop()
			return
		}
	}
}

func TestTransportMessage_WriteToSocketError(t *testing.T) {
	errOnBoot := make(chan int)
	request := make(chan entity.ContentList)
	response := make(chan entity.Response)
	notifyFromMain := make(chan int)
	notifyToMain := make(chan int)
	defer close(errOnBoot)
	defer close(request)
	defer close(response)
	defer close(notifyFromMain)

	path := filepath.Join(os.TempDir(), "test_unixdomainsocket")
	RemoveUDSFile(path, t)
	listener, _ := net.Listen("unix", path)

	ticker := *time.NewTicker(2 * time.Second)
	conf := &config.Configuration{}
	scopeChan := make(chan context.Context, 10)
	defer close(scopeChan)
	go TransportMessages(context.TODO(), conf, path, request, response, errOnBoot, notifyFromMain, notifyToMain, scopeChan, nil)
	time.Sleep(100 * time.Millisecond)

	// close socket for occurring write error
	if err := listener.Close(); err != nil {
		t.Fatal("Error when closing listener:", err)
	}
	request <- entity.ContentList{
		Method:      "GET",
		ContentType: "",
		Contents:    nil,
	}
	var resp *entity.Response
B:
	for {
		select {
		case <-ticker.C:
			t.Error("timeout on WriteToSocketError")
			ticker.Stop()
			break B
		case r := <-response:
			resp = &r
			ticker.Stop()
			break B
		}
	}
	assertInternalServerErrorResponse(t, resp)
}

func TestTransportMessage_ReadFromSocketError(t *testing.T) {
	errOnBoot := make(chan int)
	request := make(chan entity.ContentList)
	response := make(chan entity.Response)
	notifyFromMain := make(chan int)
	notifyToMain := make(chan int)
	defer close(errOnBoot)
	defer close(request)
	defer close(response)
	defer close(notifyFromMain)

	path := filepath.Join(os.TempDir(), "test_unixdomainsocket")
	RemoveUDSFile(path, t)
	listener, _ := net.Listen("unix", path)

	contentPath := "/path/to/content"
	content := entity.Content{
		ContentType: nil,
		Path:        &contentPath,
		FileName:    nil,
		FormName:    nil,
	}
	reqCL := entity.ContentList{
		Method:      "POST",
		ContentType: "application/json",
		Contents:    []*entity.Content{&content},
	}

	go func() {
		fd, _ := listener.Accept()
		headBuf := make([]byte, 8)
		if _, err := io.ReadFull(fd, headBuf); err != nil {
			t.Error("Error when reading header:", err)
		}
		var header Header
		if err := binary.Read(bytes.NewReader(headBuf), binary.BigEndian, &header); err != nil {
			t.Error("Error when reading header:", err)
		}
		if !bytes.Equal([]byte{magic0, magic1, magic2}, header.Magic[:]) {
			t.Errorf("header.Magic should be `ABE9A`, but %v", header.Magic)
		}
		if version != header.Version {
			t.Errorf("header.Version should be %v, but %v", version, header.Version)
		}

		bodyBuf := make([]byte, header.Length)
		_, err := io.ReadFull(fd, bodyBuf)
		if err != nil {
			t.Error(err)
		}
		var body entity.ContentList
		if err := json.Unmarshal(bodyBuf, &body); err != nil {
			t.Error("Error when unmarshaling body:", err)
		}
		if "POST" != body.Method {
			t.Errorf("body.Method should be `POST`, but %s", body.Method)
		}
		if "application/json" != body.ContentType {
			t.Errorf("body.ContentType should be `application/json`, but %s", body.ContentType)
		}
		if 1 != len(body.Contents) {
			t.Errorf("len(body.Contents should be `1`, but %d", len(body.Contents))
		} else {
			content := body.Contents[0]
			if content.ContentType != nil {
				t.Errorf("content.ContentType should be nil, but %s", *content.ContentType)
			}
			if content.Path == nil {
				t.Error("content.Path should not be nil")
			} else {
				if contentPath != *content.Path {
					t.Errorf("content.Path should be %s, but %s", contentPath, *content.Path)
				}
			}
			if content.FileName != nil {
				t.Errorf("content.FileName should be nil, but %s", *content.FileName)
			}
			if content.FormName != nil {
				t.Errorf("content.FormName should be nil, but %s", *content.FormName)
			}
		}
		cleanutil.Close(context.TODO(), fd, "Listener#Accept")
		// close socket for occurring read error
		if err := listener.Close(); err != nil {
			t.Error("Error when closing listener:", err)
		}
	}()

	conf := &config.Configuration{}
	scopeChan := make(chan context.Context, 10)
	defer close(scopeChan)
	go TransportMessages(
		context.TODO(), conf, path, request, response, errOnBoot, notifyFromMain, notifyToMain, scopeChan, nil)

	var resp *entity.Response
	ticker := *time.NewTicker(2 * time.Second)
	request <- reqCL

B:
	for {
		select {
		case <-ticker.C:
			t.Error("timeout on ReadFromSocketError")
			ticker.Stop()
			break B
		case r := <-response:
			resp = &r
			ticker.Stop()
			break B
		}
	}
	assertInternalServerErrorResponse(t, resp)
}

func assertInternalServerErrorResponse(t *testing.T, resp *entity.Response) {
	t.Helper()

	if resp == nil {
		t.Fatal("resp should not be nil")
	}
	if resp.ContentType == nil {
		t.Error("ContentType should be not nil")
	} else {
		if "application/json" != *resp.ContentType {
			t.Errorf(
				"ContentType should be `application/json`, but %s", *resp.ContentType)
		}
	}
	if resp.Metadata != nil {
		t.Errorf("Metadata should be nil, but %v", *resp.Metadata)
	}
	if resp.Path != nil {
		t.Errorf("Path should be nil, but %s", *resp.Path)
	}
	if resp.ErrMsg == nil {
		t.Error("ErrMsg should not be nil")
	} else {
		if !strings.HasPrefix(*resp.ErrMsg, "Internal Server Error") {
			t.Errorf("ErrMsg should start with `Internal Server Error`, but `%s`", *resp.ErrMsg)
		}
	}
	if resp.StatusCode == nil {
		t.Error("StatusCode should be not nil")
	} else {
		if http.StatusInternalServerError != *resp.StatusCode {
			t.Errorf(
				"StatusCode should be %d, but %d",
				http.StatusInternalServerError, *resp.StatusCode)
		}
	}
}
