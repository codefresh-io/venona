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
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		lgr := createLogger("Upgrade", true, logFormatter)
		lgr.Warn("Upgrade is not supported from version < 1.0.0 to version >= 1.x.x, please run the migration script: https://github.com/codefresh-io/venona/blob/master/scripts/migration.sh to upgrade to the latest version")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringVar(&upgradeCmdOpt.kube.context, "kube-context-name", "", "Set name to overwrite the context name saved in Codefresh")
	upgradeCmd.Flags().BoolVar(&upgradeCmdOpt.dryRun, "dry-run", false, "Set to to actually upgrade the kubernetes components")
}
