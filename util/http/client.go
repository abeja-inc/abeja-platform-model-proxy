package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/sethgrid/pester"

	"github.com/abeja-inc/platform-model-proxy/util/auth"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

type RetryClient struct {
	client   *pester.Client
	baseURI  *url.URL
	authInfo auth.AuthInfo
}

func NewRetryHTTPClient(
	baseURL string,
	timeout int,
	retryLimit int,
	authInfo auth.AuthInfo,
	option *http.Client) (*RetryClient, error) {

	uri, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Errorf(": %w", err)
	}
	if !strings.HasPrefix(uri.Scheme, "http") {
		return nil, errors.Errorf("unsupported scheme of baseURI: %s", uri.Scheme)
	}
	if timeout < 1 {
		return nil, errors.Errorf(
			"timeout must be more than eaual 1, but specified %d", timeout)
	}
	if retryLimit < 0 {
		return nil, errors.Errorf(
			"retryLimit must be more than equal 0, but specified %d", retryLimit)
	}

	var client *pester.Client
	if option != nil {
		client = pester.NewExtendedClient(option)
	} else {
		client = pester.New()
	}
	client.Backoff = pester.ExponentialBackoff
	client.MaxRetries = retryLimit
	client.Timeout = time.Duration(timeout) * time.Second
	client.RetryOnHTTP429 = false
	client.KeepLog = true

	return &RetryClient{
		client:   client,
		baseURI:  uri,
		authInfo: authInfo,
	}, nil
}

func (c *RetryClient) GetThrough(reqURL string) (*http.Response, error) {
	resp, err := c.client.Get(reqURL)
	if resp != nil && resp.StatusCode >= 500 {
		log.Warningf(context.TODO(), "GET request failed(status=%d)\n%s", resp.StatusCode, c.client.LogString())
	}
	return resp, err
}

func (c *RetryClient) GetJson(reqPath string, param map[string]interface{}, buf interface{}) error {
	reqUrl := c.BuildURL(reqPath, param)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return errors.Errorf(": %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return errors.Errorf("response error with StatusCode: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	if err := json.Unmarshal(body, buf); err != nil {
		return errors.Errorf(": %w", err)
	}
	return nil
}

func (c *RetryClient) Do(request *http.Request) (*http.Response, error) {
	setAuthHeader(c.authInfo, request)
	resp, err := c.client.Do(request)
	if resp != nil && resp.StatusCode >= 500 {
		log.Warningf(context.TODO(), "GET request failed(status=%d)\n%s", resp.StatusCode, c.client.LogString())
	}
	return resp, err
}

func (c *RetryClient) BuildURL(reqPath string, param map[string]interface{}) string {
	reqUrl := *(c.baseURI)
	reqUrl.Path = path.Join(reqUrl.Path, reqPath)
	if param != nil {
		q := reqUrl.Query()
		for key, value := range param {
			switch v := value.(type) {
			case string:
				q.Set(key, v)
			case int:
				q.Set(key, strconv.Itoa(v))
			}
		}
		reqUrl.RawQuery = q.Encode()
	}
	return reqUrl.String()
}

func setAuthHeader(authInfo auth.AuthInfo, req *http.Request) {
	if authInfo.AuthToken != "" {
		req.Header.Set(
			"authorization", fmt.Sprintf("Bearer %s", authInfo.AuthToken))
	} else {
		req.SetBasicAuth(authInfo.UserID, authInfo.PersonalToken)
	}
}
