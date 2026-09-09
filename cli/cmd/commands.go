// Package cmd provides the command line interface for chaosd.
package cmd

import (
	"fmt"
	"os"

	"chaosd/cli/adapters"
	"chaosd/cli/application"
	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/network"

	events_api "chaosd/cli/cmd/events/api"
	heal_api "chaosd/cli/cmd/heal/api"
	load_api "chaosd/cli/cmd/load/api"
	partition_api "chaosd/cli/cmd/partition/api"
	ps_api "chaosd/cli/cmd/ps/api"
	restart_api "chaosd/cli/cmd/restart/api"

	"github.com/spf13/cobra"
)

func Init(rootCmd *cobra.Command) {
	dockerProvider := docker.NewDockerProvider()

	eventStore, err := adapters.NewDefaultEventStore()

	if err != nil {
		msg := fmt.Errorf("failed to create event store: %v", err)

		panic(msg)
	}

	sessionStore, err := adapters.NewDefaultSessionStore()

	if err != nil {
		msg := fmt.Errorf("failed to create session store: %v", err)

		panic(msg)
	}

	networkManager := network.NewManager(
		network.NewLinuxFirewallInjector(),
		network.NewNetemInjector(),
	)

	app := application.NewApplication(
		application.WithSessionStore(sessionStore),
		application.WithEventStore(eventStore),
		application.WithDockerProvider(dockerProvider),
		application.WithNetworkManager(networkManager),
	)

	rootCmd.AddCommand(events_api.NewEventsCmd(app))
	rootCmd.AddCommand(load_api.NewLoadCmd(app))
	rootCmd.AddCommand(heal_api.NewHealCmd(app))
	rootCmd.AddCommand(partition_api.NewPartitionCmd(app))
	rootCmd.AddCommand(ps_api.NewPsCmd(app))
	rootCmd.AddCommand(restart_api.NewRestartCmd(app))
}

func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
