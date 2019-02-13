package cmd

/*
Copyright 2019 The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"github.com/spf13/viper"

	"github.com/codefresh-io/venona/venonactl/internal"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "venona",
	Short: "A command line application for Codefresh",
}

// Execute - execute the root command
func Execute() {
	err := rootCmd.Execute()
	internal.DieOnError(err)
}

func init() {
	viper.AutomaticEnv()
	viper.BindEnv("kubeconfig", "KUBECONFIG")
	viper.BindEnv("cfconfig", "CFCONFIG")

	viper.BindEnv("apihost", "API_HOST")
	viper.BindEnv("apitoken", "API_TOKEN")

	rootCmd.PersistentFlags().StringVar(&configPath, "cfconfig", viper.GetString("cfconfig"), "Config file (default is $HOME/.cfconfig) [$CFCONFIG]")
	rootCmd.PersistentFlags().StringVar(&cfAPIHost, "api-host", viper.GetString("apihost"), "Host of codefresh [$API_HOST]")
	rootCmd.PersistentFlags().StringVar(&cfAPIToken, "api-token", viper.GetString("apitoken"), "Codefresh API token [$API_TOKEN]")
	rootCmd.PersistentFlags().StringVar(&cfContext, "context", "", "Name of the context from --cfconfig (default is current-context)")

	rootCmd.PersistentFlags().StringVar(&kubeConfigPath, "kube-config-path", viper.GetString("kubeconfig"), "Path to kubeconfig file (default is $HOME/.kube/config) [$KUBECONFIG]")

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Print logs")
	rootCmd.PersistentFlags().BoolVar(&skipVerionCheck, "skip-version-check", false, "Do not compare current Venona's version with latest")

}
