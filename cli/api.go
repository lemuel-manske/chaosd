package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type ComposeFile struct {
	Version  string         `yaml:"version"`
	Services map[string]any `yaml:"services"`
}

var LoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load a compose file",
	Long:  `Load a Docker compose file into chaosd daemon.`,
	RunE:  runLoadCmd,
	Args:  cobra.ExactArgs(1),
}

var RootCmd = &cobra.Command{
	Use:   "chaosd",
	Short: "Chaosd is a laboratory for distributed systems",
	Run:   runChaosd,
}

func Init() {
	RootCmd.AddCommand(LoadCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
