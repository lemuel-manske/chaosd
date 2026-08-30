package restart

import (
	"fmt"

	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewRestartCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use: "restart",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])
			serviceName := args[1]

			results, err := app.RestartService(
				cmd.Context(),
				sessionID,
				serviceName,
			)

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
		},
		Args: cobra.ExactArgs(2),
	}
}
