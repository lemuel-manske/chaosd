package main

import (
	"chaosd/cli/cmd"

	root_api "chaosd/cli/cmd/root/api"
)

func main() {
	rootCmd := root_api.NewRootCmd()

	cmd.Init(rootCmd)
	cmd.Execute(rootCmd)
}
