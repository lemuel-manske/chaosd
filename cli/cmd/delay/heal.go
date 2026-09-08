package delay

import (
	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewHealCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use: "heal",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])

			nodeAName := args[1]
			nodeBName := args[2]

			err := app.Heal(cmd.Context(), sessionID, nodeAName, nodeBName) // TODO: heal by fault ID

			if err != nil {
				return err
			}

			cmd.Printf("%s and %s healed\n", nodeAName, nodeBName)

			return nil
		},
		Args: cobra.ExactArgs(3),
	}
}
