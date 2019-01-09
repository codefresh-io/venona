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

	"github.com/sirupsen/logrus"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/internal"
	runtimectl "github.com/codefresh-io/venona/venonactl/pkg/operators"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	humanize "github.com/dustin/go-humanize"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Get status of Codefresh's runtime-environment",
	Long:  "Pass the name of the runtime environment to see more details information about the underlying resources",
	Run: func(cmd *cobra.Command, args []string) {
		s := store.GetStore()
		table := internal.CreateTable()

		// When requested status for specific runtime
		if len(args) > 0 {
			name := args[0]
			re, err := s.CodefreshAPI.Client.GetRuntimeEnvironment(name)
			internal.DieOnError(err)
			if re.Metadata.Agent == true {
				table.SetHeader([]string{"Runtime Name", "Last Message", "Reported"})
				message := "Not reported any message yet"
				time := ""
				if re.Status.Message != "" {
					message = re.Status.Message
					time = humanize.Time(re.Status.UpdatedAt)
				}
				table.Append([]string{re.Metadata.Name, message, time})
				table.Render()
				fmt.Println()
				printTableWithKubernetesRelatedResources(re)
			} else {
				logrus.Warnf("%s has not Venona's agent", re.Metadata.Name)
			}
			return
		}

		// When requested status for all runtimes
		res, err := s.CodefreshAPI.Client.GetRuntimeEnvironments()
		internal.DieOnError(err)
		table.SetHeader([]string{"Runtime Name", "Last Message", "Reported"})
		for _, re := range res {
			if re.Metadata.Agent == true {
				message := "Not reported any message yet"
				time := ""
				if re.Status.Message != "" {
					message = re.Status.Message
					time = humanize.Time(re.Status.UpdatedAt)
				}
				table.Append([]string{re.Metadata.Name, message, time})
			}
		}
		table.Render()

		return

	},
}

func printTableWithKubernetesRelatedResources(re *codefresh.RuntimeEnvironment) {
	table := internal.CreateTable()
	table.SetHeader([]string{"Kind", "Name", "Status"})
	s := store.GetStore()
	if re.RuntimeScheduler.Cluster.Namespace != "" {
		s.KubernetesAPI.ContextName = re.RuntimeScheduler.Cluster.ClusterProvider.Selector
		s.KubernetesAPI.Namespace = re.RuntimeScheduler.Cluster.Namespace

		rows, err := runtimectl.GetOperator(runtimectl.RuntimeEnvironmentOperatorType).Status()
		internal.DieOnError(err)
		table.AppendBulk(rows)
		rows, err = runtimectl.GetOperator(runtimectl.VenonaOperatorType).Status()
		internal.DieOnError(err)
		table.AppendBulk(rows)
	}
	table.Render()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
