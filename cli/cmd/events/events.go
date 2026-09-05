package events

import (
	"fmt"

	"chaosd/cli/application"
	"chaosd/cli/internal/event"
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

			for _, storedEvent := range events {
				for _, target := range eventTargets(storedEvent) {
					fmt.Fprintf(
						cmd.OutOrStdout(),
						"%s\t%s\t%s\n",
						storedEvent.CreatedAt.Format("15:04:05"),
						storedEvent.Type,
						target,
					)
				}
			}

			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}

func eventTargets(storedEvent event.Event) []string {
	switch data := storedEvent.Data.(type) {
	case event.RestartEventData:
		return []string{data.ServiceName}
	case event.PartitionAppliedEventData:
		return []string{data.NodeAName, data.NodeBName}
	case event.HealAppliedEventData:
		return []string{data.NodeAName, data.NodeBName}
	default:
		return nil
	}
}
