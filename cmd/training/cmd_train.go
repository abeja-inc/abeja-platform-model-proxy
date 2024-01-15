package training

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	errors "golang.org/x/xerrors"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
	"github.com/abeja-inc/abeja-platform-model-proxy/version"
)

var confTrain = config.NewConfiguration()

func newCmdTrain() *cobra.Command {
	cmdTrain := &cobra.Command{
		Use:          "train",
		Short:        "execute training-code",
		PreRunE:      setupTrainConfiguration,
		RunE:         execTrain,
		SilenceUsage: true,
	}

	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindOrganizationID,
		cmdutil.BindUserModelRoot,
		cmdutil.BindPlatformAuthToken,
		cmdutil.BindPlatformUserID,
		cmdutil.BindPlatformPersonalAccessToken,
		cmdutil.BindTrainingJobDefinitionName,
		cmdutil.BindTrainingJobDefinitionVersion,
		cmdutil.BindTrainingJobID,
		cmdutil.BindTrainingResultDir,
		cmdutil.BindRuntime,
	}
	if err := cmdutil.BindOptions(cmdTrain, options); err != nil {
		// NOTE: This cobra/viper's error don't occur basically...
		log.Warningf(procCtx, "unexpected error occurred when binding command line options: "+log.ErrorFormat, err)
	}

	return cmdTrain
}

func setupTrainConfiguration(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&confTrain); err != nil {
		return err
	}
	return validateTrainConfiguration()
}

func validateTrainConfiguration() error {
	var notSetRequires []string
	if confTrain.OrganizationID == "" {
		notSetRequires = append(notSetRequires, "abeja_organization_id")
	}
	if len(notSetRequires) > 0 {
		return errors.Errorf("require flag(s) %s not set", strings.Join(notSetRequires, ", "))
	}
	if err := cmdutil.ValidateAuthParts(
		confTrain.PlatformAuthToken,
		confTrain.PlatformUserID,
		confTrain.PlatformPersonalAccessToken); err != nil {
		return errors.Errorf(": %w", err)
	}
	if confTrain.TrainingJobDefinitionVersion != 0 && confTrain.TrainingJobDefinitionName == "" {
		return errors.New("flag training_job_definition_name needs when you set training_job_definition_version")
	}
	if confTrain.TrainingJobID != "" && confTrain.TrainingJobDefinitionName == "" {
		return errors.New("flag training_job_definition_name needs when you set training_job_id")
	}
	return nil
}

func execTrain(cmd *cobra.Command, args []string) error {
	log.Infof(procCtx, "abeja-runner version: [%s] start training.", version.Version)
	log.Debugf(procCtx, "configuration: %s", confTrain.String())
	return Train(procCtx, &confTrain)
}
