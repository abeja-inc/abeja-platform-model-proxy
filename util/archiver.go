package util

import (
	"os"

	errors "golang.org/x/xerrors"

	"github.com/mholt/archiver"
)

// Unarchive unarchives archived file to `destPath`.
func Unarchive(filePath string, destPath string) error {
	fp, err := os.Open(filePath)
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	unarchiver, err := archiver.ByHeader(fp)
	if err != nil {
		// NOTE: archiver can auto-discriminate only zip, tar, and rar.
		// On the other hand, platform has zip or tar.gz as corresponding format.
		// So, if auto-discrimination fails, it is regarded as tar.gz.
		// If the uploaded file has an extension, we do not have to do this...
		unarchiver = archiver.NewTarGz()
	}

	if err = unarchiver.Unarchive(filePath, destPath); err != nil {
		return errors.Errorf(": %w", err)
	}
	return nil
}
