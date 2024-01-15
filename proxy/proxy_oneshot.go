package proxy

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/convert"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/util"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	httpclient "github.com/abeja-inc/platform-model-proxy/util/http"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

func TransportOneshotMessage(
	ctx context.Context,
	conf *config.Configuration,
	socketFilePath string,
	errOnBoot chan int,
	notifyFromMain chan int,
	notifyToMain chan int,
	option *http.Client) {

	defer close(notifyToMain)

	// open unix domain socket to runtime
	conn, err := net.Dial("unix", socketFilePath)
	if err != nil {
		log.Errorf(ctx, "Failed to dial to runtime: "+log.ErrorFormat, err)
		notifyToMain <- 1
		close(errOnBoot)
		return
	}
	defer cleanutil.Close(ctx, conn, socketFilePath)

	respReceiver := make(chan []byte, 1)
	defer close(respReceiver)

	// parse INPUT and get content of input from datalake
	cl, err := FromInput(ctx, conf, option)
	if err != nil {
		log.Errorf(ctx, "INPUT parse error: "+log.ErrorFormat, err)
		// Depending on the timing, close of notifyToMain will be detected before close of errOnBoot.
		// Sending error status to notifyToMain will solve the problem,
		// so we can do that.
		// (the log will change depending on whether errOnBoot or notifyToMain is detected,
		//  but we'll allow that to happen for now).
		// I thought about not having to include errOnBoot in this function,
		// but I don't want to expand the scope of the change, which could create other bugs,
		// so I try to keep the changes to a minimum.
		notifyToMain <- 1
		close(errOnBoot)
		return
	}
	if cl.Contents == nil {
		log.Info(ctx, "Input resource not specified.")
	}
	defer deleteTempFiles(ctx, cl, nil)

	// check OUTPUT
	datalakeChannelID, err := FromOutput(conf)
	if err != nil {
		log.Errorf(ctx, "OUTPUT parse error:"+log.ErrorFormat, err)
		notifyToMain <- 1
		close(errOnBoot)
		return
	}
	if datalakeChannelID == "" {
		log.Info(ctx, "Output datalake channel not specified.")
	}

	header, body, err := FromRequest(cl)
	if err != nil {
		log.Errorf(ctx, "json marshaling error: "+log.ErrorFormat, err)
		notifyToMain <- 1
		close(errOnBoot)
		return
	}

	// send request to runtime
	if err = binary.Write(conn, binary.BigEndian, header); err != nil {
		log.Errorf(ctx, "Write IPC request header error: "+log.ErrorFormat, err)
		notifyToMain <- 1
		close(errOnBoot)
		return
	}

	if _, err = conn.Write(body); err != nil {
		log.Errorf(ctx, "Write IPC request body error: "+log.ErrorFormat, err)
		notifyToMain <- 1
		close(errOnBoot)
		return
	}

	// receive response from runtime
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

	// wait response or signal
	select {
	case bodyBuff := <-respReceiver:
		res, err := ToResponse(bodyBuff, conf)
		if err != nil {
			log.Errorf(ctx, log.ErrorFormat, err)
			notifyToMain <- 1
		} else {
			if res.Path != nil {
				defer cleanutil.Remove(ctx, *res.Path)
			}
			handleResult(ctx, conf, res, datalakeChannelID, notifyToMain, option)
		}
	case <-notifyFromMain:
		log.Debug(ctx, "signal received")
	}
}

func handleResult(
	ctx context.Context,
	conf *config.Configuration,
	res entity.Response,
	datalakeChannelID string,
	notifyToMain chan int,
	option *http.Client) {

	status, headers, body, err := convert.FromResponse(ctx, res)
	if err != nil {
		if convertError, ok := err.(*convert.ConverterError); ok {
			log.Warning(ctx, convertError.Msg)
		} else {
			log.Warningf(ctx, log.ErrorFormat, err)
		}
		notifyToMain <- 1
		return
	}

	if datalakeChannelID != "" {
		if res.Path == nil {
			// if OUTPUT is specified but there is nothing to upload, it logs a warning.
			log.Warning(ctx, "runtime didn't return body.")
		} else {
			err := uploadResult(
				ctx, conf, datalakeChannelID, headers, body, res.ContentType, option)
			if err != nil {
				log.Errorf(ctx, log.ErrorFormat, err)
				notifyToMain <- 1
				return
			}
		}
	}

	if status > 299 {
		log.Warningf(ctx, "runtime returned error status %d.", status)
		notifyToMain <- 1
	}
}

func uploadResult(
	ctx context.Context,
	conf *config.Configuration,
	datalakeChannelID string,
	headers map[string]string,
	body *os.File,
	contentType *string,
	option *http.Client) error {

	httpClient, err :=
		httpclient.NewRetryHTTPClient(conf.APIURL, 30, 3, conf.GetAuthInfo(), option)
	if err != nil {
		return err
	}

	reqUrl := httpClient.BuildURL(
		fmt.Sprintf("/channels/%s/upload", datalakeChannelID),
		map[string]interface{}{"conflict_target": "filename"})

	req, err := http.NewRequest("POST", reqUrl, body)
	if err != nil {
		return err
	}

	for k, v := range headers {
		lk := strings.ToLower(k)
		if lk == "content-type" || strings.HasPrefix(lk, "x-abeja-meta-") {
			req.Header.Set(k, v)
		}
	}

	fileName := buildFileName(ctx, conf.RunID, contentType)
	req.Header.Set("x-abeja-meta-filename", fileName)
	fileinfo, err := body.Stat()
	if err != nil {
		return err
	}
	req.ContentLength = fileinfo.Size()

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var errMsg string
		msgBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = string(msgBytes)
		}
		return fmt.Errorf(
			"response from Datalake was error with StatusCode: %d, body: %s",
			resp.StatusCode,
			string(errMsg),
		)
	}
	return nil
}

func buildFileName(ctx context.Context, runID string, contentType *string) string {
	var ext string
	if contentType == nil || *contentType == "" {
		ext = ""
	} else {
		ext = util.GetExtension(ctx, *contentType)
	}
	return fmt.Sprintf("%s_0%s", runID, ext)
}
