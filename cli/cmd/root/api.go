package root

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chaosd",
		Short: "Chaosd is a laboratory for distributed systems",
		Run:   runChaosd,
	}
}
