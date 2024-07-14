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
<<<<<<< HEAD
=======
	"errors"

>>>>>>> master
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var upgradeCmdOpt struct {
	kube struct {
		context   string
		namespace string
	}
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
}

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
<<<<<<< HEAD
	Use:   "upgrade",
	Short: "Upgrade existing 1.X runner",
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

		lgr := createLogger("Upgrade", verbose, logFormatter)
		builder := plugins.NewBuilder(lgr)
		builder.Add(plugins.VenonaPluginType)

		s := store.GetStore()
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		dieOnError(extendStoreWithCodefershClient(lgr))
		extendStoreWithAgentAPI(lgr, "", "")
		fillKubernetesAPI(lgr, upgradeCmdOpt.kube.context, upgradeCmdOpt.kube.namespace, false)
		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)

		spn := createSpinner("Upgarding runtime (might take a few seconds)", "")
		spn.Start()
		defer spn.Stop()

		for _, p := range builder.Get() {
			values, err = p.Upgrade(cmd.Context(), &plugins.UpgradeOptions{
				Name:             s.AppName,
				ClusterNamespace: upgradeCmdOpt.kube.namespace,
				ClusterName:      upgradeCmdOpt.kube.namespace,
				KubeBuilder:      getKubeClientBuilder(upgradeCmdOpt.kube.context, upgradeCmdOpt.kube.namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster, false),
			}, values)
=======
	Use:   "upgrade [name]",
	Short: "Upgrade existing runtime-environment",
	Deprecated: "Venona binary has been deprecated, please use codefresh cli  https://codefresh.io/docs/docs/administration/codefresh-runner ",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires name of the runtime-environment")
		}

		if len(args) > 1 {
			return errors.New("Cannot upgrade multiple runtimes once")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		return
		s := store.GetStore()
		lgr := createLogger("Upgrade", verbose)
		buildBasicStore(lgr)
		extendStoreWithCodefershClient(lgr)
		extendStoreWithKubeClient(lgr)
		builder := plugins.NewBuilder(lgr)
		builderUpgradeOpt := &plugins.UpgradeOptions{
			CodefreshHost:  s.CodefreshAPI.Host,
			CodefreshToken: s.CodefreshAPI.Token,
			DryRun:         upgradeCmdOpt.dryRun,
			Name:           s.AppName,
		}

		re, _ := s.CodefreshAPI.Client.RuntimeEnvironments().Get(args[0])
		contextName := re.RuntimeScheduler.Cluster.ClusterProvider.Selector
		if upgradeCmdOpt.kube.context != "" {
			contextName = upgradeCmdOpt.kube.context
		}
		s.KubernetesAPI.ContextName = contextName
		s.KubernetesAPI.Namespace = re.RuntimeScheduler.Cluster.Namespace

		builderUpgradeOpt.ClusterNamespace = s.KubernetesAPI.Namespace

		if upgradeCmdOpt.dryRun {
			lgr.Info("Running in dry-run mode")
		} else {
			builder.Add(plugins.VenonaPluginType)
			if isUsingDefaultStorageClass(re.RuntimeScheduler.Pvcs.Dind.StorageClassName) {
				builder.Add(plugins.VolumeProvisionerPluginType)
			}
			builder.Add(plugins.RuntimeEnvironmentPluginType)
		}

		builderUpgradeOpt.KubeBuilder = getKubeClientBuilder(upgradeCmdOpt.kube.context, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster)

		var err error
		values := s.BuildValues()
		for _, p := range builder.Get() {
			values, err = p.Upgrade(builderUpgradeOpt, values)
>>>>>>> master
			if err != nil {
				dieOnError(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringVar(&upgradeCmdOpt.kube.context, "kube-context-name", "", "Set name to overwrite the context name saved in Codefresh")
	upgradeCmd.Flags().StringVar(&upgradeCmdOpt.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona is installed [$KUBE_NAMESPACE]")

	upgradeCmd.Flags().StringArrayVar(&upgradeCmdOpt.templateValues, "set-value", []string{}, "Set values for templates --set-value agentId=12345")
	upgradeCmd.Flags().StringArrayVar(&upgradeCmdOpt.templateFileValues, "set-file", []string{}, "Set values for templates from file")
	upgradeCmd.Flags().StringArrayVarP(&upgradeCmdOpt.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")
}
