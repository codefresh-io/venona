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

		lgr.Info("Deletion of app-proxy is completed")
	},
}

func init() {
	uninstallCommand.AddCommand(uninstallAppProxytCmd)
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	uninstallAppProxytCmd.Flags().StringVar(&uninstallAppProxyCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace from which the app-proxy should be uninstalled [$KUBE_NAMESPACE]")
	uninstallAppProxytCmd.Flags().StringVar(&uninstallAppProxyCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context from which the app-proxy should be uninstalled (default is current-context) [$KUBE_CONTEXT]")
	uninstallAppProxytCmd.Flags().StringArrayVar(&uninstallAppProxyCmdOptions.templateValues, "set-value", []string{}, "Set values for templates --set-value agentId=12345")
	uninstallAppProxytCmd.Flags().StringArrayVar(&uninstallAppProxyCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	uninstallAppProxytCmd.Flags().StringArrayVarP(&uninstallAppProxyCmdOptions.templateValueFiles, "values", "f", []string{}, "Specify values in a YAML file")

}
