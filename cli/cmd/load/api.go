package load

import (
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewLoadCmd(
	sessionStore session.Store,
	docker docker.DockerProvider,
) *cobra.Command {
	return &cobra.Command{
		Use: "load",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadCmd(cmd, args, sessionStore, docker)
		},
		Args: cobra.ExactArgs(1),
	}
}
