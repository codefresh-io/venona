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

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	humanize "github.com/dustin/go-humanize"

	"github.com/spf13/cobra"
)

var statusCmdOpt struct {
	kube struct {
		context string
	}
	dryRun bool
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Get status of Codefresh's runtime-environment",
	Long:  "Pass the name of the runtime environment to see more details information about the underlying resources",
	Run: func(cmd *cobra.Command, args []string) {
		lgr := createLogger("Status", verbose, logFormatter)

		buildBasicStore(lgr)
		extendStoreWithCodefershClient(lgr)
		extendStoreWithKubeClient(lgr)
		s := store.GetStore()
		table := createTable()
		// When requested status for specific runtime
		if len(args) > 0 {
			name := args[0]
			re, err := s.CodefreshAPI.Client.RuntimeEnvironments().Get(name)
			dieOnError(err)
			if re == nil {
				dieOnError(fmt.Errorf("Runtime-Environment %s not found", name))
			}
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
				printTableWithKubernetesRelatedResources(re, statusCmdOpt.kube.context, lgr)
			} else {
				lgr.Debug("Runtime-Environment has not Venona's agent", "Name", name)
			}
			return
		}

		// When requested status for all runtimes
		res, err := s.CodefreshAPI.Client.RuntimeEnvironments().List()
		dieOnError(err)
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

func printTableWithKubernetesRelatedResources(re *codefresh.RuntimeEnvironment, context string, logger logger.Logger) {
	builder := plugins.NewBuilder(logger)

	table := createTable()
	table.SetHeader([]string{"Kind", "Name", "Status"})
	s := store.GetStore()
	if re.RuntimeScheduler.Cluster.Namespace != "" {
		if context == "" {
			context = re.RuntimeScheduler.Cluster.ClusterProvider.Selector
		}
		s.KubernetesAPI.ContextName = context
		s.KubernetesAPI.Namespace = re.RuntimeScheduler.Cluster.Namespace
		builder.
			Add(plugins.RuntimeEnvironmentPluginType).
			Add(plugins.VenonaPluginType).
			Add(plugins.VolumeProvisionerPluginType)
		statusOpt := &plugins.StatusOptions{
			KubeBuilder:      getKubeClientBuilder(context, re.RuntimeScheduler.Cluster.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster, false),
			ClusterNamespace: s.KubernetesAPI.Namespace,
		}
		for _, p := range builder.Get() {
			rows, err := p.Status(statusOpt, s.BuildValues())
			dieOnError(err)
			table.AppendBulk(rows)
		}
	}
	table.Render()
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVar(&statusCmdOpt.kube.context, "kube-context-name", "", "Set name to overwrite the context name saved in Codefresh")
}
