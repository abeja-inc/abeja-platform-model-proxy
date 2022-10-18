package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	errors "golang.org/x/xerrors"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
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
		Use:          "service",
		Short:        "download user model/training-result and run model",
		PreRunE:      setupDefaultConfiguration,
		RunE:         execDefault,
		PostRun:      teardownDefault,
		SilenceUsage: true,
	}

	// bind options with viper
	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindOrganizationID,
		cmdutil.BindModelID,
		cmdutil.BindModelVersion,
		cmdutil.BindModelVersionID,
		cmdutil.BindDeploymentID,
		cmdutil.BindServiceID,
		cmdutil.BindUserModelRoot,
		cmdutil.BindPlatformAuthToken,
		cmdutil.BindDeploymentCodeDownload,
		cmdutil.BindTrainingModelDownload,
		cmdutil.BindTrainingJobID,
		cmdutil.BindTrainingJobDefinitionName,
		cmdutil.BindRuntime,
		cmdutil.BindPort,
		cmdutil.BindHealthCheckPort,
		cmdutil.BindTrainingResultDir,
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
	if confDefault.ModelID == "" {
		notSetRequires = append(notSetRequires, "abeja_model_id")
	}
	if confDefault.ModelVersionID == "" {
		notSetRequires = append(notSetRequires, "abeja_model_version_id")
	}
	if confDefault.PlatformAuthToken == "" {
		notSetRequires = append(notSetRequires, "platform_auth_token")
	}
	if len(notSetRequires) > 0 {
		return errors.Errorf("require flag(s) %s not set", strings.Join(notSetRequires, ", "))
	}
	if err := cmdutil.ValidatePortNumber(confDefault.Port); err != nil {
		return err
	}
	if confDefault.ServiceID != "" && confDefault.DeploymentID == "" {
		return errors.New("flag abeja_deployment_id needs when you set abeja_service_id")
	}
	if confDefault.GetListenAddress() == confDefault.GetHealthCheckAddress() {
		return errors.New("port and healthcheck_port should be different value")
	}
	return nil
}

func execDefault(cmd *cobra.Command, args []string) error {
	log.Info(
		procCtx, fmt.Sprintf("abeja-runner version: [%s] start download & serve.", version.Version))
	return run(procCtx, &confDefault, true)
}

func teardownDefault(cmd *cobra.Command, args []string) {
	cleanutil.RemoveAll(procCtx, confDefault.RequestedDataDir)
}

func InitServeCommand(ctx context.Context) *cobra.Command {

	cmdRoot := newCmdRoot(ctx)
	cmdRoot.AddCommand(newCmdDownload()) // for download only
	cmdRoot.AddCommand(newCmdRun())      // for run only

	return cmdRoot
}
