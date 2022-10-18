package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	errors "golang.org/x/xerrors"

	simplejson "github.com/bitly/go-simplejson"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/convert"
	"github.com/abeja-inc/platform-model-proxy/entity"
	"github.com/abeja-inc/platform-model-proxy/util"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
)

// === Protocol
//
// |--------------------------------------------------------------|---------------|
// | Header                                                       | Body          |
// |--------------------------------------------------------------|---------------|
// | MAGIC              | VERSION (byte) | LENGTH of Body(uint32) | JSON (string) |
// |--------------------|----------------|------------------------|---------------|
// | 0xAB | 0xE9 | 0xA0 | 0x01           | (4 bytes)              | ...           |
// |--------------------|----------------|------------------------|---------------|
const magic0 = 0xAB
const magic1 = 0xE9
const magic2 = 0xA0
const version = 0x01

// Header is header of protocol for communicate to runtime.
type Header struct {
	Magic   [3]byte
	Version byte
	Length  uint32
}

// DatalakeSourceResJSON is part of response of `GET file info`
type DatalakeSourceResJSON struct {
	DownloadURL string                 `json:"download_url"`
	ContentType string                 `json:"content_type"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (d *DatalakeSourceResJSON) GetDownloadURL() string {
	return d.DownloadURL
}

func (d *DatalakeSourceResJSON) GetContentType() string {
	return d.ContentType
}

func FromRequest(request *entity.ContentList) (Header, []byte, error) {
	b, err := json.Marshal(request)
	if err != nil {
		return Header{}, []byte{}, errors.Errorf("json encode error: %w", err)
	}

	header := Header{
		Magic:   [3]byte{magic0, magic1, magic2},
		Version: version,
		Length:  uint32(len(b)),
	}
	return header, b, nil
}

func ToResponse(bodyBuff []byte, conf *config.Configuration) (entity.Response, error) {
	if bytes.Equal(bodyBuff, []byte{}) {
		return entity.Response{}, errors.Errorf("communication with runtime")
	}
	var body entity.Response
	if err := json.Unmarshal(bodyBuff, &body); err != nil {
		return body, errors.Errorf("Read IPC response body error: %w", err)
	}

	metadata := make(map[string]string)
	if body.Metadata != nil {
		for key, value := range *(body.Metadata) {
			metadata[key] = value
		}
	}
	if conf.ModelID != "" {
		metadata["X-Abeja-Model-Id"] = conf.ModelID
	}
	if conf.ModelVersion != "" {
		metadata["X-Abeja-Model-Version"] = url.PathEscape(conf.ModelVersion)
	}
	if conf.DeploymentID != "" {
		metadata["X-Abeja-Deployment-Id"] = conf.DeploymentID
	}
	if conf.ServiceID != "" {
		metadata["X-Abeja-Service-Id"] = conf.ServiceID
	}
	body.Metadata = &metadata
	return body, nil
}

func FromInput(
	ctx context.Context,
	conf *config.Configuration,
	option *http.Client) (*entity.ContentList, error) {

	input := conf.Input
	if input == "" {
		return &entity.ContentList{
			AsyncRequestID: conf.RunID,             // TODO temporary
			AsyncARMSToken: conf.PlatformAuthToken, // TODO temporary
		}, nil
	}

	json, err := simplejson.NewJson([]byte(input))
	if err != nil {
		return nil, errors.Errorf("failed to parse INPUT: %w", err)
	}

	if _, err := json.Array(); err == nil {
		return convertToJsonContents(ctx, input, conf)
	}

	datalakePath, err := json.Get("$datalake:1").String()
	if err != nil {
		return convertToJsonContents(ctx, input, conf)
	}

	return convertToFileContents(ctx, conf, datalakePath, option)
}

func convertToJsonContents(
	ctx context.Context,
	input string,
	conf *config.Configuration) (*entity.ContentList, error) {

	contentType := "application/json"
	ext := util.GetExtension(ctx, contentType)
	filePath, err := convert.ToFileFromBody(input, ext, conf.RequestedDataDir)
	if err != nil {
		return nil, errors.Errorf("failed to create temporary json file: %w", err)
	}

	return buildContents(contentType, conf.RunID, conf.PlatformAuthToken, filePath, nil), nil
}

func convertToFileContents(
	ctx context.Context,
	conf *config.Configuration,
	datalakePath string,
	option *http.Client) (*entity.ContentList, error) {

	contentType, filePath, metadata, err := getFileFromDatalake(ctx, conf, datalakePath, option)
	if err != nil {
		return nil, errors.Errorf("failed to create temporary file: %w", err)
	}

	return buildContents(contentType, conf.RunID, conf.PlatformAuthToken, filePath, metadata), nil
}

func getFileFromDatalake(
	ctx context.Context,
	conf *config.Configuration,
	datalakePath string,
	option *http.Client) (string, string, map[string]interface{}, error) {

	filePath := filepath.Join(conf.RequestedDataDir, "uploaded_file")
	reqPath := fmt.Sprintf("/channels/%s", datalakePath)

	downloader, err := util.NewDownloader(conf.APIURL, conf.GetAuthInfo(), option)
	if err != nil {
		return "", "", nil, errors.Errorf("failed to parse ABEJA_API_URL: %w", err)
	}

	var contentType string
	var datalakeSourceResJSON DatalakeSourceResJSON
	contentType, err =
		downloader.Download(reqPath, filePath, &datalakeSourceResJSON)
	if err != nil {
		cleanutil.Remove(ctx, filePath)
		return "", "", nil, errors.Errorf("failed to download file from datalake: %w", err)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return contentType, filePath, datalakeSourceResJSON.Metadata, nil
}

func buildContents(
	contentType, reqID, token, contentPath string, metadata map[string]interface{},
) *entity.ContentList {
	content := &entity.Content{
		Path:     &contentPath,
		Metadata: metadata,
	}
	contents := []*entity.Content{content}
	cl := &entity.ContentList{
		Method:         "POST",
		ContentType:    contentType,
		Contents:       contents,
		AsyncRequestID: reqID, // TODO temporary
		AsyncARMSToken: token, // TODO temporary
	}
	return cl
}

func FromOutput(conf *config.Configuration) (string, error) {
	output := conf.Output
	if output == "" {
		return "", nil
	}

	json, err := simplejson.NewJson([]byte(output))
	if err != nil {
		return "", errors.Errorf("failed to parse OUTPUT: %w", err)
	}

	val, exist := json.CheckGet("$datalake:1")
	if !exist {
		return "", nil
	}
	return val.String()
}
