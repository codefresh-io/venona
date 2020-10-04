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
	"fmt"
	"strings"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// cliValues "helm.sh/helm/v3/pkg/cli/values"
	// "helm.sh/helm/v3/pkg/getter"
)

var installAgentCmdOptions struct {
	dryRun bool
	kube   struct {
		namespace    string
		inCluster    bool
		context      string
		nodeSelector string
	}
	venona struct {
		version string
	}
	agentToken           string
	agentID              string
	kubernetesRunnerType bool
	tolerations          string
	envVars              []string
	dockerRegistry       string
	templateValues       []string
	templateFileValues   []string
	templateValueFiles   []string
	resources			 store.Resources
}

var installAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Install Codefresh's agent ",
	Run: func(cmd *cobra.Command, args []string) {

		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file>
		templateValuesMap, err := templateValuesToMap(
			installAgentCmdOptions.templateValueFiles,
			installAgentCmdOptions.templateValues,
			installAgentCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Token", &cfAPIToken)
		mergeValueStr(templateValuesMap, "Namespace", &installAgentCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &installAgentCmdOptions.kube.context)
		mergeValueStr(templateValuesMap, "NodeSelector", &installAgentCmdOptions.kube.nodeSelector)
		tolerations := getTolerations()
		mergeValueStr(templateValuesMap, "Tolerations", &tolerations)
		mergeValueStr(templateValuesMap, "DockerRegistry", &installAgentCmdOptions.dockerRegistry)

		mergeValueStr(templateValuesMap, "AgentToken", &installAgentCmdOptions.agentToken)
		mergeValueStr(templateValuesMap, "AgentId", &installAgentCmdOptions.agentID)
		mergeValueStr(templateValuesMap, "Image.Tag", &installAgentCmdOptions.venona.version)

		installAgentCmdOptions.resources.Limits = &store.MemoryCPU{}
		installAgentCmdOptions.resources.Requests = &store.MemoryCPU{}

		mergeValueStr(templateValuesMap, "Runner.Requests.CPU", &installAgentCmdOptions.resources.Requests.CPU)
		mergeValueStr(templateValuesMap, "Runner.Requests.Memory", &installAgentCmdOptions.resources.Requests.Memory)
		mergeValueStr(templateValuesMap, "Runner.Limits.CPU", &installAgentCmdOptions.resources.Limits.CPU)
		mergeValueStr(templateValuesMap, "Runner.Limits.Memory", &installAgentCmdOptions.resources.Limits.Memory)

		//mergeValueStrArray(&installAgentCmdOptions.envVars, "envVars", nil, "More env vars to be declared \"key=value\"")

		mergeValueBool(templateValuesMap, "InCluster", &installAgentCmdOptions.kube.inCluster)
		mergeValueBool(templateValuesMap, "kubernetesRunnerType", &installAgentCmdOptions.kubernetesRunnerType)

		s := store.GetStore()
		lgr := createLogger("Install-agent", verbose, logFormatter)
		buildBasicStore(lgr)

		extendStoreWithAgentAPI(lgr, installAgentCmdOptions.agentToken, installAgentCmdOptions.agentID)
		extendStoreWithKubeClient(lgr)
		fillCodefreshAPI(lgr)
		builder := plugins.NewBuilder(lgr)
		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		builderInstallOpt := &plugins.InstallOptions{
			CodefreshHost: cfAPIHost,
		}

		if installAgentCmdOptions.agentToken == "" {
			installAgentCmdOptions.agentToken = cfAPIToken
		}
		if installAgentCmdOptions.agentToken == "" {
			dieOnError(fmt.Errorf("Agent token is required in order to install agent"))
		}

		if installAgentCmdOptions.agentID == "" {
			dieOnError(fmt.Errorf("Agent id is required in order to install agent"))
		}

		fillKubernetesAPI(lgr, installAgentCmdOptions.kube.context, installAgentCmdOptions.kube.namespace, false)
		s.Runner = &store.Runner{}
		s.Runner.Resources = &store.Resources{}
		fillResouces(s.Runner.Resources, &installAgentCmdOptions.resources)

		s.KubernetesAPI.Tolerations = tolerations

		if installAgentCmdOptions.venona.version != "" {
			version := installAgentCmdOptions.venona.version
			lgr.Info("Version set manually", "version", version)
			s.Image.Tag = version
			s.Version.Current.Version = version
		}
		s.DockerRegistry = installAgentCmdOptions.dockerRegistry
		if installAgentCmdOptions.envVars != nil {
			s.AdditionalEnvVars = make(map[string]string)
			for _, part := range installAgentCmdOptions.envVars {
				splited := strings.Split(part, "=")
				s.AdditionalEnvVars[splited[0]] = splited[1]
			}
		}

		s.KubernetesAPI.NodeSelector = installAgentCmdOptions.kube.nodeSelector

		builderInstallOpt.ClusterName = s.KubernetesAPI.ContextName
		builderInstallOpt.KubeBuilder = getKubeClientBuilder(builderInstallOpt.ClusterName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster)
		builderInstallOpt.ClusterNamespace = s.KubernetesAPI.Namespace

		builder.Add(plugins.VenonaPluginType)

		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)

		for _, p := range builder.Get() {
			values, err = p.Install(builderInstallOpt, values)
			if err != nil {
				dieOnError(err)
			}
		}
		lgr.Info("Agent installation completed Successfully")
	},
}

