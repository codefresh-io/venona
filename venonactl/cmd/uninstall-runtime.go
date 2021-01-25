package cmd

import (
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/store"

	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uninstallRunimeCmdOptions struct {
	kube struct {
		context   string
		namespace string
		kubePath  string
	}
	kubeVenona struct {
		namespace string
		kubePath  string
		context   string
	}
	runtimeEnvironmentName string
	storageClassName       string
	restartAgent           bool
	templateValues         []string
	templateFileValues     []string
	templateValueFiles     []string
}

var uninstallRuntimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Uninstall Codefresh's runtime",
	Run: func(cmd *cobra.Command, args []string) {
		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file>
		templateValuesMap, err := templateValuesToMap(
			uninstallRunimeCmdOptions.templateValueFiles,
			uninstallRunimeCmdOptions.templateValues,
			uninstallRunimeCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "ConfigPath", &uninstallRunimeCmdOptions.kube.kubePath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Token", &cfAPIToken)

		mergeValueStr(templateValuesMap, "Namespace", &uninstallRunimeCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &uninstallRunimeCmdOptions.kube.context)
		mergeValueStr(templateValuesMap, "RuntimeEnvironmentName", &uninstallRunimeCmdOptions.runtimeEnvironmentName)
		mergeValueStr(templateValuesMap, "StorageClass", &uninstallRunimeCmdOptions.storageClassName)

		s := store.GetStore()
		lgr := createLogger("UninstallRuntime", verbose, logFormatter)
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)

		s.CodefreshAPI = &store.CodefreshAPI{}
		s.AgentAPI = &store.AgentAPI{}

		builder := plugins.NewBuilder(lgr)
		if uninstallRunimeCmdOptions.kube.context == "" {
			dieOnError(fmt.Errorf("Context name is required in order to uninstall runtime"))
		}
		if uninstallRunimeCmdOptions.kube.namespace == "" {
			dieOnError(fmt.Errorf("Namespace name is required to in order to uninstall runtime"))
		}
		s.KubernetesAPI.ContextName = uninstallRunimeCmdOptions.kube.context
		s.KubernetesAPI.Namespace = uninstallRunimeCmdOptions.kube.namespace

		if uninstallRunimeCmdOptions.kubeVenona.kubePath == "" {
			uninstallRunimeCmdOptions.kubeVenona.kubePath = kubeConfigPath
		}
		if uninstallRunimeCmdOptions.kubeVenona.namespace == "" {
			uninstallRunimeCmdOptions.kubeVenona.namespace = uninstallRunimeCmdOptions.kube.namespace
		}
		if uninstallRunimeCmdOptions.kubeVenona.context == "" {
			uninstallRunimeCmdOptions.kubeVenona.context = uninstallRunimeCmdOptions.kube.context
		}

		if uninstallRunimeCmdOptions.kube.kubePath == "" {
			uninstallRunimeCmdOptions.kube.kubePath = kubeConfigPath
		}

		deleteOptions := &plugins.DeleteOptions{}
		// runtime
		deleteOptions.KubeBuilder = getKubeClientBuilder(uninstallRunimeCmdOptions.kube.context,
			uninstallRunimeCmdOptions.kube.namespace,
			uninstallRunimeCmdOptions.kube.kubePath,
			false,
			false)

		// agent
		deleteOptions.AgentKubeBuilder = getKubeClientBuilder(uninstallRunimeCmdOptions.kubeVenona.context,
			uninstallRunimeCmdOptions.kubeVenona.namespace,
			uninstallRunimeCmdOptions.kubeVenona.kubePath,
			false,
			false)

		builder.Add(plugins.RuntimeEnvironmentPluginType)
		if isUsingDefaultStorageClass(uninstallRunimeCmdOptions.storageClassName) {
			builder.Add(plugins.VolumeProvisionerPluginType)
		}
		builder.Add(plugins.RuntimeAttachType)
		builder.Add(plugins.EnginePluginType)
		deleteOptions.ClusterNamespace = uninstallRunimeCmdOptions.kube.namespace
		deleteOptions.AgentNamespace = uninstallRunimeCmdOptions.kubeVenona.namespace
		deleteOptions.RuntimeEnvironment = uninstallRunimeCmdOptions.runtimeEnvironmentName
		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)
		for _, p := range builder.Get() {
			err := p.Delete(deleteOptions, values)
			if err != nil {
				dieOnError(err)
			}
		}

		lgr.Info("Deletion of runtime is completed")
	},
}

func init() {
	uninstallCommand.AddCommand(uninstallRuntimeCmd)
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.kube.kubePath, "kube-config-path", viper.GetString("kubeconfig"), "Path to kubeconfig file (default is $HOME/.kube/config) [$KUBECONFIG]")

	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.runtimeEnvironmentName, "runtime-name", viper.GetString("runtime-name"), "Name of the runtime as in codefresh")

	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.kubeVenona.namespace, "kube-namespace-agent", viper.GetString("kube-namespace-agent"), "Name of the namespace where venona is installed [$KUBE_NAMESPACE]")
	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.kubeVenona.context, "kube-context-name-agent", viper.GetString("kube-context-agent"), "Name of the kubernetes context on which venona is installed (default is current-context) [$KUBE_CONTEXT]")
	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.kubeVenona.kubePath, "kube-config-path-agent", viper.GetString("kubeconfig-agent"), "Path to kubeconfig file (default is $HOME/.kube/config) for agent [$KUBECONFIG]")
	uninstallRuntimeCmd.Flags().BoolVar(&uninstallRunimeCmdOptions.restartAgent, "restart-agent", viper.GetBool("restart-agent"), "Restart agent after attach operation")

	uninstallRuntimeCmd.Flags().StringVar(&uninstallRunimeCmdOptions.storageClassName, "storage-class-name", viper.GetString("storage-class-name"), "Storage class name of the runtime to be uninstalled")

	uninstallRuntimeCmd.Flags().StringArrayVar(&uninstallRunimeCmdOptions.templateValues, "set-value", []string{}, "Set values for templates --set-value agentId=12345")
	uninstallRuntimeCmd.Flags().StringArrayVar(&uninstallRunimeCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	uninstallRuntimeCmd.Flags().StringArrayVarP(&uninstallRunimeCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

}
