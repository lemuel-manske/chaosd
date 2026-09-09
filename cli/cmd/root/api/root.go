package api

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chaosd",
		Short: "Chaosd is a laboratory for distributed systems",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(cmd.OutOrStdout(), "Chaosd command executed")

			return nil
		},
	}
}
