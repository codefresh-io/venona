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

var installRuntimeCmdOptions struct {
	codefreshToken string
	dryRun         bool
	insecure       bool
	kube           struct {
		namespace    string
		inCluster    bool
		context      string
		nodeSelector string
	}
	storageClass           string
	dockerRegistry         string
	runtimeEnvironmentName string
	kubernetesRunnerType   bool
	tolerations            string
	templateValues         []string
	templateFileValues     []string
	templateValueFiles     []string
}

var installRuntimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Install Codefresh's runtime",
	Run: func(cmd *cobra.Command, args []string) {

		// get valuesMap from --values <values.yaml> --set-value k=v --set-file k=<context-of file> 
		templateValuesMap, err := templateValuesToMap(
			installRuntimeCmdOptions.templateValueFiles, 
			installRuntimeCmdOptions.templateValues, 
			installRuntimeCmdOptions.templateFileValues)
		if err != nil {
			dieOnError(err)
		}
		// Merge cmd options with template
		mergeValueStr(templateValuesMap, "ConfigPath", &kubeConfigPath)
		mergeValueStr(templateValuesMap, "CodefreshHost", &cfAPIHost)
		mergeValueStr(templateValuesMap, "Token", &cfAPIToken)
		mergeValueStr(templateValuesMap, "Token", &installRuntimeCmdOptions.codefreshToken)		

		mergeValueStr(templateValuesMap, "Namespace", &installRuntimeCmdOptions.kube.namespace)
		mergeValueStr(templateValuesMap, "Context", &installRuntimeCmdOptions.kube.context)
		mergeValueStr(templateValuesMap, "RuntimeEnvironmentName", &installRuntimeCmdOptions.runtimeEnvironmentName)
		mergeValueStr(templateValuesMap, "NodeSelector", &installRuntimeCmdOptions.kube.nodeSelector)
		mergeValueStr(templateValuesMap, "Tolerations", &installRuntimeCmdOptions.tolerations)
		//mergeValueStrArray(&installAgentCmdOptions.envVars, "envVars", nil, "More env vars to be declared \"key=value\"")
		mergeValueStr(templateValuesMap, "DockerRegistry", &installRuntimeCmdOptions.dockerRegistry)
		mergeValueStr(templateValuesMap, "StorageClass", &installRuntimeCmdOptions.storageClass)
		
		mergeValueBool(templateValuesMap, "InCluster", &installRuntimeCmdOptions.kube.inCluster)
		mergeValueBool(templateValuesMap, "insecure", &installRuntimeCmdOptions.insecure)
		mergeValueBool(templateValuesMap, "kubernetesRunnerType", &installRuntimeCmdOptions.kubernetesRunnerType)


		s := store.GetStore()
		lgr := createLogger("Install-runtime", verbose, logFormatter)
		buildBasicStore(lgr)
		extendStoreWithAgentAPI(lgr, installRuntimeCmdOptions.codefreshToken, "")
		extendStoreWithKubeClient(lgr)

		if installRuntimeCmdOptions.runtimeEnvironmentName == "" {
			dieOnError(fmt.Errorf("Codefresh envrionment name is required"))
		}
		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		// This is temporarily and used for signing
		s.CodefreshAPI = &store.CodefreshAPI{
			Host: cfAPIHost,
		}

		if installRuntimeCmdOptions.tolerations != "" {
			var tolerationsString string

			if installRuntimeCmdOptions.tolerations[0] == '@' {
				tolerationsString = loadTolerationsFromFile(installRuntimeCmdOptions.tolerations[1:])
			} else {
				tolerationsString = installRuntimeCmdOptions.tolerations
			}

			tolerations, err := parseTolerations(tolerationsString)
			if err != nil {
				dieOnError(err)
			}

			s.KubernetesAPI.Tolerations = tolerations
		}

		s.KubernetesAPI.NodeSelector = installRuntimeCmdOptions.kube.nodeSelector
		s.DockerRegistry = installRuntimeCmdOptions.dockerRegistry

		builder := plugins.NewBuilder(lgr)
		isDefault := isUsingDefaultStorageClass(installRuntimeCmdOptions.storageClass)

		builderInstallOpt := &plugins.InstallOptions{
			StorageClass:          installRuntimeCmdOptions.storageClass,
			IsDefaultStorageClass: isDefault,
			DryRun:                installRuntimeCmdOptions.dryRun,
			KubernetesRunnerType:  installRuntimeCmdOptions.kubernetesRunnerType,
			CodefreshHost:         cfAPIHost,
			CodefreshToken:        installRuntimeCmdOptions.codefreshToken,
			RuntimeEnvironment:    installRuntimeCmdOptions.runtimeEnvironmentName,
			ClusterNamespace:      installRuntimeCmdOptions.kube.namespace,
			Insecure:              installRuntimeCmdOptions.insecure,
		}

		if installRuntimeCmdOptions.kubernetesRunnerType {
			builder.Add(plugins.EnginePluginType)
		}

		if isDefault {
			builderInstallOpt.StorageClass = plugins.DefaultStorageClassNamePrefix
		}

		fillKubernetesAPI(lgr, installRuntimeCmdOptions.kube.context, installRuntimeCmdOptions.kube.namespace, installRuntimeCmdOptions.kube.inCluster)

		if installRuntimeCmdOptions.dryRun {
			s.DryRun = installRuntimeCmdOptions.dryRun
			lgr.Info("Running in dry-run mode")
		}

		// s.ClusterInCodefresh = installRuntimeCmdOptions.clusterNameInCodefresh

		builder.Add(plugins.RuntimeEnvironmentPluginType)

		if isDefault {
			builder.Add(plugins.VolumeProvisionerPluginType)
		} else {
			lgr.Info("Custom StorageClass is set, skipping installation of default volume provisioner")
		}

		builderInstallOpt.KubeBuilder = getKubeClientBuilder(s.KubernetesAPI.ContextName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster)
		values := s.BuildValues()
		values = mergeMaps(values, templateValuesMap)

		for _, p := range builder.Get() {
			values, err = p.Install(builderInstallOpt, values)
			if err != nil {
				dieOnError(err)
			}
		}
		lgr.Info("Runtime installation completed Successfully")

	},
}

