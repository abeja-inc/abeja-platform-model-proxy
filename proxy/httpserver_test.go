package proxy

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/subprocess"
)

func TestHealthCheck(t *testing.T) {
	runtime := &subprocess.Runtime{
		Cmd:    nil,
		Status: subprocess.RuntimeStatusPreparing,
	}
	reqChan := make(chan entity.ContentList)
	resChan := make(chan entity.Response)
	defer close(reqChan)
	defer close(resChan)
	conf := config.NewConfiguration()
	conf.Port = config.DefaultHTTPListenPort
	conf.HealthCheckPort = config.DefaultHealthCheckListenPort
	server, err := CreateHTTPServer(runtime, reqChan, resChan, &conf)
	if err != nil {
		t.Fatal("unexpected error occurred", err)
	}

	cases := []struct {
		name          string
		runtimeStatus subprocess.RuntimeStatus
		httpStatus    int
		resBody       string
	}{
		{
			name:          "preparing",
			runtimeStatus: subprocess.RuntimeStatusPreparing,
			httpStatus:    http.StatusServiceUnavailable,
			resBody:       "{\"status\":\"service unavailable\"}",
		}, {
			name:          "running",
			runtimeStatus: subprocess.RuntimeStatusRunning,
			httpStatus:    http.StatusOK,
			resBody:       "{\"status\":\"ok\"}",
		}, {
			name:          "already-exited-with-success",
			runtimeStatus: subprocess.RuntimeStatusExitedWithSuccess,
			httpStatus:    http.StatusNotFound,
			resBody:       "{\"status\":\"service not found\"}",
		}, {
			name:          "already-exited-with-failure",
			runtimeStatus: subprocess.RuntimeStatusExitedWithFailure,
			httpStatus:    http.StatusServiceUnavailable,
			resBody:       "{\"status\":\"service unavailable\"}",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/health_check", nil)
			rec := httptest.NewRecorder()
			runtime.Status = c.runtimeStatus

			server.HealthCheckServer.Handler.ServeHTTP(rec, req)
			if c.httpStatus != rec.Code {
				t.Errorf("http status should be %d, but %d", c.httpStatus, rec.Code)
			}
			if c.resBody != rec.Body.String() {
				t.Errorf("response body should be [%s], but [%s]", c.resBody, rec.Body.String())
			}
		})
	}
}

