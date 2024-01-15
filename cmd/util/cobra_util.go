package util

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	pathutil "github.com/abeja-inc/abeja-platform-model-proxy/util/path"
)

func bindLocalIntOption(
	cmd *cobra.Command,
	flagKey string,
	defValue int,
	desc string,
	viperKey string,
	envKey string) error {

	cmd.Flags().Int(flagKey, defValue, desc)
	if err := viper.BindEnv(viperKey, envKey); err != nil {
		return err
	}
	if err := viper.BindPFlag(viperKey, cmd.Flags().Lookup(flagKey)); err != nil {
		return err
	}
	return nil
}

func bindLocalStringOption(
	cmd *cobra.Command,
	flagKey string,
	defValue string,
	desc string,
	viperKey string,
	envKey string) error {

	cmd.Flags().String(flagKey, defValue, desc)
	if err := viper.BindEnv(viperKey, envKey); err != nil {
		return err
	}
	if err := viper.BindPFlag(viperKey, cmd.Flags().Lookup(flagKey)); err != nil {
		return err
	}
	return nil
}

func BindAbejaAPIURL(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_api_url", config.DefaultAbejaAPIURL, "base url of abeja-api",
		"APIURL", "ABEJA_API_URL")
}

func BindOrganizationID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_organization_id", "", "identifier of organization",
		"OrganizationID", "ABEJA_ORGANIZATION_ID")
}

func BindModelID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_model_id", "", "identifier of model",
		"ModelID", "ABEJA_MODEL_ID")
}

func BindModelVersion(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_model_version", "", "model version",
		"ModelVersion", "ABEJA_MODEL_VERSION")
}

func BindModelVersionID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_model_version_id", "", "identifier of model version",
		"ModelVersionID", "ABEJA_MODEL_VERSION_ID")
}

func BindDeploymentID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_deployment_id", "", "identifier of deployment",
		"DeploymentID", "ABEJA_DEPLOYMENT_ID")
}

func BindServiceID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_service_id", "", "identifier of service",
		"ServiceID", "ABEJA_SERVICE_ID")
}

func BindDeploymentCodeDownload(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_deployment_code_download", "", "deployment code download path",
		"DeploymentCodeDownload", "ABEJA_DEPLOYMENT_CODE_DOWNLOAD")
}

func BindTrainingModelDownload(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_training_model_download", "", "training model download path",
		"TrainingModelDownload", "ABEJA_TRAINING_MODEL_DOWNLOAD")
}

func BindUserModelRoot(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_user_model_root", "",
		"root path of the directory where the user model is located",
		"UserModelRoot", "ABEJA_USER_MODEL_ROOT")
}

func BindPlatformAuthToken(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "platform_auth_token", "", "authentication token for platform",
		"PlatformAuthToken", "PLATFORM_AUTH_TOKEN")
}

func BindPlatformUserID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_platform_user_id", "", "identifier of user",
		"PlatformUserID", "ABEJA_PLATFORM_USER_ID")
}

func BindPlatformPersonalAccessToken(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_platform_personal_access_token", "", "personal access token of user",
		"PlatformPersonalAccessToken", "ABEJA_PLATFORM_PERSONAL_ACCESS_TOKEN")
}

func BindTrainingJobID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "training_job_id", "", "identifier of training job",
		"TrainingJobID", "TRAINING_JOB_ID")
}

func BindTrainingJobIDList(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "training_job_ids", "", "comma separated list of identifier of training job",
		"TrainingJobIDS", "TRAINING_JOB_IDS")
}

func BindTrainingJobDefinitionName(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "training_job_definition_name", "", "name of training job definition",
		"TrainingJobDefinitionName", "TRAINING_JOB_DEFINITION_NAME")
}

func BindTrainingJobDefinitionVersion(cmd *cobra.Command) error {
	return bindLocalIntOption(
		cmd, "training_job_definition_version", 0, "version of training job definition",
		"TrainingJobDefinitionVersion", "TRAINING_JOB_DEFINITION_VERSION")
}

func BindTensorboardID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "tensorboard_id", "", "identifier of tensorboard",
		"TensorboardID", "TENSORBOARD_ID")
}

func BindRunID(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_run_id", "", "identifier of run", "RunID", "ABEJA_RUN_ID")
}

func BindRuntime(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_runtime", config.DefaultRuntime, "runtime language of inference service",
		"Runtime", "ABEJA_RUNTIME")
}

func BindPort(cmd *cobra.Command) error {
	return bindLocalIntOption(
		cmd, "port", config.DefaultHTTPListenPort, "listen port of service", "Port", "PORT")
}

func BindHealthCheckPort(cmd *cobra.Command) error {
	return bindLocalIntOption(
		cmd, "healthcheck_port", config.DefaultHealthCheckListenPort,
		"listen port of health check", "HealthCheckPort", "HEALTHCHECK_PORT")
}

func BindInput(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "input", "", "input data", "Input", "INPUT")
}

func BindOutput(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "output", "", "destination information of output", "Output", "OUTPUT")
}

func BindTrainingResultDir(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "abeja_training_result_dir", pathutil.DefaultTrainingResultDir,
		"directory for placing training-result", "TrainingResultDir", "ABEJA_TRAINING_RESULT_DIR")
}

func BindMountTargetDir(cmd *cobra.Command) error {
	return bindLocalStringOption(
		cmd, "mount_target_dir", config.DefaultMountTargetDir,
		"directory to mount shared file system", "MountTargetDir", "ABEJA_MOUNT_TARGET_DIR")
}

func BindOptions(cmd *cobra.Command, options []func(*cobra.Command) error) error {
	for _, option := range options {
		if err := option(cmd); err != nil {
			return err
		}
	}
	return nil
}
