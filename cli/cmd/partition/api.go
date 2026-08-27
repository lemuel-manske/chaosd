package partition

import (
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
	return &cobra.Command{
		Use: "partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPartitionCmd(cmd, args, sessionStore, dockerProvider, networkManager)
		},
		Args: cobra.ExactArgs(3),
	}
}
