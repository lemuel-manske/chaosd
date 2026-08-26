package cmd

import (
	"fmt"
	"os"

	"chaosd/cli/cmd/load"
	"chaosd/cli/cmd/ps"
	"chaosd/cli/cmd/restart"
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	dockerProvider := docker.NewDockerProvider()

	sessionStore, err := session.DefaultStore()

	if err != nil {
		msg := fmt.Errorf("failed to create session store: %v", err)

		panic(msg)
	}

	rootCmd.AddCommand(load.NewLoadCmd(sessionStore, dockerProvider))
	rootCmd.AddCommand(ps.NewPsCmd(sessionStore, dockerProvider))
	rootCmd.AddCommand(restart.NewRestartCmd(sessionStore, dockerProvider))
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
