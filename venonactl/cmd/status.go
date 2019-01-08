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
	"github.com/codefresh-io/venona/venonactl/internal"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/sirupsen/logrus"

	runtimectl "github.com/codefresh-io/venona/venonactl/pkg/operators"
	"github.com/spf13/cobra"
)

var headers = []string{"Kind", "Name", "Status", "Message"}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Get status of Codefresh's runtime-environment",
	Long:  "pass name of the runtime-environment to get staus reported by venona's agent to Codefresh",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			name := args[0]
			re, err := store.GetStore().CodefreshAPI.Client.GetRuntimeEnvironment(name)
			internal.DieOnError(err)
			if re.Metadata.Agent == true {
				logrus.WithField("Updated_At", re.Status.UpdatedAt).Infof("Venona last reported message: %s", re.Status.Message)
			} else {
				logrus.Info("Runtime wasnt configured with Venona's agent")
			}
		}

		if verbose == true {
			table := internal.CreateTable()
			table.SetHeader(headers)

			rows, err := runtimectl.GetOperator(runtimectl.RuntimeEnvironmentOperatorType).Status()
			internal.DieOnError(err)
			table.AppendBulk(rows)

			rows, err = runtimectl.GetOperator(runtimectl.VenonaOperatorType).Status()
			internal.DieOnError(err)
			table.AppendBulk(rows)

			logrus.Infof("\n\nKubernetes resources:")
			table.Render()
		}

	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
