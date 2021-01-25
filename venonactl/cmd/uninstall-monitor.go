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
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
}

var uninstallMonitorAgentCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Uninstall Codefresh's monitor",
	Run: func(cmd *cobra.Command, args []string) {

		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file>
		templateValuesMap, err := templateValuesToMap(
			uninstallMonitorAgentCmdOptions.templateValueFiles,
			uninstallMonitorAgentCmdOptions.templateValues,
			uninstallMonitorAgentCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "ConfigPath", &uninstallMonitorAgentCmdOptions.kube.kubePath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Token", &cfAPIToken)
		mergeValueStr(templateValuesMap, "Namespace", &uninstallMonitorAgentCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &uninstallMonitorAgentCmdOptions.kube.context)

		s := store.GetStore()
		lgr := createLogger("UninstallMonitor", verbose, logFormatter)
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

		// stub  , not need actually for monitor
		s.AgentAPI = &store.AgentAPI{
			Token: "",
			Id:    "",
		}

		deleteOptions := &plugins.DeleteOptions{}
		// runtime
		deleteOptions.KubeBuilder = getKubeClientBuilder(uninstallMonitorAgentCmdOptions.kube.context,
			s.KubernetesAPI.Namespace,
			uninstallMonitorAgentCmdOptions.kube.kubePath,
			false,
			false)

		builder.Add(plugins.MonitorAgentPluginType)
		deleteOptions.ClusterNamespace = s.KubernetesAPI.Namespace
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
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallMonitorAgentCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which monitor should be uninstalled [$KUBE_NAMESPACE]")
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallMonitorAgentCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which monitor should be uninstalled (default is current-context) [$KUBE_CONTEXT]")
	uninstallMonitorAgentCmd.Flags().StringVar(&uninstallMonitorAgentCmdOptions.kube.kubePath, "kube-config-path", viper.GetString("kubeconfig"), "Path to kubeconfig file (default is $HOME/.kube/config) [$KUBECONFIG]")
	uninstallMonitorAgentCmd.Flags().StringArrayVar(&uninstallMonitorAgentCmdOptions.templateValues, "set-value", []string{}, "Set values for templates --set-value agentId=12345")
	uninstallMonitorAgentCmd.Flags().StringArrayVar(&uninstallMonitorAgentCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	uninstallMonitorAgentCmd.Flags().StringArrayVarP(&uninstallMonitorAgentCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

}
