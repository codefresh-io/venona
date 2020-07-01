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
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var upgradeCmdOpt struct {
	kube struct {
		context   string
		namespace string
	}
}

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade existing 1.X runner",
	Run: func(cmd *cobra.Command, args []string) {
		lgr := createLogger("Upgrade", true, logFormatter)
		builder := plugins.NewBuilder(lgr)
		builder.Add(plugins.VenonaPluginType)

		s := store.GetStore()
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		extendStoreWithCodefershClient(lgr)
		extendStoreWithAgentAPI(lgr, "", "")
		fillKubernetesAPI(lgr, upgradeCmdOpt.kube.context, upgradeCmdOpt.kube.namespace, false)
		values := s.BuildValues()
		spn := createSpinner("Upgarding runtime (might take a few seconds)", "")
		spn.Start()
		defer spn.Stop()
		var err error
		for _, p := range builder.Get() {
			values, err = p.Upgrade(&plugins.UpgradeOptions{
				ClusterNamespace: upgradeCmdOpt.kube.namespace,
				ClusterName:      upgradeCmdOpt.kube.namespace,
				KubeBuilder:      getKubeClientBuilder(upgradeCmdOpt.kube.context, upgradeCmdOpt.kube.namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster),
			}, values)
			if err != nil {
				dieOnError(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringVar(&upgradeCmdOpt.kube.context, "kube-context-name", "", "Set name to overwrite the context name saved in Codefresh")
	upgradeCmd.Flags().StringVar(&upgradeCmdOpt.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona is installed [$KUBE_NAMESPACE]")
}
