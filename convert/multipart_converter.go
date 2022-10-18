package convert

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/util"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

type multipartConverter struct{}

func (conv *multipartConverter) IsTarget(
	ctx context.Context, method string, contentType string) bool {

	if method != http.MethodPost && method != http.MethodPut {
		return false
	}

	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Warningf(ctx, "failed to parse `Content-Type` header: "+log.ErrorFormat, err)
		return false
	}
	if mt == "multipart/form-data" {
		return true
	}
	return false
}

func (conv *multipartConverter) ToContent(
	ctx context.Context,
	r *http.Request,
	conf *config.Configuration) (*entity.ContentList, error) {

	baseContentType := r.Header.Get("Content-Type")
	var contents []*entity.Content

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, errors.Errorf(": %w", err)
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		contentType := part.Header.Get("Content-Type")
		formName := part.FormName()
		fileName := part.FileName()
		log.Debugf(
			ctx, "In part: Content-Type: [%s], FormName: [%s], FileName: [%s]",
			contentType, formName, fileName)

		ext := util.GetExtension(ctx, contentType)
		tmpFilePath, err := ToFileFromReader(part, ext, conf.RequestedDataDir)
		cleanutil.Close(
			ctx,
			part,
			fmt.Sprintf(
				"part: Content-Type:[%s], FormName:[%s], FileName:[%s]",
				contentType, formName, fileName))
		if err != nil {
			return nil, errors.Errorf(": %w", err)
		}

		content := &entity.Content{
			Path: &tmpFilePath,
		}
		if contentType != "" {
			content.ContentType = &contentType
		}
		if fileName != "" {
			content.FileName = &fileName
		}
		if formName != "" {
			content.FormName = &formName
		}
		contents = append(contents, content)
	}
	return &entity.ContentList{
		Method:      r.Method,
		ContentType: baseContentType,
		Contents:    contents,
		Ctx:         ctx,
	}, nil
}

func (conv *multipartConverter) FromResponse(ctx context.Context, res entity.Response) (
	statusCode int, headers map[string]string, body *os.File, err error) {
	return 0, nil, nil, errors.New("not implemented error")
}
