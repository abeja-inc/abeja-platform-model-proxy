package service

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	errors "golang.org/x/xerrors"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
	cleanutil "github.com/abeja-inc/platform-model-proxy/util/clean"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
	"github.com/abeja-inc/platform-model-proxy/version"
)

var confRun = config.NewConfiguration()

func newCmdRun() *cobra.Command {
	cmdRun := &cobra.Command{
		Use:          "run",
		Short:        "run user-code",
		PreRunE:      setupRunConfiguration,
		RunE:         execRun,
		PostRun:      teardownRun,
		SilenceUsage: true,
	}

	// bind options with viper
	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindUserModelRoot,
		cmdutil.BindOrganizationID,
		cmdutil.BindModelVersion,
		cmdutil.BindDeploymentID,
		cmdutil.BindServiceID,
		cmdutil.BindRuntime,
		cmdutil.BindPort,
		cmdutil.BindHealthCheckPort,
		cmdutil.BindTrainingResultDir,
	}
	if err := cmdutil.BindOptions(cmdRun, options); err != nil {
		// NOTE: This cobra/viper's error don't occur basically...
		log.Warningf(
			procCtx,
			"unexpected error occurred when binding command line options: "+log.ErrorFormat,
			err)
	}

	return cmdRun
}

func setupRunConfiguration(cmd *cobra.Command, args []string) error {
	if err := viper.Unmarshal(&confRun); err != nil {
		return err
	}
	return validateRunConfiguration()
}

func validateRunConfiguration() error {
	if err := cmdutil.ValidatePortNumber(confRun.Port); err != nil {
		return err
	}
	if confRun.ServiceID != "" {
		if confRun.OrganizationID == "" || confRun.DeploymentID == "" {
			return errors.New(
				"flag abeja_organization_id and abeja_deployment_id need when you set abeja_service_id")
		}
	}
	if confRun.GetListenAddress() == confRun.GetHealthCheckAddress() {
		return errors.New("port and healthcheck_port should be different value")
	}
	return nil
}

func execRun(cmd *cobra.Command, args []string) error {
	log.Infof(procCtx, "abeja-runner version: [%s] start serving.", version.Version)
	return run(procCtx, &confRun, false)
}

func teardownRun(cmd *cobra.Command, args []string) {
	cleanutil.RemoveAll(procCtx, confRun.RequestedDataDir)
}
