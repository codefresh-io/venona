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

var installAppProxyCmdOptions struct {
	kube struct {
		namespace string
		context   string
	}
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
	resources          map[string]interface{}
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

		mergeValueStr(templateValuesMap, "Namespace", &installAppProxyCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &installAppProxyCmdOptions.kube.context)

		mergeValueMSI(templateValuesMap, "AppProxy.resoruces", &installAppProxyCmdOptions.resources)

		s := store.GetStore()
		lgr := createLogger("Install-agent", verbose, logFormatter)
		buildBasicStore(lgr)
		builder := plugins.NewBuilder(lgr)
		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		builderInstallOpt := &plugins.InstallOptions{
			CodefreshHost: cfAPIHost,
		}
		s.AppProxy.Resources = installAppProxyCmdOptions.resources
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
		spn := createSpinner("Installing app proxy (might take a few minutes)", "")
		spn.Start()
		defer spn.Stop()
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
	installAppProxyCmd.Flags().StringVar(&installAppProxyCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	installAppProxyCmd.Flags().StringVar(&installAppProxyCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	installAppProxyCmd.Flags().StringArrayVarP(&installAppProxyCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

}
