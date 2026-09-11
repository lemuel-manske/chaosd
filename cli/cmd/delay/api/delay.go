package api

import (
	"fmt"
	"time"

	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewDelayCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use:  "delay",
		Args: cobra.ExactArgs(4),

		Short: "Add network delay between two nodes",
		Long: `Add network delay between two running nodes in a Chaosd session.

The delay is applied to traffic between the selected nodes across their
shared network paths and remains active until the fault is healed.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])

			nodeAName := args[1]
			nodeBName := args[2]

			delay, err := time.ParseDuration(args[3])

			if err != nil {
				return fmt.Errorf("invalid delay duration: %v", err)
			}

			faultID, err := app.Delay(cmd.Context(), sessionID, nodeAName, nodeBName, delay)

			if err != nil {
				return err
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"%s\n",
				faultID,
			)

			return nil
		},
	}
}
