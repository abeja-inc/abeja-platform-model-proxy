package preprocess

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strings"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/util"
	"github.com/abeja-inc/platform-model-proxy/util/auth"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
)

// Preprocessor is struct for preprocess.
type Preprocessor struct {
	BaseURL                   string
	ModelRootDir              string
	OrganizationID            string
	ModelID                   string
	ModelVersionID            string
	AuthInfo                  auth.AuthInfo
	DeploymentCodeDownload    *string
	TrainingModelDownload     *string
	TrainingJobID             *string
	TrainingJobDefinitionName *string
	TrainingResultDir         string
}

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

// TrainingJobResJSON is struct for extract `artifacts.complete.uri` from ABEJA-Platform API.
type TrainingJobResJSON struct {
	Artifacts struct {
		Complete struct {
			DownloadURL string `json:"uri"`
		} `json:"complete"`
	} `json:"artifacts"`
}

func (t *TrainingJobResJSON) GetDownloadURL() string {
	return t.Artifacts.Complete.DownloadURL
}

func (_ *TrainingJobResJSON) GetContentType() string {
	return ""
}

// NewPreprocessor returns a Preprocessor.
func NewPreprocessor(ctx context.Context, conf *config.Configuration) (*Preprocessor, error) {
	baseURL := conf.APIURL
	orgID := conf.OrganizationID
	modelID := conf.ModelID
	versionID := conf.ModelVersionID
	authInfo := conf.GetAuthInfo()
	deploymentCodeDownload := conf.DeploymentCodeDownload
	trainingModelDownload := conf.TrainingModelDownload
	trainingJobID := conf.TrainingJobID
	trainingJobDefinitionName := conf.TrainingJobDefinitionName
	workingDir, err := conf.GetWorkingDir()
	if err != nil {
		return nil, errors.Errorf(": %w", err)
	}
	trainingResultDir, err := conf.GetTrainingResultDir()
	if err != nil {
		return nil, errors.Errorf(": %w", err)
	}

	requiredParams := map[string]string{
		"ABEJA_API_URL":          baseURL,
		"ABEJA_ORGANIZATION_ID":  orgID,
		"ABEJA_MODEL_ID":         modelID,
		"ABEJA_MODEL_VERSION_ID": versionID,
		"PLATFORM_AUTH_TOKEN":    authInfo.AuthToken,
	}
	if err := checkRequiredParams(requiredParams); err != nil {
		return nil, errors.Errorf(": %w", err)
	}

	preprocessor := &Preprocessor{
		BaseURL:                   baseURL,
		ModelRootDir:              workingDir,
		OrganizationID:            orgID,
		ModelID:                   modelID,
		ModelVersionID:            versionID,
		AuthInfo:                  authInfo,
		DeploymentCodeDownload:    nil,
		TrainingModelDownload:     nil,
		TrainingJobID:             nil,
		TrainingJobDefinitionName: nil,
		TrainingResultDir:         trainingResultDir,
	}

	if deploymentCodeDownload != "" {
		preprocessor.DeploymentCodeDownload = &deploymentCodeDownload
	}
	if trainingModelDownload != "" {
		preprocessor.TrainingModelDownload = &trainingModelDownload
	}

	if (trainingJobDefinitionName != "" && trainingJobID == "") ||
		(trainingJobDefinitionName == "" && trainingJobID != "") {
		log.Warning(ctx, "TRAINING_JOB_ID and TRAINING_JOB_DEFINITION_NAME must be set.")
	}
	if trainingJobID != "" && trainingJobDefinitionName != "" {
		preprocessor.TrainingJobID = &trainingJobID
		preprocessor.TrainingJobDefinitionName = &trainingJobDefinitionName
	}
	return preprocessor, nil
}

func checkRequiredParams(params map[string]string) error {
	var errKeys []string
	for key, value := range params {
		if value == "" {
			errKeys = append(errKeys, key)
		}
	}
	if len(errKeys) > 0 {
		return errors.Errorf("required parameter(s) missing: %s", strings.Join(errKeys, ", "))
	}
	return nil
}

