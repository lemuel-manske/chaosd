package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chaosd",
		Short: "Chaosd is a laboratory for distributed systems",
		Run:   run,
	}
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println(cmd.OutOrStdout(), "Chaosd command executed")
}
