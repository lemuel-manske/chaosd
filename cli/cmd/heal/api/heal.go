package api

import (
	"fmt"

	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewHealCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use:  "heal",
		Args: cobra.ExactArgs(2),

		Short: "Heal an active fault",

		Long: `Heal an active fault in a Chaosd session.

The fault is identified by its fault ID and is reverted using the effects
recorded when the fault was originally applied.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])
			faultID := session.FaultID(args[1])

			err := app.Heal(cmd.Context(), sessionID, faultID)

			if err != nil {
				return err
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"fault %s healed\n",
				faultID,
			)

			return nil
		},
	}
}
