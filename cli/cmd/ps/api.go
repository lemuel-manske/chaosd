package ps

import (
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewPsCmd(
	sessionStore *session.Store,
	docker docker.DockerProvider,
) *cobra.Command {
	return &cobra.Command{
		Use: "ps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPsCmd(cmd, args, sessionStore, docker)
		},
		Args: cobra.ExactArgs(1),
	}
}
