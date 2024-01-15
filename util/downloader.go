package util

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/platform-model-proxy/util/auth"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	httpclient "github.com/abeja-inc/platform-model-proxy/util/http"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

const downloaderTimeout = 600 // 10 minutes

type DecoderRes interface {
	GetDownloadURL() string
	GetContentType() string
}

// Downloader is struct for download entities.
type Downloader struct {
	Client *httpclient.RetryClient
}

// NewDownloader returns `Downloader`.
func NewDownloader(baseURL string, authInfo auth.AuthInfo, option *http.Client) (*Downloader, error) {
	httpClient, err := httpclient.NewRetryHTTPClient(baseURL, downloaderTimeout, 10, authInfo, option)
	if err != nil {
		return nil, errors.Errorf("failed to build http client: %w", err)
	}

	return &Downloader{Client: httpClient}, nil
}

// Download downloads the entity from the download_uri contained in the apipath response
// to the `destPath`.
func (d Downloader) Download(apiPath string, destPath string, decoderRes DecoderRes) (string, error) {

	err := d.Client.GetJson(apiPath, nil, decoderRes)
	if err != nil {
		log.Error(context.TODO(), "failed to request meta data")
		return "", errors.Errorf("failed to request to %s: %w", apiPath, err)
	}

	signedURL := decoderRes.GetDownloadURL()
	contentType := decoderRes.GetContentType()
	resp2, err := d.Client.GetThrough(signedURL)
	if err != nil {
		log.Error(context.TODO(), "failed to request download data")
		return "", errors.Errorf("failed to request to %s: %w", signedURL, err)
	}
	defer cleanutil.Close(context.TODO(), resp2.Body, fmt.Sprintf("response body of %s", signedURL))
	if resp2.StatusCode != http.StatusOK {
		msg, _ := ioutil.ReadAll(resp2.Body)
		log.Errorf(context.TODO(), "failed to download with status %d\n", resp2.StatusCode)
		return "", errors.Errorf(
			"failed to download from %s with status %d, body = [%s]",
			signedURL, resp2.StatusCode, msg)
	}

	fp, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Error(context.TODO(), "failed to open download data")
		return "", errors.Errorf("failed to open %s: %w", destPath, err)
	}
	defer cleanutil.Close(context.TODO(), fp, destPath)
	if _, err = io.Copy(fp, resp2.Body); err != nil {
		log.Error(context.TODO(), "failed to copy response-body")
		return "", errors.Errorf("failed to copying response-body to %s: %w", destPath, err)
	}

	return contentType, nil
}
