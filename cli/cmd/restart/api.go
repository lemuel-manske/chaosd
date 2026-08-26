package restart

import (
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewRestartCmd(
	sessionStore *session.Store,
	docker docker.DockerProvider,
) *cobra.Command {
	return &cobra.Command{
		Use: "restart",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestartCmd(cmd, args, sessionStore, docker)
		},
		Args: cobra.ExactArgs(2),
	}
}
