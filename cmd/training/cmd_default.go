package training

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	errors "golang.org/x/xerrors"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
	"github.com/abeja-inc/platform-model-proxy/version"
)

var (
	procCtx     context.Context
	confDefault = config.NewConfiguration()
)

func newCmdRoot(ctx context.Context) *cobra.Command {
	procCtx = ctx
	cmdRoot := &cobra.Command{
		Use:          "training",
		Short:        "download training-code and train",
		PreRunE:      setupDefaultConfiguration,
		RunE:         execDefault,
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
	if err := cmdutil.BindOptions(cmdRoot, options); err != nil {
		// NOTE: This cobra/viper's error don't occur basically...
		log.Warningf(
			procCtx,
			"unexpected error occurred when binding command line options: "+log.ErrorFormat,
			err)
	}

	return cmdRoot
}

func setupDefaultConfiguration(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&confDefault); err != nil {
		return errors.Errorf(": %w", err)
	}
	return validateDefaultConfiguration()
}

func validateDefaultConfiguration() error {
	var notSetRequires []string
	if confDefault.OrganizationID == "" {
		notSetRequires = append(notSetRequires, "abeja_organization_id")
	}
	if confDefault.TrainingJobDefinitionName == "" {
		notSetRequires = append(notSetRequires, "training_job_definition_name")
	}
	if len(notSetRequires) > 0 {
		return errors.Errorf("require flag(s) %s not set", strings.Join(notSetRequires, ", "))
	}
	if err := cmdutil.ValidateAuthParts(
		confDefault.PlatformAuthToken,
		confDefault.PlatformUserID,
		confDefault.PlatformPersonalAccessToken); err != nil {
		return errors.Errorf(": %w", err)
	}
	if err := cmdutil.ValidateTrainingJobDefinitionVersion(confDefault.TrainingJobDefinitionVersion); err != nil {
		return errors.Errorf(": %w", err)
	}
	return nil
}

func execDefault(cmd *cobra.Command, args []string) error {
	log.Infof(procCtx, "abeja-runner version: [%s] start download and train.", version.Version)
	log.Debugf(procCtx, "configuration: %s", confDefault.String())
	if err := Download(procCtx, &confDefault); err != nil {
		return errors.Errorf(": %w", err)
	}
	return Train(procCtx, &confDefault)
}

func InitTrainCommand(ctx context.Context) *cobra.Command {
	cmdRoot := newCmdRoot(ctx)
	cmdRoot.AddCommand(newCmdDownload())
	cmdRoot.AddCommand(newCmdTrain())
	return cmdRoot
}
