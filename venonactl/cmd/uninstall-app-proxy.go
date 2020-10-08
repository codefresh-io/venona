package cmd

import (
	"github.com/codefresh-io/venona/venonactl/pkg/store"

	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uninstallAppProxyCmdOptions struct {
	kube struct {
		context   string
		namespace string
	}
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
}

var uninstallAppProxytCmd = &cobra.Command{
	Use:   "app-proxy",
	Short: "Uninstall Codefresh's App-Proxy component",
	Run: func(cmd *cobra.Command, args []string) {

		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file>
		templateValuesMap, err := templateValuesToMap(
			uninstallAppProxyCmdOptions.templateValueFiles,
			uninstallAppProxyCmdOptions.templateValues,
			uninstallAppProxyCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Namespace", &uninstallAppProxyCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &uninstallAppProxyCmdOptions.kube.context)

		s := store.GetStore()
		lgr := createLogger("UninstallAppProxy", verbose, logFormatter)
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		fillKubernetesAPI(lgr, uninstallAppProxyCmdOptions.kube.context, uninstallAppProxyCmdOptions.kube.namespace, false)

		builder := plugins.NewBuilder(lgr)

		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		s.CodefreshAPI = &store.CodefreshAPI{
			Host: cfAPIHost,
		}
		s.AgentAPI = &store.AgentAPI{
			Token: "",
			Id:    "",
		}

		deleteOptions := &plugins.DeleteOptions{}
		deleteOptions.KubeBuilder = getKubeClientBuilder(
			uninstallAppProxyCmdOptions.kube.context,
			uninstallAppProxyCmdOptions.kube.namespace,
			kubeConfigPath,
			false)

		builder.Add(plugins.AppProxyPluginType)
		deleteOptions.ClusterNamespace = uninstallAppProxyCmdOptions.kube.namespace
		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)
		for _, p := range builder.Get() {
			err := p.Delete(deleteOptions, values)
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
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallAppProxyCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which monitor should be uninstalled [$KUBE_NAMESPACE]")
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallAppProxyCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which monitor should be uninstalled (default is current-context) [$KUBE_CONTEXT]")
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallAppProxyCmdOptions.kube.kubePath, "kube-config-path", viper.GetString("kubeconfig"), "Path to kubeconfig file (default is $HOME/.kube/config) [$KUBECONFIG]")
	uninstallMonitorAgentCmd.Flags().StringArrayVar(&uninstallAppProxyCmdOptions.templateValues, "set-value", []string{}, "Set values for templates --set-value agentId=12345")
	uninstallMonitorAgentCmd.Flags().StringArrayVar(&uninstallAppProxyCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	uninstallMonitorAgentCmd.Flags().StringArrayVarP(&uninstallAppProxyCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

}