// Prepare downloads user-model and trained-model if specified.
func (p *Preprocessor) Prepare(ctx context.Context) error {
	downloader, err := util.NewDownloader(p.BaseURL, p.AuthInfo, nil)
	if err != nil {
		return errors.Errorf("failed to make downloader: %w", err)
	}

	if p.DeploymentCodeDownload != nil {
		if err := prepareDeploymentCode(
			ctx, *p.DeploymentCodeDownload, p.ModelRootDir, downloader); err != nil {
			return errors.Errorf(": %w", err)
		}
	} else {
		if err := prepareModel(
			ctx, p.OrganizationID, p.ModelID, p.ModelVersionID,
			p.ModelRootDir, downloader); err != nil {
			return errors.Errorf(": %w", err)
		}
	}

	if p.TrainingModelDownload != nil {
		if err := prepareTrainingModel(
			ctx, *p.TrainingModelDownload,
			p.TrainingResultDir,
			downloader); err != nil {
			return errors.Errorf(": %w", err)
		}
	} else if p.TrainingJobID != nil {
		if err := prepareTrainingJobResult(
			ctx, p.OrganizationID,
			*p.TrainingJobDefinitionName,
			*p.TrainingJobID,
			p.TrainingResultDir,
			downloader); err != nil {
			return errors.Errorf(": %w", err)
		}
	}
	return nil
}

func prepareDeploymentCode(
	ctx context.Context,
	reqPath string,
	destPath string,
	downloader *util.Downloader) error {

	tempFilePath, err := download(ctx, reqPath, downloader, new(SourceResJSON))
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	defer cleanutil.Remove(ctx, tempFilePath)

	if _, err = os.Stat(destPath); err != nil {
		if err = os.Mkdir(destPath, os.ModeDir); err != nil {
			return errors.Errorf(
				"failed to make directory for user-model: %s, error: %w", destPath, err)
		}
	}

	if err = util.Unarchive(tempFilePath, destPath); err != nil {
		return errors.Errorf("failed to unarchive deployment code: %w", err)
	}
	return nil
}

func prepareModel(
	ctx context.Context,
	orgID string,
	modelID string,
	versionID string,
	destPath string,
	downloader *util.Downloader) error {

	reqPath :=
		path.Join("organizations", orgID, "models", modelID, "versions", versionID, "source")
	tempFilePath, err := download(ctx, reqPath, downloader, new(SourceResJSON))
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	defer cleanutil.Remove(ctx, tempFilePath)

	if _, err = os.Stat(destPath); err != nil {
		if err = os.Mkdir(destPath, os.ModeDir); err != nil {
			return errors.Errorf(
				"failed to make directory for user-model: %s, error: %w", destPath, err)
		}
	}

	if err = util.Unarchive(tempFilePath, destPath); err != nil {
		return errors.Errorf("failed to unarchive model version: %w", err)
	}
	return nil
}

func prepareTrainingModel(
	ctx context.Context,
	reqPath string,
	destPath string,
	downloader *util.Downloader) error {

	tempFilePath, err := download(ctx, reqPath, downloader, new(SourceResJSON))
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	defer cleanutil.Remove(ctx, tempFilePath)

	if _, err = os.Stat(destPath); err != nil {
		if err = os.Mkdir(destPath, os.ModeDir); err != nil {
			return errors.Errorf(
				"failed to make directory for training-result: %s, error: %w",
				destPath, err)
		}
	}

	if err = util.Unarchive(tempFilePath, destPath); err != nil {
		return errors.Errorf("failed to unarchive model: %w", err)
	}
	return nil
}

func prepareTrainingJobResult(
	ctx context.Context,
	orgID string,
	jobDefName string,
	jobID string,
	destPath string,
	downloader *util.Downloader) error {

	reqPath :=
		path.Join("organizations", orgID, "training/definitions", jobDefName, "jobs", jobID, "result")
	tempFilePath, err := download(ctx, reqPath, downloader, new(TrainingJobResJSON))
	if err != nil {
		return errors.Errorf(": %w", err)
	}
	defer cleanutil.Remove(ctx, tempFilePath)

	if _, err = os.Stat(destPath); err != nil {
		if err = os.Mkdir(destPath, os.ModeDir); err != nil {
			return errors.Errorf(
				"failed to make directory for training-result: %s, error: %w",
				destPath, err)
		}
	}

	if err = util.Unarchive(tempFilePath, destPath); err != nil {
		return errors.Errorf("failed to unarchive training job result: %w", err)
	}
	return nil
}

func download(
	ctx context.Context,
	reqPath string,
	downloader *util.Downloader,
	decoderRes util.DecoderRes) (string, error) {
	fp, err := ioutil.TempFile("", "model")
	if err != nil {
		return "", errors.Errorf("failed to create temporary file for downloading: %w", err)
	}
	filePath := fp.Name()
	cleanutil.Close(ctx, fp, filePath)

	if _, err = downloader.Download(reqPath, filePath, decoderRes); err != nil {
		cleanutil.Remove(ctx, filePath)
		return "", errors.Errorf("failed to download model: %w", err)
	}
	return filePath, nil
}
