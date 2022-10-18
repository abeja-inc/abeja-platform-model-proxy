package tensorboard

import "context"

type MockDownloader struct {
	downloadErrs []error
}

func (d MockDownloader) Download(ctx context.Context, reqPath string, destPath string) error {
	if len(d.downloadErrs) == 0 {
		return nil
	} else {
		err := d.downloadErrs[0]
		d.downloadErrs = d.downloadErrs[1:] //nolint // SA4005: ineffective assignment to field MockDownloader.downloadErrs
		return err
	}
}
