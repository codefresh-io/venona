package cmd

/*
Copyright 2020 The Codefresh Authors.

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

var installMonitorAgentCmdOptions struct {
	kube struct {
		namespace    string
		context      string
		nodeSelector string
	}
	clusterId          string
	helm3              bool
	codefreshToken     string
	codefreshHost      string
	dockerRegistry     string
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
	resources          map[string]interface{}
	dryRun             bool
}

// installK8sAgentCmd represents the install command
var installMonitorAgentCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Install Codefresh's monitor agent on cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file>
		templateValuesMap, err := templateValuesToMap(
			installMonitorAgentCmdOptions.templateValueFiles,
			installMonitorAgentCmdOptions.templateValues,
			installMonitorAgentCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "CodefreshHost", &installMonitorAgentCmdOptions.codefreshHost)
		mergeValueStr(templateValuesMap, "Token", &cfAPIToken)
		mergeValueStr(templateValuesMap, "Token", &installMonitorAgentCmdOptions.codefreshToken)
		mergeValueStr(templateValuesMap, "Namespace", &installMonitorAgentCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &installMonitorAgentCmdOptions.kube.context)

		mergeValueMSI(templateValuesMap, "Monitor.resources", &installMonitorAgentCmdOptions.resources)

		mergeValueStr(templateValuesMap, "DockerRegistry", &installMonitorAgentCmdOptions.dockerRegistry)
		mergeValueStr(templateValuesMap, "ClusterId", &installMonitorAgentCmdOptions.clusterId)
		mergeValueBool(templateValuesMap, "helm3", &installMonitorAgentCmdOptions.helm3)

		s := store.GetStore()

		lgr := createLogger("Install-monitor-agent", verbose, logFormatter)
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		fillKubernetesAPI(lgr, installMonitorAgentCmdOptions.kube.context, installMonitorAgentCmdOptions.kube.namespace, false)
		s.Monitor.Resources = installMonitorAgentCmdOptions.resources

		builder := plugins.NewBuilder(lgr)
		builder.Add(plugins.MonitorAgentPluginType)

		builderInstallOpt := &plugins.InstallOptions{
			ClusterNamespace: s.KubernetesAPI.Namespace,
			DryRun:           installMonitorAgentCmdOptions.dryRun,
		}

		builderInstallOpt.KubeBuilder = getKubeClientBuilder(s.KubernetesAPI.ContextName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster, installMonitorAgentCmdOptions.dryRun)

		if installMonitorAgentCmdOptions.clusterId == "" {
			dieOnError(fmt.Errorf("Cluster id is required in order to install monitor"))
		}

		s.ClusterId = installMonitorAgentCmdOptions.clusterId
		s.Helm3 = installMonitorAgentCmdOptions.helm3
		s.DockerRegistry = installMonitorAgentCmdOptions.dockerRegistry

		if installMonitorAgentCmdOptions.codefreshHost == "" {
			installMonitorAgentCmdOptions.codefreshHost = "https://g.codefresh.io"
		}

		if installMonitorAgentCmdOptions.codefreshToken == "" {
			dieOnError(fmt.Errorf("Codefresh token is required in order to install monitor"))
		}

		s.CodefreshAPI = &store.CodefreshAPI{
			Host:  installMonitorAgentCmdOptions.codefreshHost,
			Token: installMonitorAgentCmdOptions.codefreshToken,
		}

		// stub  , not need actually for monitor
		s.AgentAPI = &store.AgentAPI{
			Token: "",
			Id:    "",
		}

		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)

		for _, p := range builder.Get() {
			_, err := p.Install(cmd.Context(), builderInstallOpt, values)
			dieOnError(err)
		}
		lgr.Info("Monitor agent installation completed Successfully")
	},
}

func init() {
	installCommand.AddCommand(installMonitorAgentCmd)

	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	installMonitorAgentCmd.Flags().StringVar(&installMonitorAgentCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which monitor should be installed [$KUBE_NAMESPACE]")
	installMonitorAgentCmd.Flags().StringVar(&installMonitorAgentCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which monitor should be installed (default is current-context) [$KUBE_CONTEXT]")
	installMonitorAgentCmd.Flags().StringVar(&installMonitorAgentCmdOptions.clusterId, "clusterId", "", "Cluster Id")
	installMonitorAgentCmd.Flags().StringVar(&installMonitorAgentCmdOptions.codefreshToken, "codefreshToken", "", "Codefresh token")
	installMonitorAgentCmd.Flags().StringVar(&installMonitorAgentCmdOptions.dockerRegistry, "docker-registry", "", "The prefix for the container registry that will be used for pulling the required components images. Example: --docker-registry=\"docker.io\"")

	installMonitorAgentCmd.Flags().StringVar(&installMonitorAgentCmdOptions.codefreshHost, "codefreshHost", "", "Override codefresh host if you use your own codefresh installation")
	installMonitorAgentCmd.Flags().BoolVar(&installMonitorAgentCmdOptions.dryRun, "dry-run", false, "Set to true to simulate installation")
	installMonitorAgentCmd.Flags().BoolVar(&installMonitorAgentCmdOptions.helm3, "helm3", false, "Set flag if cluster use helm3")
	installMonitorAgentCmd.Flags().StringArrayVar(&installMonitorAgentCmdOptions.templateValues, "set-value", []string{}, "Set values for templates, example: --set-value LocalVolumesDir=/mnt/disks/ssd0/codefresh-volumes")
	installMonitorAgentCmd.Flags().StringArrayVar(&installMonitorAgentCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file, example: --set-file Storage.GoogleServiceAccount=/path/to/service-account.json")
	installMonitorAgentCmd.Flags().StringArrayVarP(&installMonitorAgentCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

}
