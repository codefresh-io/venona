package cmd

import (
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

var cnf *viper.Viper = viper.New()

var rootCmdOptions struct {
}

var rootCmd = &cobra.Command{
	Use:     "venona",
	Version: "1.2.0", // TODO: read from ld flags
	Long:    "Codefresh agent process",
}

// Execute - execute the root command
func Execute() {
	err := rootCmd.Execute()
	dieOnError(err)
}
