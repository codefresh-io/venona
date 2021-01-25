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

var upgradeAppProxyCmdOptions struct {
	kube struct {
		context   string
		namespace string
	}
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
}

var upgradeAppProxyCmd = &cobra.Command{
	Use:   "app-proxy",
	Short: "Upgrade App proxy",
	Run: func(cmd *cobra.Command, args []string) {

		templateValuesMap, err := templateValuesToMap(
			upgradeAppProxyCmdOptions.templateValueFiles,
			upgradeAppProxyCmdOptions.templateValues,
			upgradeAppProxyCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}

		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Namespace", &upgradeAppProxyCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &upgradeAppProxyCmdOptions.kube.context)

		lgr := createLogger("Upgrade-AppProxy", verbose, logFormatter)
		builder := plugins.NewBuilder(lgr)

		builder.Add(plugins.AppProxyPluginType)

		s := store.GetStore()
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		dieOnError(extendStoreWithCodefershClient(lgr))
		extendStoreWithAgentAPI(lgr, "", "")
		fillKubernetesAPI(lgr, upgradeAppProxyCmdOptions.kube.context, upgradeAppProxyCmdOptions.kube.namespace, false)
		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)

		for _, p := range builder.Get() {
			values, err = p.Upgrade(&plugins.UpgradeOptions{
				Name:             store.AppProxyApplicationName,
				ClusterNamespace: upgradeAppProxyCmdOptions.kube.namespace,
				ClusterName:      upgradeAppProxyCmdOptions.kube.namespace,
				KubeBuilder:      getKubeClientBuilder(upgradeAppProxyCmdOptions.kube.context, upgradeAppProxyCmdOptions.kube.namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster, false),
			}, values)
			if err != nil {
				dieOnError(err)
			}
		}
	},
}

func init() {
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")
	upgradeCmd.AddCommand(upgradeAppProxyCmd)
	upgradeAppProxyCmd.Flags().StringVar(&upgradeAppProxyCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	upgradeAppProxyCmd.Flags().StringVar(&upgradeAppProxyCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	upgradeAppProxyCmd.Flags().StringArrayVarP(&upgradeAppProxyCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")
	upgradeAppProxyCmd.Flags().StringArrayVar(&upgradeAppProxyCmdOptions.templateValues, "set-value", []string{}, "Set values for templates, example: --set-value LocalVolumesDir=/mnt/disks/ssd0/codefresh-volumes")
}
