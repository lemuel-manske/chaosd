package api

import (
	"fmt"

	"chaosd/cli/application"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func NewPartitionCmd(
	app application.Application,
) *cobra.Command {
	return &cobra.Command{
		Use:  "partition",
		Args: cobra.ExactArgs(3),

		Short: "Partition two nodes",
		Long: `Create a network partition between two running nodes in a Chaosd session.

Traffic between the selected nodes is blocked in both directions across
their shared network paths until the fault is healed.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := session.SessionID(args[0])

			nodeAName := args[1]
			nodeBName := args[2]

			faultID, err := app.Partition(cmd.Context(), sessionID, nodeAName, nodeBName)

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
