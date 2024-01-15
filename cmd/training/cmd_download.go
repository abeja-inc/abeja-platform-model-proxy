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

var confDownload = config.NewConfiguration()

func newCmdDownload() *cobra.Command {
	cmdDownload := &cobra.Command{
		Use:          "download",
		Short:        "download training-code",
		PreRunE:      setupDownloadConfiguration,
		RunE:         execDownload,
		SilenceUsage: true,
	}

	// bind options with viper
	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindOrganizationID,
		cmdutil.BindUserModelRoot,
		cmdutil.BindPlatformAuthToken,
		cmdutil.BindPlatformUserID,
		cmdutil.BindPlatformPersonalAccessToken,
		cmdutil.BindTrainingJobDefinitionName,
		cmdutil.BindTrainingJobDefinitionVersion,
	}
	if err := cmdutil.BindOptions(cmdDownload, options); err != nil {
		// NOTE: This cobra/viper's error don't occur basically...
		log.Warningf(procCtx, "unexpected error occurred when binding command line options: "+log.ErrorFormat, err)
	}

	return cmdDownload
}

func setupDownloadConfiguration(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&confDownload); err != nil {
		return errors.Errorf(": %w", err)
	}
	return validateDownloadConfiguration()
}

func validateDownloadConfiguration() error {
	var notSetRequires []string
	if confDownload.OrganizationID == "" {
		notSetRequires = append(notSetRequires, "abeja_organization_id")
	}
	if confDownload.TrainingJobDefinitionName == "" {
		notSetRequires = append(notSetRequires, "training_job_definition_name")
	}
	if len(notSetRequires) > 0 {
		return errors.Errorf("require flag(s) %s not set", strings.Join(notSetRequires, ", "))
	}
	if err := cmdutil.ValidateAuthParts(
		confDownload.PlatformAuthToken,
		confDownload.PlatformUserID,
		confDownload.PlatformPersonalAccessToken); err != nil {
		return errors.Errorf(": %w", err)
	}
	if err := cmdutil.ValidateTrainingJobDefinitionVersion(confDownload.TrainingJobDefinitionVersion); err != nil {
		return errors.Errorf(": %w", err)
	}
	return nil
}

func execDownload(cmd *cobra.Command, args []string) error {
	log.Infof(procCtx, "abeja-runner version: [%s] start download train-code.", version.Version)
	log.Debugf(procCtx, "configuration: %s", confDownload.String())
	return Download(procCtx, &confDownload)
}
