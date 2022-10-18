package util

import (
	"os"

	errors "golang.org/x/xerrors"
)

func ValidatePortNumber(port int) error {
	if port < 1024 || port > 65535 {
		return errors.Errorf("port [%d] must be greater than 1023 and less than 65536", port)
	}
	return nil
}

func ValidateTrainingJobDefinitionVersion(version int) error {
	if version < 1 {
		return errors.Errorf("training_job_definition_version [%d] must be greater than 0", version)
	}
	return nil
}

func ValidateAuthParts(authToken string, userID string, userToken string) error {
	if authToken == "" {
		if userID == "" || userToken == "" {
			return errors.Errorf(
				"%s or (%s and %s) need but not set.",
				"platform_auth_token",
				"abeja_platform_user_id",
				"abeja_platform_personal_access_token")
		}
	}
	return nil
}

func ValidateServingCode(
	deploymentCodeDownload, organizationID, modelID, modelVersionID string) error {

	if deploymentCodeDownload != "" {
		return nil
	}
	if organizationID != "" && modelID != "" && modelVersionID != "" {
		return nil
	}
	return errors.Errorf(
		"%s or (%s and %s and %s) need but not set.",
		"abeja_deployment_code_download",
		"abeja_organization_id",
		"abeja_model_id",
		"abeja_model_version_id")
}

func ValidateTrainedModel(
	trainingModelDownload, organizationID, trainingJobDefinitionName, trainingJobID string) error {

	if trainingModelDownload == "" &&
		trainingJobDefinitionName == "" &&
		trainingJobID == "" {
		return nil
	}
	if trainingModelDownload != "" {
		return nil
	}
	if organizationID != "" && trainingJobDefinitionName != "" && trainingJobID != "" {
		return nil
	}
	return errors.Errorf(
		"When using the training result, please specify %s or (%s, %s and %s).",
		"abeja_training_model_download",
		"abeja_organization_id",
		"training_job_definition_name",
		"training_job_id")
}

func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.Errorf(": %w", err)
		}
	}
	return nil
}
