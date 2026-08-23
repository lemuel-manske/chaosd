package ps

import (
	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
)

func NewPsCmd(docker docker.DockerProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPsCmd(cmd, args, docker)
		},
		Args: cobra.ExactArgs(1),
	}
}
