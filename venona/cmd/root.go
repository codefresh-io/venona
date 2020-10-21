// Copyright 2020 The Codefresh Authors.
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
	"github.com/spf13/cobra"
)

const (
	// AppName holds the name of the application, to be used in monitoring tools
	AppName = "Codefresh-Runner"
)

var version string

var rootCmd = &cobra.Command{
	Use:     "venona",
	Version: version,
	Long:    "Codefresh agent process",
}

// Execute - execute the root command
func Execute() {
	err := rootCmd.Execute()
	dieOnError(err)
}
