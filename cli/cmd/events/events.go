package events

import (
	"fmt"

	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewEventsCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use: "events",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])

			events, err := app.ListEvents(cmd.Context(), sessionID)

			if err != nil {
				return fmt.Errorf("failed to list events: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "TIME\tTYPE\tTARGET")

			for _, event := range events {
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"%s\t%s\t%v\n",
					event.CreatedAt.Format("15:04:05"),
					event.Type,
					event.Data,
				)
			}

			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}
