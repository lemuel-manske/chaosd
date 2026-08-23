package cmd

import (
	"fmt"
	"os"

	"chaosd/cli/cmd/load"
	"chaosd/cli/cmd/ps"
	"chaosd/cli/cmd/restart"
	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	dockerProvider := docker.NewDockerProvider()

	rootCmd.AddCommand(load.NewLoadCmd(dockerProvider))
	rootCmd.AddCommand(ps.NewPsCmd(dockerProvider))
	rootCmd.AddCommand(restart.NewRestartCmd(dockerProvider))
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