func init() {
	installCommand.AddCommand(installRuntimeCmd)

	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")

	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.codefreshToken, "codefreshToken", "", "Codefresh token")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.runtimeEnvironmentName, "runtimeName", viper.GetString("runtimeName"), "Name of the runtime as in codefresh")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.storageClass, "storage-class", "", "Set a name of your custom storage class, note: this will not install volume provisioning components")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.dockerRegistry, "docker-registry", "", "The prefix for the container registry that will be used for pulling the required components images. Example: --docker-registry=\"docker.io\"")

	installRuntimeCmd.Flags().BoolVar(&installRuntimeCmdOptions.insecure, "insecure", false, "Set to true to disable TLS when comunicating with the codefresh platform")
	installRuntimeCmd.Flags().BoolVar(&installRuntimeCmdOptions.kube.inCluster, "in-cluster", false, "Set flag if venona is been installed from inside a cluster")
	installRuntimeCmd.Flags().BoolVar(&installRuntimeCmdOptions.dryRun, "dry-run", false, "Set to true to simulate installation")
	installRuntimeCmd.Flags().BoolVar(&installRuntimeCmdOptions.kubernetesRunnerType, "kubernetes-runner-type", false, "Set the runner type to kubernetes (alpha feature)")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.kube.nodeSelector, "kube-node-selector", "", "The kubernetes node selector \"key=value\" to be used by venona resources (default is no node selector)")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.tolerations, "tolerations", "", "The kubernetes tolerations as JSON string to be used by venona resources (default is no tolerations)")

	installRuntimeCmd.Flags().StringArrayVar(&installRuntimeCmdOptions.templateValues, "set-value", []string{}, "Set values for templates, example: --set-value LocalVolumesDir=/mnt/disks/ssd0/codefresh-volumes")
	installRuntimeCmd.Flags().StringArrayVar(&installRuntimeCmdOptions.templateFileValues, "set-file", []string{}, "Set values for templates from file, example: --set-file Storage.GoogleServiceAccount=/path/to/service-account.json")
	installRuntimeCmd.Flags().StringArrayVarP(&installRuntimeCmdOptions.templateValueFiles, "values", "f", []string{}, "specify values in a YAML file")
}
