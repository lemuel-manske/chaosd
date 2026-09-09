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
		Use: "delay",
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
		Args: cobra.ExactArgs(4),
	}
}
