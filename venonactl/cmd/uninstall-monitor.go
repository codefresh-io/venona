package cmd

import (
	"github.com/codefresh-io/venona/venonactl/pkg/store"

	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uninstallMonitorAgentCmdOptions struct {
	kube struct {
		context   string
		namespace string
		kubePath  string
	}
}

var uninstallMonitorAgentCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Uninstall Codefresh's monitor",
	Run: func(cmd *cobra.Command, args []string) {
		s := store.GetStore()
		lgr := createLogger("UninstallMonitor", verbose)
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		fillKubernetesAPI(lgr, uninstallMonitorAgentCmdOptions.kube.context, uninstallMonitorAgentCmdOptions.kube.namespace, false)

		builder := plugins.NewBuilder(lgr)

		if uninstallMonitorAgentCmdOptions.kube.kubePath == "" {
			uninstallMonitorAgentCmdOptions.kube.kubePath = kubeConfigPath
		}

		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		// This is temporarily and used for signing
		s.CodefreshAPI = &store.CodefreshAPI{
			Host: cfAPIHost,
		}

		deleteOptions := &plugins.DeleteOptions{}
		// runtime
		deleteOptions.KubeBuilder = getKubeClientBuilder(uninstallMonitorAgentCmdOptions.kube.context,
			s.KubernetesAPI.Namespace,
			uninstallMonitorAgentCmdOptions.kube.kubePath,
			false)

		builder.Add(plugins.MonitorAgentPluginType)
		deleteOptions.ClusterNamespace = s.KubernetesAPI.Namespace
		for _, p := range builder.Get() {
			err := p.Delete(deleteOptions, s.BuildMinimizedValues())
			if err != nil {
				dieOnError(err)
			}
		}

		lgr.Info("Deletion of monitor is completed")
	},
}

func init() {
	uninstallCommand.AddCommand(uninstallMonitorAgentCmd)
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallMonitorAgentCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallMonitorAgentCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallMonitorAgentCmdOptions.kube.kubePath, "kube-config-path", viper.GetString("kubeconfig"), "Path to kubeconfig file (default is $HOME/.kube/config) [$KUBECONFIG]")

}
