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
	"github.com/sirupsen/logrus"

	runtimectl "github.com/codefresh-io/isser/isserctl/pkg/operators"
	"github.com/spf13/cobra"
)

// deleteCmd represents the status command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Codefresh's runtime-environment",
	Run: func(cmd *cobra.Command, args []string) {
		err := runtimectl.GetOperator(runtimectl.RuntimeEnvironmentOperatorType).Delete()
		internal.DieOnError(err)

		err = runtimectl.GetOperator(runtimectl.IsserOperatorType).Delete()
		internal.DieOnError(err)
		logrus.Info("Deletion completed")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
