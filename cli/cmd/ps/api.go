package ps

import (
	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
)

func NewPsCmd(docker docker.DockerProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "List the topology of the loaded compose file",
		Long:  `List the topology of the loaded compose file, including the service name, container ID, container name, and state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPsCmd(cmd, args, docker)
		},
		Args: cobra.ExactArgs(1),
	}
}
