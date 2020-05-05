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

var installK8sAgentCmdOptions struct {
	kube struct {
		namespace    string
		inCluster    bool
		context      string
		nodeSelector string
	}
}

// installK8sAgentCmd represents the install command
var installK8sAgentCmd = &cobra.Command{
	Use:   "k8sagent",
	Short: "Install Codefresh's k8s agent on cluster",
	Run: func(cmd *cobra.Command, args []string) {

		s := store.GetStore()

		lgr := createLogger("Install-k8s-agent", verbose)
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		fillKubernetesAPI(lgr, installK8sAgentCmdOptions.kube.context, installK8sAgentCmdOptions.kube.namespace, installK8sAgentCmdOptions.kube.inCluster)

		builder := plugins.NewBuilder(lgr)
		builder.Add(plugins.K8sAgentPluginType)

		builderInstallOpt := &plugins.InstallOptions{
			ClusterNamespace: s.KubernetesAPI.Namespace,
		}

		builderInstallOpt.KubeBuilder = getKubeClientBuilder(s.KubernetesAPI.ContextName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster)

		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		// This is temporarily and used for signing
		s.CodefreshAPI = &store.CodefreshAPI{
			Host: cfAPIHost,
		}

		values := s.BuildMinimizedValues()

		for _, p := range builder.Get() {
			_, err := p.Install(builderInstallOpt, values)
			if err != nil {
				dieOnError(err)
			}
		}
		lgr.Info("Agent installation completed Successfully")
	},
}

func init() {
	installCommand.AddCommand(installK8sAgentCmd)

	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	installK8sAgentCmd.Flags().StringVar(&installK8sAgentCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	installK8sAgentCmd.Flags().StringVar(&installK8sAgentCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")

	installK8sAgentCmd.Flags().BoolVar(&installK8sAgentCmdOptions.kube.inCluster, "in-cluster", false, "Set flag if venona is been installed from inside a cluster")

}
