package cmd

import (
	"github.com/spf13/cobra"
)

var version string

var rootCmd = &cobra.Command{
	Use:     "venona",
	Version: version,
	Long:    "Codefresh agent process",
}

// Execute - execute the root command
func Execute() {
	err := rootCmd.Execute()
	dieOnError(err)
}
