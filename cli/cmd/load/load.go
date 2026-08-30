package load

import (
	"fmt"

	"chaosd/cli/application"

	"github.com/spf13/cobra"
)

func NewLoadCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use: "load",
		RunE: func(cmd *cobra.Command, args []string) error {
			composeFilePath := args[0]

			sessionID, err := app.Load(cmd.Context(), composeFilePath)

			if err != nil {
				return fmt.Errorf("failed to load compose file: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), sessionID)

			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}
