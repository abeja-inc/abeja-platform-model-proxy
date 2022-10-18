package proxy

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/convert"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/util"
	"github.com/abeja-inc/platform-model-proxy/util/auth"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	httpclient "github.com/abeja-inc/platform-model-proxy/util/http"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

const errorMessageForAsync = `{
  "status": 502,
  "headers": {
    "content-type": "application/json"
  },
  "body": {
    "error": "internal_server_error",
    "error_description": "Internal Server Error: unexpected error of %s"
  }
}`

func responseSyncUnexpectedError(code int, msg string, sendto chan entity.Response) {

	ct := "application/json"
	res := entity.Response{
		ContentType: &ct,
		Metadata:    nil,
		Path:        nil,
		ErrMsg:      &msg,
		StatusCode:  &code,
	}
	sendto <- res
}

func sendAsyncErrorToARMS(
	ctx context.Context,
	conf *config.Configuration,
	path string,
	token string,
	message string,
	option *http.Client) {

	authInfo := auth.AuthInfo{AuthToken: token}
	httpClient, err :=
		httpclient.NewRetryHTTPClient(conf.APIURL, 30, 3, 3, authInfo, option)
	if err != nil {
		log.Error(ctx, "unexpected error occurred in sending error async response: ", err)
		return
	}
	reqUrl := httpClient.BuildURL(path, nil)
	body := fmt.Sprintf(errorMessageForAsync, message)

	req, err := http.NewRequest(
		"PUT", reqUrl, strings.NewReader(body))
	if err != nil {
		log.Error(ctx, "unexpected error occurred in sending error async response: ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error(ctx, "unexpected error occurred in sending error async response: ", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Warningf(ctx, "response error from GW with StatusCode: %d", resp.StatusCode)
		return
	}
}

func responseInternalServerError(
	ctx context.Context,
	conf *config.Configuration,
	cl entity.ContentList,
	sendto chan entity.Response,
	message string,
	option *http.Client) {

	if cl.AsyncRequestID == "" {
		responseSyncUnexpectedError(
			http.StatusInternalServerError,
			"Internal Server Error: unexpected error of "+message,
			sendto)
	} else {
		path := buildARMSEndPoint(ctx, conf, cl.AsyncRequestID)
		sendAsyncErrorToARMS(ctx, conf, path, cl.AsyncARMSToken, message, option)
	}
}

func buildARMSEndPoint(ctx context.Context, conf *config.Configuration, requestID string) string {
	endpoint := fmt.Sprintf(
		"/organizations/%s/deployments/%s/results/%s",
		conf.OrganizationID, conf.DeploymentID, requestID)
	log.Debug(ctx, fmt.Sprintf("ARMS endpoint: %s", endpoint))
	return endpoint
}

// TransportMessages transports request from user to runtime and response from runtime to user.
func TransportMessages(
	procCtx context.Context,
	conf *config.Configuration,
	socketFilePath string,
	request chan entity.ContentList,
	response chan entity.Response,
	errOnBoot chan int,
	notifyFromMain chan int,
	notifyToMain chan int,
	scopeChan chan context.Context,
	option *http.Client) {

	conn, err := net.Dial("unix", socketFilePath)
	if err != nil {
		log.Errorf(procCtx, "Failed to dial to runtime: "+log.ErrorFormat, err)
		close(errOnBoot)
		return
	}
	defer cleanutil.Close(procCtx, conn, socketFilePath)

	respReceiver := make(chan []byte, 1)
	defer close(respReceiver)

FINISHED:
	for {
		contents, ok := <-request
		if !ok {
			close(notifyToMain)
			break
		}
		ctx := contents.Ctx
		scopeChan <- ctx

		header, body, err := FromRequest(&contents)
		if err != nil {
			log.Errorf(ctx, "json encode error: "+log.ErrorFormat, err)
			responseInternalServerError(ctx, conf, contents, response, "encoding from request", option)
			continue
		}

		if err = binary.Write(conn, binary.BigEndian, header); err != nil {
			log.Errorf(ctx, "Write IPC request header error: "+log.ErrorFormat, err)
			responseInternalServerError(ctx, conf, contents, response, "communication with runtime", option)
			continue
		}

		if _, err = conn.Write(body); err != nil {
			log.Errorf(ctx, "Write IPC request body error: "+log.ErrorFormat, err)
			responseInternalServerError(ctx, conf, contents, response, "communication with runtime", option)
			continue
		}

		go func() {
			headBuff := make([]byte, 8)
			if _, err = io.ReadFull(conn, headBuff); err != nil {
				log.Errorf(ctx, "Read IPC response header error: "+log.ErrorFormat, err)
				respReceiver <- []byte{}
				return
			}

			if err := binary.Read(bytes.NewReader(headBuff), binary.BigEndian, &header); err != nil {
				log.Errorf(ctx, "Read IPC response header error: "+log.ErrorFormat, err)
				respReceiver <- []byte{}
				return
			}

			log.Debug(ctx, "response body length = "+fmt.Sprint(header.Length))
			bodyBuff := make([]byte, header.Length)
			if _, err = io.ReadFull(conn, bodyBuff); err != nil {
				log.Errorf(ctx, "Read IPC response body error: "+log.ErrorFormat, err)
				respReceiver <- []byte{}
				return
			}
			log.Debugf(ctx, "response body = %s", string(bodyBuff))

			respReceiver <- bodyBuff
		}()

		select {
		case bodyBuff := <-respReceiver:
			res, err := ToResponse(bodyBuff, conf)
			if err != nil {
				responseInternalServerError(ctx, conf, contents, response, err.Error(), option)
			} else {
				sendResponse(ctx, res, response, conf, contents, option)
			}
			scopeChan <- procCtx
		case <-notifyFromMain:
			responseInternalServerError(ctx, conf, contents, response, "received signal", option)
			close(notifyToMain)
			scopeChan <- procCtx
			break FINISHED
		}
	}
	log.Debug(procCtx, "finish transporting")
}

func sendResponse(
	ctx context.Context,
	res entity.Response,
	response chan entity.Response,
	conf *config.Configuration,
	contents entity.ContentList,
	option *http.Client) {

	if contents.AsyncRequestID != "" {
		// async. send response to ARMS
		log.Debug(ctx, "send async response to GW...")
		sendAsyncResponse(ctx, conf, res, contents, option)
	} else {
		// sync
		log.Debug(ctx, "send sync response to client...")
		response <- res
	}
}

func sendAsyncResponse(
	ctx context.Context,
	conf *config.Configuration,
	res entity.Response,
	contents entity.ContentList,
	option *http.Client) {

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	done := make(chan error)
	defer close(done)

	path := buildARMSEndPoint(ctx, conf, contents.AsyncRequestID)

	statusCode, headers, body, err := convert.FromResponse(ctx, res)
	if err != nil {
		log.Errorf(ctx, "unexpected error occurred in sending async response: "+log.ErrorFormat, err)
		sendAsyncErrorToARMS(ctx, conf, path, contents.AsyncARMSToken, "response from runtime", option)
		return
	}
	defer deleteTempFiles(ctx, &contents, body)

	go func() {
		defer cleanutil.Close(ctx, pr, "pipeReader")
		authInfo := auth.AuthInfo{AuthToken: contents.AsyncARMSToken}
		httpClient, err :=
			httpclient.NewRetryHTTPClient(conf.APIURL, 30, 3, 3, authInfo, option)
		if err != nil {
			done <- err
			return
		}

		reqUrl := httpClient.BuildURL(path, nil)

		req, err := http.NewRequest("PUT", reqUrl, pr)
		if err != nil {
			done <- err
			return
		}
		req.Header.Set("Content-Type", mw.FormDataContentType())

		resp, err := httpClient.Do(req)
		if err != nil {
			done <- err
			return
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			log.Errorf(ctx, "response from GW was error with StatusCode: %d", resp.StatusCode)
		}
		done <- nil
	}()

	go func() {
		defer cleanutil.Close(ctx, pw, "pipeWriter")
		defer cleanutil.Close(ctx, mw, "multiWriter")

		// part: status
		statusHeader := createPartHeader("status", "text/plain")
		statusPart, err := mw.CreatePart(statusHeader)
		if err != nil {
			xerr := errors.Errorf("unexpected error occurred in creating statusPart: %w", err)
			done <- xerr
			return
		}
		_, err = statusPart.Write([]byte(strconv.Itoa(statusCode)))
		if err != nil {
			xerr := errors.Errorf("unexpected error occurred in writing status to statusPart: %w", err)
			done <- xerr
			return
		}

		// part: headers
		// NOTE: remove Content-Length because it's incorrect in this case.
		delete(headers, convert.KeyContentLength)
		headersHeader := createPartHeader("headers", "application/json")
		headersPart, err := mw.CreatePart(headersHeader)
		if err != nil {
			xerr := errors.Errorf("unexpected error occurred in creating headersPart: %w", err)
			done <- xerr
			return
		}
		headerBytes, err := json.Marshal(headers)
		if err != nil {
			xerr := errors.Errorf(
				"unexpected error occurred in marshaling headers from res.Metadata: %w", err)
			done <- xerr
			return
		}
		if _, err := headersPart.Write(headerBytes); err != nil {
			xerr := errors.Errorf(
				"unexpected error occurred in writing headers to headersPart: %w", err)
			done <- xerr
			return
		}

		// part: body
		bodyHeader := createPartHeader("body", util.ToStringValue(res.ContentType, "text/plain"))
		bodyPart, err := mw.CreatePart(bodyHeader)
		if err != nil {
			xerr := errors.Errorf("unexpected error occurred in creating bodyPart: %w", err)
			done <- xerr
			return
		}
		if body != nil {
			if _, err = io.Copy(bodyPart, body); err != nil {
				xerr := errors.Errorf(
					"unexpected error occurred in writing body to bodyPart: %w", err)
				done <- xerr
				return
			}
		} else {
			if _, err := bodyPart.Write([]byte("")); err != nil {
				xerr := errors.Errorf(
					"unexpected error occurred in writing body to bodyPart: %w", err)
				done <- xerr
				return
			}
		}
	}()

	checkAsync(ctx, conf, path, contents.AsyncARMSToken, done, option)
}

func checkAsync(
	ctx context.Context,
	conf *config.Configuration,
	path string,
	token string,
	sendErr chan error,
	option *http.Client) {

	sendError := <-sendErr
	if sendError != nil {
		log.Error(ctx, "unexpected error occurred in sending async response: ", sendError)
		sendAsyncErrorToARMS(ctx, conf, path, token, "build response of runtime", option)
	}
}

func createPartHeader(name string, contentType string) textproto.MIMEHeader {
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf("form-data; name=%s", name))
	partHeader.Set("Content-Type", contentType)
	return partHeader
}
