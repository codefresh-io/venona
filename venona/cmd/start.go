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
	"fmt"
	"time"

	"github.com/codefresh-io/go/venona/pkg/agent"
	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/config"
	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultCodefreshHost = "https://g.codefresh.io"
)

var startCmdOptions struct {
	codefreshToken                 string
	codefreshHost                  string
	verbose                        bool
	agentID                        string
	taskPullingSecondsInterval     int64
	statusReportingSecondsInterval int64
	configDir                      string
	serverPort                     string
}

var startCmd = &cobra.Command{
	Use: "start",

	Run: func(cmd *cobra.Command, args []string) {
		log := logger.New(logger.Options{
			Verbose: startCmdOptions.verbose,
		})
		configs, err := config.Load(startCmdOptions.configDir, ".*.runtime.yaml", log.New("module", "config-loader"))
		dieOnError(err)
		runtimes := map[string]runtime.Runtime{}
		for _, config := range configs {

			k, err := kubernetes.New(kubernetes.Options{
				Token: config.Token,
				Type:  config.Type,
				Host:  config.Host,
				Cert:  config.Cert,
			})
			if err != nil {
				log.Error("Failed to load kubernetes", "error", err.Error())
				continue
			}
			re := runtime.New(runtime.Options{
				Kubernetes: k,
			})
			runtimes[config.Name] = re

		}

		cf := codefresh.New(codefresh.Options{
			Host:    startCmdOptions.codefreshHost,
			Token:   startCmdOptions.codefreshToken,
			AgentID: startCmdOptions.agentID,
			Logger:  log.New("module", "service", "service", "codefresh"),
		})
		// load runtimes

		agent := agent.Agent{
			Codefresh:          cf,
			Logger:             log.New("module", "agent"),
			Runtimes:           runtimes,
			ID:                 startCmdOptions.agentID,
			TaskPullerTicker:   time.NewTicker(time.Duration(startCmdOptions.taskPullingSecondsInterval) * time.Second),
			ReportStatusTicker: time.NewTicker(time.Duration(startCmdOptions.statusReportingSecondsInterval) * time.Second),
		}
		dieOnError(agent.Start())

		server := server.Server{
			Port:   fmt.Sprintf(":%s", startCmdOptions.serverPort),
			Logger: log.New("module", "server"),
		}
		dieOnError(server.Start())

	},
	Long: "Start venona process",
}

func init() {
	dieOnError(viper.BindEnv("codefresh-token", "CODEFRESH_TOKEN"))
	dieOnError(viper.BindEnv("codefresh-host", "CODEFRESH_HOST"))
	dieOnError(viper.BindEnv("agent-id", "AGENT_ID"))
	dieOnError(viper.BindEnv("config-dir", "CONFIG_DIR"))
	dieOnError(viper.BindEnv("port", "PORT"))

	viper.SetDefault("codefresh-host", defaultCodefreshHost)
	viper.SetDefault("port", "8080")

	startCmd.Flags().BoolVar(&startCmdOptions.verbose, "verbose", true, "Show more logs")
	startCmd.Flags().StringVar(&startCmdOptions.agentID, "agent-id", viper.GetString("agent-id"), "ID of the agent [$AGENT_ID]")
	startCmd.Flags().StringVar(&startCmdOptions.configDir, "config-dir", viper.GetString("config-dir"), "path to configuration folder [$CONFIG_DIR]")
	startCmd.Flags().StringVar(&startCmdOptions.codefreshToken, "codefresh-token", viper.GetString("codefresh-token"), "Codefresh API token [$CODEFRESH_TOKEN]")
	startCmd.Flags().StringVar(&startCmdOptions.serverPort, "port", viper.GetString("port"), "The port to start the server [$PORT]")
	startCmd.Flags().StringVar(&startCmdOptions.codefreshHost, "codefresh-host", viper.GetString("codefresh-host"), "Codefresh API host default [$CODEFRESH_HOST]")
	startCmd.Flags().Int64Var(&startCmdOptions.taskPullingSecondsInterval, "task-pulling-interval", 3, "The interval to pull new tasks from Codefresh")
	startCmd.Flags().Int64Var(&startCmdOptions.statusReportingSecondsInterval, "status-reporting-interval", 10, "The interval to report status back to Codefresh")

	startCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			startCmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})

	dieOnError(startCmd.MarkFlagRequired("codefresh-token"))
	dieOnError(startCmd.MarkFlagRequired("agent-id"))
	dieOnError(startCmd.MarkFlagRequired("port"))

	rootCmd.AddCommand(startCmd)
}
