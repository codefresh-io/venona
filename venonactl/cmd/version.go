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
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Venona version",
	Run: func(cmd *cobra.Command, args []string) {
		s := store.GetStore()
		lgr := createLogger("Version", verbose)
		buildBasicStore(lgr)
		fmt.Printf("Date: %s\n", s.Version.Current.Date)
		fmt.Printf("Commit: %s\n", s.Version.Current.Commit)
		fmt.Printf("Local Version: %s\n", s.Version.Current.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	
	versionCmd.Flags().StringVar(&installAgentCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona is installed [$KUBE_NAMESPACE]")
	versionCmd.Flags().StringVar(&installAgentCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona is installed (default is current-context) [$KUBE_CONTEXT]")
	versionCmd.Flags().StringVar(&installAgentCmdOptions.kube.context, "server-version-only", viper.GetString("kube-context"), "Name of the kubernetes context on which venona is installed (default is current-context) [$KUBE_CONTEXT]")

}
