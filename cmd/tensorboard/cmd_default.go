package tensorboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdutil "github.com/abeja-inc/abeja-platform-model-proxy/cmd/util"
	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
	"github.com/abeja-inc/abeja-platform-model-proxy/version"
)

var (
	procCtx     context.Context
	confDefault = config.NewConfiguration()
)

func newCmdRoot(ctx context.Context) *cobra.Command {
	procCtx = ctx
	cmdRoot := &cobra.Command{
		Use:          "tensorboard",
		Short:        "download user training-results",
		PreRunE:      setupDefaultConfiguration,
		RunE:         execDefault,
		SilenceUsage: true,
	}

	// bind options with viper
	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindOrganizationID,
		cmdutil.BindPlatformAuthToken,
		cmdutil.BindTrainingJobDefinitionName,
		cmdutil.BindTrainingJobIDList,
		cmdutil.BindTensorboardID,
		cmdutil.BindMountTargetDir,
	}
	if err := cmdutil.BindOptions(cmdRoot, options); err != nil {
		// NOTE: This cobra/viper's error don't occur basically...
		log.Warningf(procCtx, "unexpected error occurred when binding command line options: "+log.ErrorFormat, err)
	}

	return cmdRoot
}

func setupDefaultConfiguration(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&confDefault); err != nil {
		return err
	}
	return validateDefaultConfiguration()
}

func validateDefaultConfiguration() error {
	var notSetRequires []string
	if confDefault.OrganizationID == "" {
		notSetRequires = append(notSetRequires, "abeja_organization_id")
	}
	if confDefault.PlatformAuthToken == "" {
		notSetRequires = append(notSetRequires, "platform_auth_token")
	}
	if confDefault.TrainingJobDefinitionName == "" {
		notSetRequires = append(notSetRequires, "training_job_definition_name")
	}
	if confDefault.TrainingJobIDS == "" {
		notSetRequires = append(notSetRequires, "training_job_ids")
	}
	if confDefault.TensorboardID == "" {
		notSetRequires = append(notSetRequires, "tensorboard_id")
	}
	if len(notSetRequires) > 0 {
		return errors.Errorf("require flag(s) %s not set", strings.Join(notSetRequires, ", "))
	}
	return nil
}

func execDefault(cmd *cobra.Command, args []string) error {
	log.Info(
		procCtx, fmt.Sprintf("abeja-runner version: [%s] start download & serve.", version.Version))
	return run(procCtx, &confDefault)
}

func InitTensorBoardCommand(ctx context.Context) *cobra.Command {
	return newCmdRoot(ctx)
}
