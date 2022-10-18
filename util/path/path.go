package path

import (
	"os"
	"path/filepath"
	"strings"

	errors "golang.org/x/xerrors"
)

// DefaultTrainingResultDir is name of directory that stored model of training-result.
const DefaultTrainingResultDir = "abejainc_training_result"

func GetWorkingDir(userRoot string) (string, error) {
	if strings.HasPrefix(userRoot, "/") {
		return userRoot, nil
	}
	current, err := os.Getwd()
	if err != nil {
		return "", errors.Errorf(": %w", err)
	}
	currentAbs, err := filepath.Abs(current)
	if err != nil {
		return "", errors.Errorf(": %w", err)
	}
	if userRoot == "" {
		return currentAbs, nil
	}
	return filepath.Join(currentAbs, userRoot), nil
}

func GetTrainingResultDir(trainingResultDir string, userRoot string) (string, error) {
	if strings.HasPrefix(trainingResultDir, "/") {
		return trainingResultDir, nil
	}
	workingDir, err := GetWorkingDir(userRoot)
	if err != nil {
		return "", errors.Errorf(": %w", err)
	}
	if trainingResultDir == "" {
		return filepath.Join(workingDir, DefaultTrainingResultDir), nil
	}
	return filepath.Join(workingDir, trainingResultDir), nil
}
