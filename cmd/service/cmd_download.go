package service

import (
	"fmt"
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
		Short:        "download user-code (and model if necessary)",
		PreRunE:      setupDownloadConfiguration,
		RunE:         execDownload,
		SilenceUsage: true,
	}

	// bind options with viper
	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindOrganizationID,
		cmdutil.BindModelID,
		cmdutil.BindModelVersionID,
		cmdutil.BindUserModelRoot,
		cmdutil.BindPlatformAuthToken,
		cmdutil.BindDeploymentCodeDownload,
		cmdutil.BindTrainingModelDownload,
		cmdutil.BindTrainingJobID,
		cmdutil.BindTrainingJobDefinitionName,
		cmdutil.BindTrainingResultDir,
	}
	if err := cmdutil.BindOptions(cmdDownload, options); err != nil {
		// NOTE: This cobra/viper's error don't occur basically...
		log.Warningf(procCtx, "unexpected error occurred when binding command line options: "+log.ErrorFormat, err)
	}

	return cmdDownload
}

func setupDownloadConfiguration(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&confDownload); err != nil {
		return err
	}
	return validateDownloadConfiguration()
}

func validateDownloadConfiguration() error {
	var notSetRequires []string
	if confDownload.OrganizationID == "" {
		notSetRequires = append(notSetRequires, "abeja_organization_id")
	}
	if confDownload.ModelID == "" {
		notSetRequires = append(notSetRequires, "abeja_model_id")
	}
	if confDownload.ModelVersionID == "" {
		notSetRequires = append(notSetRequires, "abeja_model_version_id")
	}
	if confDownload.PlatformAuthToken == "" {
		notSetRequires = append(notSetRequires, "platform_auth_token")
	}
	if len(notSetRequires) > 0 {
		return errors.Errorf("require flag(s) %s not set", strings.Join(notSetRequires, ", "))
	}
	return nil
}

func execDownload(cmd *cobra.Command, args []string) error {
	log.Info(
		procCtx,
		fmt.Sprintf("abeja-runner version: [%s] start download serving code.", version.Version))
	return download(procCtx, &confDownload)
}
