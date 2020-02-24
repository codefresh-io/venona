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
	"github.com/spf13/cobra"
)

const (
	clusterNameMaxLength = 20
	namespaceMaxLength   = 20
)

var installCmdOptions struct {
	dryRun                 bool
	clusterNameInCodefresh string
	kube                   struct {
		namespace    string
		inCluster    bool
		context      string
		nodeSelector string
	}
	storageClass string
	venona       struct {
		version string
	}
	setDefaultRuntime             bool
	installOnlyRuntimeEnvironment bool
	skipRuntimeInstallation       bool
	runtimeEnvironmentName        string
	kubernetesRunnerType          bool
	buildNodeSelector             string
	buildAnnotations              []string
	tolerationJSONString          string
}

// installVenonaCmd represents the install command
var installVenonaCmd = &cobra.Command{
	Use:   "all",
	Short: "Install Codefresh's resource on cluster",
}

func init() {
	installCommand.AddCommand(installVenonaCmd)
}
