package ps

import (
	"chaosd/cli/application"
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewPsCmd(
	sessionStore session.Store,
	docker docker.DockerProvider,
) *cobra.Command {
	app := application.Application{
		SessionStore:   sessionStore,
		DockerProvider: docker,
	}

	return &cobra.Command{
		Use: "ps",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, app)
		},
		Args: cobra.ExactArgs(1),
	}
}

func run(
	cmd *cobra.Command,
	args []string,
	app application.Application,
) error {
	sessionID := args[0]

	t, err := app.GetTopology(cmd.Context(), sessionID)

	if err != nil {
		return err
	}

	t.Print(cmd.OutOrStdout())

	return nil
}