func TestRequest(t *testing.T) {
	runtime := &subprocess.Runtime{
		Cmd:    nil,
		Status: subprocess.RuntimeStatusRunning,
	}
	reqChan := make(chan entity.ContentList)
	resChan := make(chan entity.Response)
	defer close(reqChan)
	defer close(resChan)
	conf := config.NewConfiguration()
	conf.Port = config.DefaultHTTPListenPort
	server, err := CreateHTTPServer(runtime, reqChan, resChan, &conf)
	if err != nil {
		t.Fatal("unexpected error occurred", err)
	}

	cases := []struct {
		name             string
		reqMethod        string
		reqContentType   string
		additionalHeader entity.Header
		reqBody          string
		reqBinary        string
		resStatusCode    int
		resContentType   string
		resBody          string
		modelId          string
		modelVersion     string
		deploymentId     string
		serviceId        string
		resModelId       string
		resModelVersion  string
		resDeploymentId  string
		resServiceId     string
	}{
		{
			name:           "POST-json",
			reqMethod:      "POST",
			reqContentType: "application/json",
			additionalHeader: entity.Header{
				Key:    "X-Abeja-Model-Version",
				Values: []string{"1.0"},
			},
			reqBody:         "{\"foo\":\"bar\"}",
			reqBinary:       "",
			resStatusCode:   http.StatusOK,
			resContentType:  "application/json",
			resBody:         "{\"baz\":\"qux\"}",
			modelId:         "1300000000000",
			modelVersion:    "1.0.0",
			deploymentId:    "1400000000000",
			serviceId:       "1500000000000",
			resModelId:      "1300000000000",
			resModelVersion: "1.0.0",
			resDeploymentId: "1400000000000",
			resServiceId:    "1500000000000",
		}, {
			name:           "GET",
			reqMethod:      "GET",
			reqContentType: "application/x-www-form-urlencoded",
			additionalHeader: entity.Header{
				Key:    "Set-Cookie",
				Values: []string{"cookie1", "cookie2"},
			},
			reqBody:         "foo=bar&baz=qux",
			reqBinary:       "",
			resStatusCode:   http.StatusBadRequest,
			resContentType:  "application/json",
			resBody:         "{\"message\":\"invalid request\"}",
			modelId:         "1300000000000",
			modelVersion:    "1.0.0",
			deploymentId:    "1400000000000",
			serviceId:       "1500000000000",
			resModelId:      "1300000000000",
			resModelVersion: "1.0.0",
			resDeploymentId: "1400000000000",
			resServiceId:    "1500000000000",
		}, {
			name:             "POST-text/plain",
			reqMethod:        "POST",
			reqContentType:   "text/plain",
			additionalHeader: entity.Header{},
			reqBody:          "foo = bar, baz. \nqux...",
			reqBinary:        "",
			resStatusCode:    http.StatusOK,
			resContentType:   "application/json",
			resBody:          "{\"message\":\"ok\"}",
			modelId:          "1300000000000",
			modelVersion:     "1.0.0",
			deploymentId:     "1400000000000",
			serviceId:        "1500000000000",
			resModelId:       "1300000000000",
			resModelVersion:  "1.0.0",
			resDeploymentId:  "1400000000000",
			resServiceId:     "1500000000000",
		}, {
			name:             "POST-text/csv",
			reqMethod:        "POST",
			reqContentType:   "text/csv",
			additionalHeader: entity.Header{},
			reqBody:          "foo,bar,baz\nqux,quux,corge",
			reqBinary:        "",
			resStatusCode:    http.StatusOK,
			resContentType:   "text/csv",
			resBody:          "hoge,fuga\npiyo,hogege",
			modelId:          "1300000000000",
			modelVersion:     "1.0.0",
			deploymentId:     "1400000000000",
			serviceId:        "1500000000000",
			resModelId:       "1300000000000",
			resModelVersion:  "1.0.0",
			resDeploymentId:  "1400000000000",
			resServiceId:     "1500000000000",
		}, {
			name:             "POST-image/jpeg",
			reqMethod:        "POST",
			reqContentType:   "image/jpeg",
			additionalHeader: entity.Header{},
			reqBody:          "",
			reqBinary:        "../test_resources/cat.jpg",
			resStatusCode:    http.StatusOK,
			resContentType:   "application/json",
			resBody:          "{\"message\":\"ok\"}",
			modelId:          "1300000000000",
			modelVersion:     "201906−model2",
			deploymentId:     "1400000000000",
			serviceId:        "1500000000000",
			resModelId:       "1300000000000",
			resModelVersion:  "201906%E2%88%92model2",
			resDeploymentId:  "1400000000000",
			resServiceId:     "1500000000000",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := func(reqChan chan entity.ContentList, resChan chan entity.Response) {
				cl := <-reqChan
				if c.reqMethod != cl.Method {
					t.Errorf("request method should be %s, but %s", c.reqMethod, cl.Method)
				}
				if c.reqContentType != cl.ContentType {
					t.Errorf(
						"request content-type should be [%s], but [%s]",
						c.reqContentType, cl.ContentType)
				}
				if cl.Headers[0].Key != "content-type" {
					t.Errorf(
						"request header[0].Key should be [%s], but [%s]",
						"content-type", cl.Headers[0].Key)
				}
				if cl.Headers[0].Values[0] != c.reqContentType {
					t.Errorf(
						"request header[0].Values should be [%s], but [%s]",
						c.reqContentType, cl.Headers[0].Values[0])
				}
				if c.additionalHeader.Key != "" {
					if len(cl.Headers) != 2 {
						t.Errorf("request header size should be 2, but %d", len(cl.Headers))
					}
					if cl.Headers[1].Key != strings.ToLower(c.additionalHeader.Key) {
						t.Errorf(
							"request header[1].Key should be %s, but %s",
							strings.ToLower(c.additionalHeader.Key), cl.Headers[1].Key)
					}
					if len(cl.Headers[1].Values) != len(c.additionalHeader.Values) {
						t.Errorf(
							"request header[1].Values size should be %d, but %d",
							len(c.additionalHeader.Values), len(cl.Headers[1].Values))
					}
					for i := 0; i < len(cl.Headers[1].Values); i++ {
						if cl.Headers[1].Values[i] != c.additionalHeader.Values[i] {
							t.Errorf(
								"request header[1].Values[%d] should be %s, but %s",
								i, c.additionalHeader.Values[i], cl.Headers[1].Values[i])
						}
					}
				}
				if len(cl.Contents) != 1 {
					t.Errorf("request content length should be 1, but %d", len(cl.Contents))
				}
				ctt := cl.Contents[0]
				if ctt.ContentType != nil {
					t.Errorf("content's Content-Type should be nil, but %s", *ctt.ContentType)
				}
				if ctt.Path == nil {
					t.Error("content's Path should not be nil, but nil")
				} else {
					reqBody, err := ioutil.ReadFile(*ctt.Path)
					if err != nil {
						t.Error("unexpected error occurred", err)
					}
					if c.reqBinary == "" {
						if c.reqBody != string(reqBody) {
							t.Errorf(
								"req body should be [%s], but [%s]", c.reqBody, string(reqBody))
						}
					} else {
						expectBody, err := ioutil.ReadFile(c.reqBinary)
						if err != nil {
							t.Error("unexpected error occurred", err)
						}
						if !bytes.Equal(expectBody, reqBody) {
							t.Errorf("req body is different of source binary")
						}
					}
				}

				f, err := ioutil.TempFile("", "")
				if err != nil {
					t.Error("unexpected error occurred", err)
				}
				filePath := f.Name()
				if err := f.Close(); err != nil {
					t.Error("Error when closing file:", err)
				}
				if err = ioutil.WriteFile(filePath, []byte(c.resBody), 0644); err != nil {
					t.Error("unexpected error occurred", err)
				}

				metadata := map[string]string{
					"X-Abeja-Model-Id":      c.modelId,
					"X-Abeja-Model-Version": url.PathEscape(c.modelVersion),
					"X-Abeja-Deployment-Id": c.deploymentId,
					"X-Abeja-Service-Id":    c.serviceId,
				}
				resContentType := c.resContentType
				resStatusCode := c.resStatusCode
				res := entity.Response{
					ContentType: &resContentType,
					Metadata:    &metadata,
					Path:        &filePath,
					ErrMsg:      nil,
					StatusCode:  &resStatusCode,
				}
				resChan <- res
			}
			go f(reqChan, resChan)

			var body io.Reader
			path := "/"
			if c.reqMethod == "POST" {
				if c.reqBinary == "" {
					body = bytes.NewBufferString(c.reqBody)
				} else {
					data, err := ioutil.ReadFile(c.reqBinary)
					if err != nil {
						t.Fatalf("failed to read from file %s: %s", c.reqBinary, err.Error())
					}
					body = bytes.NewBuffer(data)
				}
			} else {
				path = path + "?" + c.reqBody
			}
			conf.ModelID = c.modelId
			conf.ModelVersion = c.modelVersion
			conf.DeploymentID = c.deploymentId
			conf.ServiceID = c.serviceId
			req := httptest.NewRequest(c.reqMethod, path, body)

			req.Header.Add("Content-Type", c.reqContentType)
			if len(c.additionalHeader.Values) > 0 {
				for i := 0; i < len(c.additionalHeader.Values); i++ {
					req.Header.Add(c.additionalHeader.Key, c.additionalHeader.Values[i])
				}
			}
			rec := httptest.NewRecorder()
			server.Server.Handler.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			if c.resStatusCode != res.StatusCode {
				t.Errorf("http status should be %d, but %d", c.resStatusCode, rec.Code)
			}
			if c.resBody != rec.Body.String() {
				t.Errorf("response body should be [%s], but [%s]", c.resBody, rec.Body.String())
			}
			resContentType := res.Header.Get("Content-Type")
			if c.resContentType != resContentType {
				t.Errorf("response Content-Type should be [%s], but [%s]", c.resContentType, resContentType)
			}
			resModelId := res.Header.Get("x-abeja-model-id")
			if c.resModelId != resModelId {
				t.Errorf("response model id should be [%s], but [%s]", c.resModelId, resModelId)
			}
			resModelVersion := res.Header.Get("x-abeja-model-version")
			if c.resModelVersion != resModelVersion {
				t.Errorf("response model version should be [%s], but [%s]", c.resModelVersion, resModelVersion)
			}
			resDeploymentId := res.Header.Get("x-abeja-deployment-id")
			if c.resDeploymentId != resDeploymentId {
				t.Errorf("response deployment id should be [%s], but [%s]", c.resDeploymentId, resDeploymentId)
			}
			resServiceId := res.Header.Get("x-abeja-service-id")
			if c.resServiceId != resServiceId {
				t.Errorf("response service id should be [%s], but [%s]", c.resServiceId, resServiceId)
			}
		})
	}
}
