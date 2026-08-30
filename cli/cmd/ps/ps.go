package ps

import (
	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewPsCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use: "ps",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])

			t, err := app.GetTopology(cmd.Context(), sessionID)

			if err != nil {
				return err
			}

			t.Print(cmd.OutOrStdout())

			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}
