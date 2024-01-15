package training

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/util"
	cleanutil "github.com/abeja-inc/abeja-platform-model-proxy/util/clean"
)

// SourceResJSON is struct for extract `download_uri` from ABEJA-Platform API.
type SourceResJSON struct {
	DownloadURL string `json:"download_uri"`
}

func (s *SourceResJSON) GetDownloadURL() string {
	return s.DownloadURL
}

func (_ *SourceResJSON) GetContentType() string {
	return ""
}

func Download(ctx context.Context, conf *config.Configuration) error {
	downloader, err := util.NewDownloader(conf.APIURL, conf.GetAuthInfo(), nil)
	if err != nil {
		return errors.Errorf("failed to make downloader: %w", err)
	}

	reqPath := path.Join(
		"organizations", conf.OrganizationID,
		"training", "definitions", conf.TrainingJobDefinitionName,
		"versions", strconv.Itoa(conf.TrainingJobDefinitionVersion), "source")
	tempFilePath, err := download(ctx, reqPath, downloader)
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	defer cleanutil.Remove(ctx, tempFilePath)

	destPath, err := conf.GetWorkingDir()
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	if _, err = os.Stat(destPath); err != nil {
		if err = os.Mkdir(destPath, os.ModeDir); err != nil {
			return errors.Errorf(
				"failed to make directory for user-model: %s, error: %w", destPath, err)
		}
	}

	if err = util.Unarchive(tempFilePath, destPath); err != nil {
		return errors.Errorf("failed to unarchive model: %w", err)
	}
	return nil
}

func download(
	ctx context.Context,
	reqPath string,
	downloader *util.Downloader) (string, error) {

	fp, err := ioutil.TempFile("", "train")
	if err != nil {
		return "", errors.Errorf("failed to create temporary file for download: %w", err)
	}
	filePath := fp.Name()
	cleanutil.Close(ctx, fp, filePath)

	if _, err = downloader.Download(reqPath, filePath, new(SourceResJSON)); err != nil {
		cleanutil.Remove(ctx, filePath)
		return "", errors.Errorf("failed to download training-code: %w", err)
	}
	return filePath, nil
}
