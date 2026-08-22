package cli

import (
	"fmt"
	"os"

	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
)

type ComposeFile struct {
	Name     string         `yaml:"name"`
	Version  string         `yaml:"version"`
	Services map[string]any `yaml:"services"`
}

func NewLoadCmd(docker docker.DockerProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "load",
		Short: "Load a compose file",
		Long:  `Load a Docker compose file into chaosd daemon.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLoadCmd(cmd, args, docker)
		},
		Args: cobra.ExactArgs(1),
	}
}

var RootCmd = &cobra.Command{
	Use:   "chaosd",
	Short: "Chaosd is a laboratory for distributed systems",
	Run:   runChaosd,
}

func Init() {
	dockerProvider := docker.NewDockerProvider()

	RootCmd.AddCommand(NewLoadCmd(dockerProvider))
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
