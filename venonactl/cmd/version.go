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

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Venona version",
	Run: func(cmd *cobra.Command, args []string) {
		s := store.GetStore()
		buildBasicStore()
		fmt.Printf("Date: %s\n", s.Version.Current.Date)
		fmt.Printf("Commit: %s\n", s.Version.Current.Commit)
		fmt.Printf("Local Version: %s\n", s.Version.Current.Version)
		fmt.Printf("Latest version: %s\n", s.Version.Latest.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
