package restart

import (
	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
)

func NewRestartCmd(docker docker.DockerProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestartCmd(cmd, args, docker)
		},
		Args: cobra.ExactArgs(2),
	}
}
