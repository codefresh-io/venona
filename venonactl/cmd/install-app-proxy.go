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

	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installAppProxyCmdOptions struct {
	kube struct {
		namespace string
		context   string
	}
	dockerRegistry     string
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
	resources          map[string]interface{}
	host               string
	ingressClass       string
	dryRun             bool
}

var installAppProxyCmd = &cobra.Command{
	Use:   "app-proxy",
	Short: "Install App proxy ",
	Run: func(cmd *cobra.Command, args []string) {

		templateValuesMap, err := templateValuesToMap(
			installAppProxyCmdOptions.templateValueFiles,
			installAppProxyCmdOptions.templateValues,
			installAppProxyCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}

		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Namespace", &installAppProxyCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &installAppProxyCmdOptions.kube.context)
		mergeValueStr(templateValuesMap, "AppProxy.Ingress.Host", &installAppProxyCmdOptions.host)
		mergeValueStr(templateValuesMap, "AppProxy.Ingress.IngressClass", &installAppProxyCmdOptions.ingressClass)
		mergeValueStr(templateValuesMap, "DockerRegistry", &installAppProxyCmdOptions.dockerRegistry)
		mergeValueMSI(templateValuesMap, "AppProxy.resources", &installAppProxyCmdOptions.resources)

		s := store.GetStore()
		lgr := createLogger("Install-agent", verbose, logFormatter)
		buildBasicStore(lgr)
		builder := plugins.NewBuilder(lgr)
		if installAppProxyCmdOptions.host == "" {
			dieOnError(fmt.Errorf("host options is required in order to install app-proxy"))
		}
		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		builderInstallOpt := &plugins.InstallOptions{
			CodefreshHost: cfAPIHost,
			DryRun:        installAppProxyCmdOptions.dryRun,
		}
		s.AppProxy.Resources = installAppProxyCmdOptions.resources
		s.DockerRegistry = installAppProxyCmdOptions.dockerRegistry
		extendStoreWithKubeClient(lgr)
		fillCodefreshAPI(lgr)
		fillKubernetesAPI(lgr, installAppProxyCmdOptions.kube.context, installAppProxyCmdOptions.kube.namespace, false)
		s.AgentAPI = &store.AgentAPI{
			Token: "",
			Id:    "",
		}

		builderInstallOpt.ClusterName = s.KubernetesAPI.ContextName
		builderInstallOpt.KubeBuilder = getKubeClientBuilder(builderInstallOpt.ClusterName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster)
		builderInstallOpt.ClusterNamespace = s.KubernetesAPI.Namespace
		builder.Add(plugins.AppProxyPluginType)

		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)
		for _, p := range builder.Get() {
			values, err = p.Install(builderInstallOpt, values)
			if err != nil {
				dieOnError(err)
			}
		}

		lgr.Info("App proxy installation completed Successfully")
	},
}

func init() {
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	installCommand.AddCommand(installAppProxyCmd)
	installAppProxyCmd.Flags().StringVar(&installAppProxyCmdOptions.dockerRegistry, "docker-registry", "", "The prefix for the container registry that will be used for pulling the required components images. Example: --docker-registry=\"docker.io\"")
	installAppProxyCmd.Flags().StringVar(&installAppProxyCmdOptions.host, "host", "", "Host name that will be used by the ingress to route traffic to the App-Proxy service")
	installAppProxyCmd.Flags().StringVar(&installAppProxyCmdOptions.ingressClass, "ingress-class", "", "The ingress class name that will be used by the App-Proxy ingress. If left empty, the default one will be used")
	installAppProxyCmd.Flags().StringVar(&installAppProxyCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	installAppProxyCmd.Flags().StringVar(&installAppProxyCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	installAppProxyCmd.Flags().StringArrayVarP(&installAppProxyCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")
	installAppProxyCmd.Flags().StringArrayVar(&installAppProxyCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	installAppProxyCmd.Flags().StringArrayVar(&installAppProxyCmdOptions.templateValues, "set-value", []string{}, "Set values for templates, example: --set-value LocalVolumesDir=/mnt/disks/ssd0/codefresh-volumes")
	installAppProxyCmd.Flags().BoolVar(&installAppProxyCmdOptions.dryRun, "dry-run", false, "Set to true to simulate installation")
}
