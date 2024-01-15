package cmd

import (
	"context"

	"github.com/spf13/cobra"

	batchcmd "github.com/abeja-inc/abeja-platform-model-proxy/cmd/batch"
	servecmd "github.com/abeja-inc/abeja-platform-model-proxy/cmd/service"
	tensorboardcmd "github.com/abeja-inc/abeja-platform-model-proxy/cmd/tensorboard"
	traincmd "github.com/abeja-inc/abeja-platform-model-proxy/cmd/training"
	log "github.com/abeja-inc/abeja-platform-model-proxy/util/logging"
	"github.com/abeja-inc/abeja-platform-model-proxy/version"
)

var procCtx context.Context

func newCmdRoot() *cobra.Command {
	cmdRoot := &cobra.Command{
		Use:          "abeja-runner",
		Short:        "abeja-runner version: " + version.Version,
		RunE:         execDefault,
		SilenceUsage: true,
	}
	return cmdRoot
}

func execDefault(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func initCommand() *cobra.Command {

	cmdRoot := newCmdRoot()

	serveCmd := servecmd.InitServeCommand(procCtx)
	cmdRoot.AddCommand(serveCmd)

	trainCmd := traincmd.InitTrainCommand(procCtx)
	cmdRoot.AddCommand(trainCmd)

	batchCmd := batchcmd.InitBatchCommand(procCtx)
	cmdRoot.AddCommand(batchCmd)

	tensorBoardCmd := tensorboardcmd.InitTensorBoardCommand(procCtx)
	cmdRoot.AddCommand(tensorBoardCmd)

	return cmdRoot
}

func Execute(ctx context.Context) int {
	var state int = 0
	procCtx = ctx
	cmd := initCommand()
	if err := cmd.Execute(); err != nil {
		log.Warningf(procCtx, "error occurred: "+log.ErrorFormat, err)
		state = 1
	}
	return state
}
