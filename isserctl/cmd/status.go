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
package cmd

import (
	"github.com/codefresh-io/isser/isserctl/internal"

	runtimectl "github.com/codefresh-io/isser/isserctl/pkg/operators"
	"github.com/spf13/cobra"
)

var headers = []string{"Kind", "Name", "Status", "Message"}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get status of Codefresh's runtime-environment",
	Run: func(cmd *cobra.Command, args []string) {
		table := internal.CreateTable()
		table.SetHeader(headers)

		rows, err := runtimectl.GetOperator(runtimectl.RuntimeEnvironmentOperatorType).Status()
		internal.DieOnError(err)
		table.AppendBulk(rows)

		rows, err = runtimectl.GetOperator(runtimectl.IsserOperatorType).Status()
		internal.DieOnError(err)
		table.AppendBulk(rows)

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
