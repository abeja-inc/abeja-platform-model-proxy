package batch

import (
	"context"
	"fmt"
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
		Use:          "batch",
		Short:        "download batch code and run",
		PreRunE:      setupDefaultConfiguration,
		RunE:         execDefault,
		SilenceUsage: true,
	}

	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindOrganizationID,
		cmdutil.BindModelID,
		cmdutil.BindModelVersion,
		cmdutil.BindModelVersionID,
		cmdutil.BindDeploymentID,
		cmdutil.BindDeploymentCodeDownload,
		cmdutil.BindServiceID,
		cmdutil.BindTrainingModelDownload,
		cmdutil.BindTrainingJobDefinitionName,
		cmdutil.BindTrainingJobID,
		cmdutil.BindRunID,
		cmdutil.BindPlatformAuthToken,
		cmdutil.BindPlatformUserID,
		cmdutil.BindPlatformPersonalAccessToken,
		cmdutil.BindRuntime,
		cmdutil.BindUserModelRoot,
		cmdutil.BindTrainingResultDir,
		cmdutil.BindInput,
		cmdutil.BindOutput,
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
		return err
	}
	return validateDefaultConfiguration()
}

func validateDefaultConfiguration() error {
	var notSetRequires []string
	if confDefault.OrganizationID == "" {
		notSetRequires = append(notSetRequires, "abeja_organization_id")
	}
	if len(notSetRequires) > 0 {
		return errors.Errorf("require flag(s) %s not set", strings.Join(notSetRequires, ", "))
	}

	if err := cmdutil.ValidateAuthParts(
		confDefault.PlatformAuthToken,
		confDefault.PlatformUserID,
		confDefault.PlatformPersonalAccessToken); err != nil {
		return err
	}
	if err := cmdutil.ValidateServingCode(
		confDefault.DeploymentCodeDownload,
		confDefault.OrganizationID,
		confDefault.ModelID,
		confDefault.ModelVersionID); err != nil {
		return err
	}
	if err := cmdutil.ValidateTrainedModel(
		confDefault.TrainingModelDownload,
		confDefault.OrganizationID,
		confDefault.TrainingJobDefinitionName,
		confDefault.TrainingJobID); err != nil {
		return err
	}
	return nil
}

func execDefault(cmd *cobra.Command, args []string) error {
	log.Info(
		procCtx,
		fmt.Sprintf("abeja-runner version: [%s] start download and run batch.", version.Version))
	log.Debugf(procCtx, "configuration: %s", confDefault.String())
	if err := download(procCtx, &confDefault); err != nil {
		return errors.Errorf(": %w", err)
	}
	return run(procCtx, &confDefault)
}

func InitBatchCommand(ctx context.Context) *cobra.Command {
	cmdRoot := newCmdRoot(ctx)
	cmdRoot.AddCommand(newCmdRun())

	return cmdRoot
}
