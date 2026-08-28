package load

import (
	"chaosd/cli/application"
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"
	"fmt"

	"github.com/spf13/cobra"
)

func NewLoadCmd(
	sessionStore session.Store,
	dockerProvider docker.DockerProvider,
) *cobra.Command {
	app := application.Application{
		SessionStore:   sessionStore,
		DockerProvider: dockerProvider,
	}

	return &cobra.Command{
		Use: "load",
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
	composeFilePath := args[0]

	sessionID, err := app.Load(cmd.Context(), composeFilePath)

	if err != nil {
		return fmt.Errorf("failed to load compose file: %v", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), sessionID)

	return nil
}
