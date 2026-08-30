package cmd

import (
	"fmt"
	"os"

	"chaosd/cli/application"
	"chaosd/cli/cmd/events"
	"chaosd/cli/cmd/load"
	"chaosd/cli/cmd/partition"
	"chaosd/cli/cmd/ps"
	"chaosd/cli/cmd/restart"
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/session"

	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	dockerProvider := docker.NewDockerProvider()

	sessionStore, err := session.NewDefaultStore()

	if err != nil {
		msg := fmt.Errorf("failed to create session store: %v", err)

		panic(msg)
	}

	networkManager := network.NewManager(
		network.NewLinuxFirewallInjector(),
	)

	app := application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	rootCmd.AddCommand(events.NewEventsCmd(*app))
	rootCmd.AddCommand(load.NewLoadCmd(*app))
	rootCmd.AddCommand(partition.NewHealCmd(*app))
	rootCmd.AddCommand(partition.NewPartitionCmd(*app))
	rootCmd.AddCommand(ps.NewPsCmd(*app))
	rootCmd.AddCommand(restart.NewRestartCmd(*app))
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
