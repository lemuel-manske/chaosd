package api

import (
	"fmt"

	"chaosd/cli/application"

	"github.com/spf13/cobra"
)

func NewLoadCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use:  "load",
		Args: cobra.ExactArgs(1),

		Short: "Load a Docker Compose project",

		Long: `Load a Docker Compose project into Chaosd and create a new session.

Chaosd validates the Compose file, discovers the project's containers and
network topology, and returns a session ID for subsequent commands.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			composeFilePath := args[0]

			sessionID, err := app.Load(cmd.Context(), composeFilePath)

			if err != nil {
				return fmt.Errorf("failed to load compose file: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), sessionID)

			return nil
		},
	}
}
