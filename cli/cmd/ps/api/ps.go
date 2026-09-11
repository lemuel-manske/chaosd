package api

import (
	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewPsCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use:  "ps",
		Args: cobra.ExactArgs(1),

		Short: "Show the current session topology",

		Long: `Show the current containers and state of a Chaosd session.

Chaosd rebuilds the topology from the Docker environment so the output
reflects the current state of the Compose project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])

			t, err := app.GetTopology(cmd.Context(), sessionID)

			if err != nil {
				return err
			}

			t.Print(cmd.OutOrStdout())

			return nil
		},
	}
}
