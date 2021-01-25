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

var allTestPluginTypes = []string{
	plugins.RuntimeEnvironmentPluginType,
	plugins.VenonaPluginType,
	plugins.MonitorAgentPluginType,
	plugins.VolumeProvisionerPluginType,
	plugins.EnginePluginType,
	plugins.RuntimeAttachType,
	plugins.NetworkTesterPluginType,
}

var testCommandOptions struct {
	kube struct {
		namespace string
		context   string
		inCluster bool
	}
	insecure           bool
	plugin             []string
	templateValues     []string
	templateFileValues []string
	templateValueFiles []string
}

var testCommand = &cobra.Command{
	Use:   "test",
	Short: "Run test on the target cluster prior installation",
	Run: func(cmd *cobra.Command, args []string) {
		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file>
		templateValuesMap, err := templateValuesToMap(
			testCommandOptions.templateValueFiles,
			testCommandOptions.templateValues,
			testCommandOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Token", &cfAPIToken)
		mergeValueStr(templateValuesMap, "Namespace", &testCommandOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &testCommandOptions.kube.context)
		mergeValueBool(templateValuesMap, "InCluster", &testCommandOptions.kube.inCluster)
		mergeValueBool(templateValuesMap, "insecure", &testCommandOptions.insecure)

		lgr := createLogger("test", verbose, logFormatter)
		s := store.GetStore()
		buildBasicStore(lgr)
		extendStoreWithKubeClient(lgr)
		fillKubernetesAPI(lgr, testCommandOptions.kube.context, testCommandOptions.kube.namespace, testCommandOptions.kube.inCluster)
		fillCodefreshAPI(lgr)
		extendStoreWithAgentAPI(lgr, "", "")
		setVerbosity(verbose)
		setInsecure(testCommandOptions.insecure)

		builder := plugins.NewBuilder(lgr)
		for _, p := range testCommandOptions.plugin {
			if p == plugins.RuntimeEnvironmentPluginType {
				builder.Add(plugins.RuntimeEnvironmentPluginType)
			}
			if p == plugins.VenonaPluginType {
				builder.Add(plugins.VenonaPluginType)
			}
			if p == plugins.MonitorAgentPluginType {
				builder.Add(plugins.MonitorAgentPluginType)
			}
			if p == plugins.VolumeProvisionerPluginType {
				builder.Add(plugins.VolumeProvisionerPluginType)
			}
			if p == plugins.EnginePluginType {
				builder.Add(plugins.EnginePluginType)
			}
			if p == plugins.RuntimeAttachType {
				builder.Add(plugins.RuntimeAttachType)
			}
			if p == plugins.NetworkTesterPluginType {
				builder.Add(plugins.NetworkTesterPluginType)
			}
		}

		options := plugins.TestOptions{
			KubeBuilder:      getKubeClientBuilder(s.KubernetesAPI.ContextName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster, false),
			ClusterNamespace: s.KubernetesAPI.Namespace,
		}

		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)

		var finalerr error
		lgr.Info("Testing requirements")

		for _, p := range builder.Get() {
			err := p.Test(options, values)
			if err != nil && finalerr == nil {
				finalerr = err
			}
		}
		dieOnError(finalerr)

		lgr.Info("Cluster passed acceptance test")
	},
}

func init() {
	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")

	testCommand.Flags().StringVar(&testCommandOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which monitor should be installed [$KUBE_NAMESPACE]")
	testCommand.Flags().StringVar(&testCommandOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which monitor should be installed (default is current-context) [$KUBE_CONTEXT]")
	testCommand.Flags().BoolVar(&testCommandOptions.insecure, "insecure", false, "Set to true to disable certificate validation when using TLS connections")
	testCommand.Flags().BoolVar(&testCommandOptions.kube.inCluster, "in-cluster", false, "Set flag if the command is running inside a cluster")
	testCommand.Flags().StringArrayVar(&testCommandOptions.plugin, "installer", allTestPluginTypes, "Which test to run, based on the installer type")

	testCommand.Flags().StringArrayVar(&testCommandOptions.templateValues, "set-value", []string{}, "Set values for templates, example: --set-value LocalVolumesDir=/mnt/disks/ssd0/codefresh-volumes")
	testCommand.Flags().StringArrayVar(&testCommandOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file, example: --set-file Storage.GoogleServiceAccount=/path/to/service-account.json")
	testCommand.Flags().StringArrayVarP(&testCommandOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")

	rootCmd.AddCommand(testCommand)
}
