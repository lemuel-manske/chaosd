package partition

import (
	"fmt"

	"chaosd/cli/application"
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewPartitionCmd(
	sessionStore session.Store,
	dockerProvider docker.DockerProvider,
	networkManager network.Manager,
) *cobra.Command {
	app := application.Application{
		SessionStore:   sessionStore,
		DockerProvider: dockerProvider,
		NetworkManager: networkManager,
	}

	return &cobra.Command{
		Use: "partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, app)
		},
		Args: cobra.ExactArgs(3),
	}
}

func run(
	cmd *cobra.Command,
	args []string,
	app application.Application,
) error {
	sessionID := args[0]

	nodeAName := args[1]
	nodeBName := args[2]

	err := app.Partition(cmd.Context(), sessionID, nodeAName, nodeBName)

	if err != nil {
		return err
	}

	fmt.Fprintf(
		cmd.OutOrStdout(),
		"%s and %s partitioned\n",
		nodeAName,
		nodeBName,
	)

	return nil
}
