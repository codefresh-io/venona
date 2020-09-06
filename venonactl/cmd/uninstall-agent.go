package cmd


import (
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/store"

	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uninstallAgentCmdOptions struct {
	kube struct {
		context   string
		namespace string
		kubePath  string
	}
	templateValues       []string
	templateFileValues   []string
	templateValueFiles   []string		
}

var uninstallAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Uninstall Codefresh's agent",
	Run: func(cmd *cobra.Command, args []string) {

		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file> 
		templateValuesMap, err := templateValuesToMap(
			uninstallAgentCmdOptions.templateValueFiles, 
			uninstallAgentCmdOptions.templateValues, 
			uninstallAgentCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Token", &cfAPIToken)		
		mergeValueStr(templateValuesMap, "Namespace", &uninstallAgentCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &uninstallAgentCmdOptions.kube.context)

		s := store.GetStore()
		lgr := createLogger("UninstallAgent", verbose, logFormatter)
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		
		s.CodefreshAPI = &store.CodefreshAPI{}
		s.AgentAPI = &store.AgentAPI{}

		
		builder := plugins.NewBuilder(lgr)
		if uninstallAgentCmdOptions.kube.context == "" {
			dieOnError(fmt.Errorf("Context name is required in order to uninstall agent"))
		}
		if uninstallAgentCmdOptions.kube.namespace == "" {
			dieOnError(fmt.Errorf("Namespace name is required to in order to uninstall agent"))
		}


			deleteOptions := &plugins.DeleteOptions{}
			s.KubernetesAPI.ContextName = uninstallAgentCmdOptions.kube.context
			s.KubernetesAPI.Namespace = uninstallAgentCmdOptions.kube.namespace

			builder.Add(plugins.VenonaPluginType)
			deleteOptions.KubeBuilder = getKubeClientBuilder(s.KubernetesAPI.ContextName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, false)
			deleteOptions.ClusterNamespace = uninstallAgentCmdOptions.kube.namespace

			values := s.BuildValues()
			values = mergeMaps(values, templateValuesMap)
			for _, p := range builder.Get() {
					err := p.Delete(deleteOptions, values)
					if err != nil {
						dieOnError(err)
					}
				}

			lgr.Info("Deletion of agent is completed")
	},
}

func init() {
	uninstallCommand.AddCommand(uninstallAgentCmd)
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	uninstallAgentCmd.Flags().StringVar(&uninstallAgentCmdOptions.kube.context, "kube-context-name", "", "Name of the kubernetes context on which venona should be uninstalled (default is current-context) [$KUBE_CONTEXT]")
	uninstallAgentCmd.Flags().StringVar(&uninstallAgentCmdOptions.kube.namespace, "kube-namespace", "", "Name of the namespace on which venona should be uninstalled [$KUBE_NAMESPACE]")
	uninstallAgentCmd.Flags().StringArrayVar(&uninstallAgentCmdOptions.templateValues, "set-value", []string{}, "Set values for templates --set-value agentId=12345")
	uninstallAgentCmd.Flags().StringArrayVar(&uninstallAgentCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	uninstallAgentCmd.Flags().StringArrayVarP(&uninstallAgentCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

}