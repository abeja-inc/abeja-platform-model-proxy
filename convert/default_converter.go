package convert

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strconv"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/entity"
	"github.com/abeja-inc/abeja-platform-model-proxy/util"
	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
	"github.com/abeja-inc/abeja-platform-model-proxy/version"
)

type defaultConverter struct{}

func (conv *defaultConverter) IsTarget(
	ctx context.Context, method string, contentType string) bool {

	if method != DummyMethodForResponse {
		if method == http.MethodGet {
			// When Http-Method is GET, Content-Type is ignored
			return true
		}

		if method != http.MethodPost && method != http.MethodPut {
			return false
		}
	}

	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Warningf(ctx, "failed to parse `Content-Type` header: "+log.ErrorFormat, err)
		return false
	}
	if mt == "multipart/form-data" {
		return false
	}
	return true
}

func (conv *defaultConverter) ToContent(
	ctx context.Context,
	r *http.Request,
	conf *config.Configuration) (*entity.ContentList, error) {
	if r.Method == http.MethodGet {
		body := r.URL.RawQuery
		tmpFilePath, err := ToFileFromBody(body, util.DefaultExt, conf.RequestedDataDir)
		if err != nil {
			return nil, errors.Errorf(": %w", err)
		}

		content := &entity.Content{
			Path: &tmpFilePath,
		}
		contents := []*entity.Content{content}
		return &entity.ContentList{
			// consider query-string as x-www-form-urlencoded
			Method:      r.Method,
			ContentType: "application/x-www-form-urlencoded",
			Contents:    contents,
			Ctx:         ctx,
		}, nil
	}

	contentType := r.Header.Get("Content-Type")
	bufbody := new(bytes.Buffer)
	_, err := bufbody.ReadFrom(r.Body)
	if err != nil {
		return nil, errors.Errorf(": %w", err)
	}
	body := bufbody.String()

	ext := util.GetExtension(ctx, contentType)
	tmpFilePath, err := ToFileFromBody(body, ext, conf.RequestedDataDir)
	if err != nil {
		return nil, errors.Errorf(": %w", err)
	}
	content := &entity.Content{
		Path: &tmpFilePath,
	}
	contents := []*entity.Content{content}
	return &entity.ContentList{
		Method:      r.Method,
		ContentType: contentType,
		Contents:    contents,
		Ctx:         ctx,
	}, nil
}

func (conv *defaultConverter) FromResponse(ctx context.Context, res entity.Response) (
	statusCode int, headers map[string]string, body *os.File, err error) {

	statusCode = http.StatusOK
	contentType := "application/json"
	headers = make(map[string]string)
	headers[KeyContentType] = contentType
	headers[KeyAbejaProxyVersion] = version.Version
	headers[KeyContentLength] = "0"
	// SAMPv2 limits the number of concurrent requests by LimitListener,
	// but it accepts requests from multiple clients at the same time.
	// Keepalive is disabled because it can't process requests until
	// the previous connection is closed, which causes a wait time.
	headers[KeyConnection] = "close"

	if res.StatusCode != nil {
		statusCode = *res.StatusCode
	} else {
		log.Debug(ctx, "no status_code in response of user-model. set 200")
	}

	if res.ErrMsg != nil {
		return 0, headers, nil, &ConverterError{
			Msg:        *res.ErrMsg,
			StatusCode: statusCode,
			Err:        nil,
			frame:      errors.Caller(0),
		}
	}

	if res.ContentType != nil {
		headers[KeyContentType] = *res.ContentType
	} else {
		log.Debug(ctx, "no content-type in response of user-model. use default(application/json).")
	}

	if res.Metadata != nil {
		for key, value := range *res.Metadata {
			headers[key] = value
		}
	}

	if res.Path == nil {
		log.Debug(ctx, "no path in response of user-model.")
		fp, _ := ioutil.TempFile("", "")
		return statusCode, headers, fp, nil
	}

	fileInfo, err := os.Stat(*res.Path)
	if err != nil {
		log.Errorf(ctx, "failed to get file information: "+log.ErrorFormat, err)
		return 0, headers, nil, &ConverterError{
			Msg:        "unexpected error",
			StatusCode: http.StatusServiceUnavailable,
			Err:        err,
			frame:      errors.Caller(0),
		}
	}
	headers[KeyContentLength] = strconv.FormatInt(fileInfo.Size(), 10)

	fp, err := FromFile(*res.Path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warningf(ctx, "file [%s] that specified in response of user-model not exist.", *res.Path)
		} else {
			log.Errorf(ctx, "failed to load file: "+log.ErrorFormat, err)
		}
		return 0, headers, nil, &ConverterError{
			Msg:        "unexpected error",
			StatusCode: http.StatusServiceUnavailable,
			Err:        err,
			frame:      errors.Caller(0),
		}
	}

	return statusCode, headers, fp, nil
}
