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
	"errors"

	"github.com/codefresh-io/venona/venonactl/internal"
	"github.com/codefresh-io/venona/venonactl/pkg/operators"
	runtimectl "github.com/codefresh-io/venona/venonactl/pkg/operators"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var upgradeCmdOpt struct {
	kube struct {
		context string
	}
	dryRun bool
}

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade [name]",
	Short: "Upgrade existing runtime-environment",
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
		s := store.GetStore()
		prepareLogger()
		buildBasicStore()
		extendStoreWithCodefershClient()
		extendStoreWithKubeClient()

		re, _ := s.CodefreshAPI.Client.RuntimeEnvironments().Get(args[0])
		contextName := re.RuntimeScheduler.Cluster.ClusterProvider.Selector
		if upgradeCmdOpt.kube.context != "" {
			contextName = upgradeCmdOpt.kube.context
		}
		s.KubernetesAPI.ContextName = contextName
		s.KubernetesAPI.Namespace = re.RuntimeScheduler.Cluster.Namespace
		if upgradeCmdOpt.dryRun {
			logrus.Info("Running in dry-run mode")
		} else {
			operators.GetOperator(operators.VenonaOperatorType).Upgrade()
			if isUsingDefaultStorageClass(re.RuntimeScheduler.Pvcs.Dind.StorageClassName) {
				err := runtimectl.GetOperator(runtimectl.VolumeProvisionerOperatorType).Delete()
				internal.DieOnError(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringVar(&upgradeCmdOpt.kube.context, "kube-context-name", "", "Set name to overwrite the context name saved in Codefresh")
	upgradeCmd.Flags().BoolVar(&upgradeCmdOpt.dryRun, "dry-run", false, "Set to to actually upgrade the kubernetes components")
}
