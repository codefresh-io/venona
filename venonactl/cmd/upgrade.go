// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"

	"github.com/codefresh-io/venona/venonactl/pkg/operators"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
)

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
		kubeContextFlag := cmd.Flag("kube-context-name")
		re, _ := s.CodefreshAPI.Client.RuntimeEnvironments().Get(args[0])
		contextName := re.RuntimeScheduler.Cluster.ClusterProvider.Selector
		if kubeContextFlag != nil {
			contextName = kubeContextFlag.Value.String()
		}
		s.KubernetesAPI.ContextName = contextName
		s.KubernetesAPI.Namespace = re.RuntimeScheduler.Cluster.Namespace
		operators.GetOperator(operators.VenonaOperatorType).Upgrade()
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
