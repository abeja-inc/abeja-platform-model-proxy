package tensorboard

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/util"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

const retryCount uint64 = 5

type CompletedJSON struct {
	URL string `json:"uri"`
}

type ArtifactsJSON struct {
	Completed CompletedJSON `json:"complete"`
}

type ResultResJSON struct {
	Artifacts ArtifactsJSON `json:"artifacts"`
}

func (s *ResultResJSON) GetDownloadURL() string {
	return s.Artifacts.Completed.URL
}

func (_ *ResultResJSON) GetContentType() string {
	return ""
}

type Downloader interface {
	Download(ctx context.Context, reqPath string, destPath string) error
}

type TrainingJobResultDownloader struct {
	apiPath    string
	downloader *util.Downloader
}

func (d TrainingJobResultDownloader) Download(ctx context.Context, reqPath string, destPath string) error {
	if _, err := d.downloader.Download(reqPath, destPath, new(ResultResJSON)); err != nil {
		cleanutil.Remove(ctx, destPath)
		return errors.Errorf("failed to download training-code: %w", err)
	}
	return nil
}

func removeDuplicates(l []string) []string {
	var set []string
	duplicated := map[string]bool{}
	for _, item := range l {
		if _, ok := duplicated[item]; !ok {
			set = append(set, item)
			duplicated[item] = true
		}
	}
	return set
}

func run(ctx context.Context, conf *config.Configuration) error {
	downloader, err := util.NewDownloader(conf.APIURL, conf.GetAuthInfo(), nil)
	if err != nil {
		return errors.Errorf("failed to make downloader: %w", err)
	}
	trainingJobResultDownloader := TrainingJobResultDownloader{
		apiPath:    conf.APIURL,
		downloader: downloader,
	}

	trainingJobIds := removeDuplicates(strings.Split(conf.TrainingJobIDS, ","))
	for _, trainingJobId := range trainingJobIds {
		operation := func() error {
			return downloadAndUnarchive(ctx, conf, trainingJobResultDownloader, util.Unarchive, trainingJobId)
		}
		// TODO: handle the case of client errors like Forbidden, NotFound
		if err := withExponentialBackOffRetry(ctx, operation, retryCount); err != nil {
			// TODO: need to update TensorBoard status as failed.
			return err
		}
	}
	log.Infof(ctx, "succeeded to ready for training jobs, %s", conf.TrainingJobIDS)
	return nil
}

func withExponentialBackOffRetry(
	ctx context.Context,
	operation func() error,
	maxRetry uint64,
) error {
	b := backoff.NewExponentialBackOff()
	notify := func(err error, t time.Duration) {
		log.Warningf(ctx, "retrying due to error: %v", err)
	}
	if maxRetry > 0 {
		return backoff.RetryNotify(operation, backoff.WithMaxRetries(b, maxRetry), notify)
	}
	return backoff.RetryNotify(operation, b, notify)
}

func downloadAndUnarchive(
	ctx context.Context,
	conf *config.Configuration,
	downloader Downloader,
	unarchive func(filePath string, destination string) error,
	trainingJobId string,
) error {
	log.Infof(ctx, "start download and unarchive training job result of %s", trainingJobId)

	tmpDir, err := ioutil.TempDir("", "tensorboard")
	if err != nil {
		return errors.Errorf("failed to create temporary file for download: %w", err)
	}
	archiveFilePath := filepath.Join(tmpDir, fmt.Sprintf("archive-%d", time.Now().UnixNano()))
	defer cleanutil.RemoveAll(ctx, tmpDir)

	reqPath := filepath.Join(
		"organizations", conf.OrganizationID,
		"training", "definitions", conf.TrainingJobDefinitionName,
		"jobs", trainingJobId, "result")

	if err := downloader.Download(ctx, reqPath, archiveFilePath); err != nil {
		return errors.Errorf("failed to download archive: %w", err)
	}

	destPath := filepath.Join(conf.MountTargetDir, "tensorboards", conf.TensorboardID, "training_jobs", trainingJobId)
	if _, err = os.Stat(destPath); err == nil {
		// if directory of destPath exists, remove them
		// this is not efficient, but cannot make sure the existing ones are not defected.
		if err := os.RemoveAll(destPath); err != nil {
			return errors.Errorf("failed to delete existing directory for training-result: %s, error: %w",
				destPath, err)
		}
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return errors.Errorf(
			"failed to make directory for training-result: %s, error: %w",
			destPath, err)
	}

	if err = unarchive(archiveFilePath, destPath); err != nil {
		return errors.Errorf("failed to unarchive training job result: %w", err)
	}
	log.Infof(ctx, "succeeded to download and unarchive training job result of %s", trainingJobId)
	return nil
}
