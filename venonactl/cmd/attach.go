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

var attachRuntimeCmdOptions struct {
	runtimeEnvironmentName string
	kube                   struct {
		namespace      string
		inCluster      bool
		context        string
		kubePath       string
		serviceAccount string
	}
	kubeVenona struct {
		namespace string
		kubePath  string
		context   string
	}
	restartAgent bool
}

var attachRuntimeCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach Codefresh runtime to agent",
	Run: func(cmd *cobra.Command, args []string) {

		s := store.GetStore()
		lgr := createLogger("Attach-runtime", verbose, logFormatter)
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)

		s.CodefreshAPI = &store.CodefreshAPI{}
		s.AgentAPI = &store.AgentAPI{}

		if attachRuntimeCmdOptions.kubeVenona.kubePath == "" {
			attachRuntimeCmdOptions.kubeVenona.kubePath = kubeConfigPath
		}
		if attachRuntimeCmdOptions.kubeVenona.namespace == "" {
			attachRuntimeCmdOptions.kubeVenona.namespace = attachRuntimeCmdOptions.kube.namespace
		}
		if attachRuntimeCmdOptions.kubeVenona.context == "" {
			attachRuntimeCmdOptions.kubeVenona.context = attachRuntimeCmdOptions.kube.context
		}

		attachRuntimeCmdOptions.kube.serviceAccount = "venona"

		if attachRuntimeCmdOptions.kube.kubePath == "" {
			attachRuntimeCmdOptions.kube.kubePath = kubeConfigPath
		}

		fillKubernetesAPI(lgr, attachRuntimeCmdOptions.kubeVenona.context, attachRuntimeCmdOptions.kubeVenona.namespace, false)

		builder := plugins.NewBuilder(lgr)

		builderInstallOpt := &plugins.InstallOptions{
			ClusterNamespace:      attachRuntimeCmdOptions.kubeVenona.namespace,
			RuntimeEnvironment:    attachRuntimeCmdOptions.runtimeEnvironmentName,
			RuntimeClusterName:    attachRuntimeCmdOptions.kube.namespace,
			RuntimeServiceAccount: attachRuntimeCmdOptions.kube.serviceAccount,
			RestartAgent:          attachRuntimeCmdOptions.restartAgent,
		}

		// runtime
		builderInstallOpt.KubeBuilder = getKubeClientBuilder(attachRuntimeCmdOptions.kube.context, attachRuntimeCmdOptions.kube.namespace, attachRuntimeCmdOptions.kube.kubePath, false)

		// agent
		builderInstallOpt.AgentKubeBuilder = getKubeClientBuilder(attachRuntimeCmdOptions.kubeVenona.context,
			attachRuntimeCmdOptions.kubeVenona.namespace,
			attachRuntimeCmdOptions.kubeVenona.kubePath,
			false)

		builder.Add(plugins.RuntimeAttachType)

		var err error
		values := s.BuildValues()
		spn := createSpinner("Attaching runtime to agent (might take a few seconds)", "")
		spn.Start()
		for _, p := range builder.Get() {
			values, err = p.Install(builderInstallOpt, values)
			if err != nil {
				dieOnError(err)
			}
		}
		spn.Stop()
		lgr.Info("Attach to runtime completed Successfully")

	},
}

func init() {
	rootCmd.AddCommand(attachRuntimeCmd)
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")

	attachRuntimeCmd.Flags().StringVar(&attachRuntimeCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	attachRuntimeCmd.Flags().StringVar(&attachRuntimeCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	attachRuntimeCmd.Flags().StringVar(&attachRuntimeCmdOptions.kube.kubePath, "kube-config-path", viper.GetString("kubeconfig"), "Path to kubeconfig file (default is $HOME/.kube/config) [$KUBECONFIG]")

	attachRuntimeCmd.Flags().StringVar(&attachRuntimeCmdOptions.runtimeEnvironmentName, "runtime-name", viper.GetString("runtime-name"), "Name of the runtime as in codefresh")

	attachRuntimeCmd.Flags().StringVar(&attachRuntimeCmdOptions.kubeVenona.namespace, "kube-namespace-agent", viper.GetString("kube-namespace-agent"), "Name of the namespace where venona is installed [$KUBE_NAMESPACE]")
	attachRuntimeCmd.Flags().StringVar(&attachRuntimeCmdOptions.kubeVenona.context, "kube-context-name-agent", viper.GetString("kube-context-agent"), "Name of the kubernetes context on which venona is installed (default is current-context) [$KUBE_CONTEXT]")
	attachRuntimeCmd.Flags().StringVar(&attachRuntimeCmdOptions.kubeVenona.kubePath, "kube-config-path-agent", viper.GetString("kubeconfig-agent"), "Path to kubeconfig file (default is $HOME/.kube/config) for agent [$KUBECONFIG]")
	attachRuntimeCmd.Flags().BoolVar(&attachRuntimeCmdOptions.restartAgent, "restart-agent", viper.GetBool("restart-agent"), "Restart agent after attach operation")

}