func getTolerations() string {

	if installAgentCmdOptions.tolerations != "" {
		var tolerationsString string

		if installAgentCmdOptions.tolerations[0] == '@' {
			tolerationsString = loadTolerationsFromFile(installAgentCmdOptions.tolerations[1:])
		} else {
			tolerationsString = installAgentCmdOptions.tolerations
		}

		tolerations, err := parseTolerations(tolerationsString)
		if err != nil {
			dieOnError(err)
		}
		return tolerations

	}
	return ""
}


func init() {
	installCommand.AddCommand(installAgentCmd)

	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.agentToken, "agentToken", "", "Agent token created by codefresh")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.agentID, "agentId", "", "Agent id created by codefresh")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.venona.version, "venona-version", "", "Version of venona to install (default is the latest)")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.kube.nodeSelector, "kube-node-selector", "", "The kubernetes node selector \"key=value\" to be used by venona resources (default is no node selector)")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.tolerations, "tolerations", "", "The kubernetes tolerations as JSON string to be used by venona resources (default is no tolerations)")
	installAgentCmd.Flags().StringArrayVar(&installAgentCmdOptions.envVars, "envVars", nil, "More env vars to be declared \"key=value\"")
	installAgentCmd.Flags().StringVar(&installAgentCmdOptions.dockerRegistry, "docker-registry", "", "The prefix for the container registry that will be used for pulling the required components images. Example: --docker-registry=\"docker.io\"")

	installAgentCmd.Flags().BoolVar(&installAgentCmdOptions.kube.inCluster, "in-cluster", false, "Set flag if venona is been installed from inside a cluster")
	installAgentCmd.Flags().BoolVar(&installAgentCmdOptions.dryRun, "dry-run", false, "Set to true to simulate installation")
	installAgentCmd.Flags().BoolVar(&installAgentCmdOptions.kubernetesRunnerType, "kubernetes-runner-type", false, "Set the runner type to kubernetes (alpha feature)")

	installAgentCmd.Flags().StringArrayVar(&installAgentCmdOptions.templateValues, "set-value", []string{}, "Set values for templates --set-value agentId=12345")
	installAgentCmd.Flags().StringArrayVar(&installAgentCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	installAgentCmd.Flags().StringArrayVarP(&installAgentCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

}

func fillCodefreshAPI(logger logger.Logger) {
	s := store.GetStore()
	s.CodefreshAPI = &store.CodefreshAPI{
		Host: cfAPIHost,
	}

}
