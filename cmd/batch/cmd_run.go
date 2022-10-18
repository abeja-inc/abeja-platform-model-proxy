package batch

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdutil "github.com/abeja-inc/platform-model-proxy/cmd/util"
	"github.com/abeja-inc/platform-model-proxy/config"
	log "github.com/abeja-inc/platform-model-proxy/util/logging"
	"github.com/abeja-inc/platform-model-proxy/version"
)

var confRun = config.NewConfiguration()

func newCmdRun() *cobra.Command {
	cmdRun := &cobra.Command{
		Use:          "run",
		Short:        "run batch code",
		PreRunE:      setupRunConfiguration,
		RunE:         execRun,
		SilenceUsage: true,
	}

	options := []func(*cobra.Command) error{
		cmdutil.BindAbejaAPIURL,
		cmdutil.BindOrganizationID,
		cmdutil.BindModelID,
		cmdutil.BindModelVersion,
		cmdutil.BindDeploymentID,
		cmdutil.BindServiceID,
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
	if strings.Contains(confRun.Input, "$datalake:1") || confRun.Output != "" {
		if err := cmdutil.ValidateAuthParts(
			confRun.PlatformAuthToken,
			confRun.PlatformUserID,
			confRun.PlatformPersonalAccessToken); err != nil {
			return err
		}
	}
	return nil
}

func execRun(cmd *cobra.Command, args []string) error {
	log.Info(
		procCtx,
		fmt.Sprintf("abeja-runner version: [%s] start running batch code.", version.Version))
	log.Debugf(procCtx, "configuration: %s", confRun.String())
	return run(procCtx, &confRun)
}
