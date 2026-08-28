package restart

import (
	"fmt"

	"chaosd/cli/application"
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewRestartCmd(
	sessionStore session.Store,
	docker docker.DockerProvider,
) *cobra.Command {
	app := application.Application{
		SessionStore:   sessionStore,
		DockerProvider: docker,
	}

	return &cobra.Command{
		Use: "restart",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, app)
		},
		Args: cobra.ExactArgs(2),
	}
}

func run(
	cmd *cobra.Command,
	args []string,
	app application.Application,
) error {
	sessionID := args[0]
	serviceName := args[1]

	results, err := app.RestartService(cmd.Context(), sessionID, serviceName)

	if err != nil {
		return fmt.Errorf("failed to restart service: %v", err)
	}

	var restartFailed bool

	for _, r := range results {
		if r.Err != nil {
			restartFailed = true

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"%s failed to restart\n",
				r.Node.ContainerName,
			)

			continue
		}

		fmt.Fprintf(
			cmd.OutOrStdout(),
			"%s restarted\n",
			r.Node.ContainerName,
		)
	}

	if restartFailed {
		return fmt.Errorf("one or more containers failed to restart")
	}

	return nil
}
