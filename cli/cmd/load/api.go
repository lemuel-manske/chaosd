package load

import (
	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
)

func NewLoadCmd(docker docker.DockerProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "load",
		Short: "Load a compose file",
		Long:  `Load a Docker compose file into chaosd daemon.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadCmd(cmd, args, docker)
		},
		Args: cobra.ExactArgs(1),
	}
}
