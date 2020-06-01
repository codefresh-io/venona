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

	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var allTestPluginTypes = []string{
	plugins.RuntimeEnvironmentPluginType,
	plugins.VenonaPluginType,
	plugins.MonitorAgentPluginType,
	plugins.VolumeProvisionerPluginType,
	plugins.EnginePluginType,
	plugins.RuntimeAttachType,
}

var testCommandOptions struct {
	kube struct {
		namespace string
		context   string
	}
	plugin []string
}

var testCommand = &cobra.Command{
	Use:   "test",
	Short: "Run test on the target cluster prior installation",
	Run: func(cmd *cobra.Command, args []string) {
		lgr := createLogger("test", verbose)
		s := store.GetStore()
		extendStoreWithKubeClient(lgr)

		builder := plugins.NewBuilder(lgr)
		for _, p := range testCommandOptions.plugin {
			if p == plugins.RuntimeEnvironmentPluginType {
				builder.Add(plugins.RuntimeEnvironmentPluginType)
			}
			if p == plugins.VenonaPluginType {
				builder.Add(plugins.VenonaPluginType)
			}
			if p == plugins.MonitorAgentPluginType {
				builder.Add(plugins.MonitorAgentPluginType)
			}
			if p == plugins.VolumeProvisionerPluginType {
				builder.Add(plugins.VolumeProvisionerPluginType)
			}
			if p == plugins.EnginePluginType {
				builder.Add(plugins.EnginePluginType)
			}
			if p == plugins.RuntimeAttachType {
				builder.Add(plugins.RuntimeAttachType)
			}
		}
		var finalerr error
		for _, p := range builder.Get() {
			lgr.Info("Testing requirements", "installer", p.Name())
			err := p.Test(plugins.TestOptions{
				KubeBuilder:      getKubeClientBuilder(s.KubernetesAPI.ContextName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, false),
				ClusterNamespace: s.KubernetesAPI.Namespace,
			})
			if err != nil {
				if finalerr != nil {
					finalerr = fmt.Errorf("%s - %s", finalerr.Error(), err.Error())
				} else {
					finalerr = fmt.Errorf("%s", err.Error())

				}
			}
		}
		dieOnError(finalerr)

		lgr.Info("Cluster passed acceptance test")
	},
}

func init() {
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")

	testCommand.Flags().StringVar(&testCommandOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which monitor should be installed [$KUBE_NAMESPACE]")
	testCommand.Flags().StringVar(&testCommandOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which monitor should be installed (default is current-context) [$KUBE_CONTEXT]")
	testCommand.Flags().StringArrayVar(&testCommandOptions.plugin, "installer", allTestPluginTypes, "Which test to run, based on the installer type")

	rootCmd.AddCommand(testCommand)
}
