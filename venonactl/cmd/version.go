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

	"github.com/codefresh-io/venona/venonactl/pkg/store"

	"github.com/spf13/cobra"
)

// variables been set with ldflags flag
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Venona version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Date: %s\n", date)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Local Version: %s\n", version)
		fmt.Printf("Latest version: %s\n", store.GetLatestVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
