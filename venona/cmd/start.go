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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

type startOptions struct {
	codefreshToken                 string
	codefreshHost                  string
	verbose                        bool
	rejectTLSUnauthorized          bool
	agentID                        string
	taskPullingSecondsInterval     int64
	statusReportingSecondsInterval int64
	configDir                      string
	serverPort                     string
}

var (
	startCmdOptions startOptions
	handleSignal    = signal.Notify
	exit            = os.Exit
)

var startCmd = &cobra.Command{
	Use: "start",

	Run: func(cmd *cobra.Command, args []string) {
		run(startCmdOptions)
	},
	Long: "Start venona process",
}

func init() {
	dieOnError(viper.BindEnv("codefresh-token", "CODEFRESH_TOKEN"))
	dieOnError(viper.BindEnv("codefresh-host", "CODEFRESH_HOST"))
	dieOnError(viper.BindEnv("agent-id", "AGENT_ID"))
	dieOnError(viper.BindEnv("config-dir", "VENONA_CONFIG_DIR"))
	dieOnError(viper.BindEnv("port", "PORT"))
	dieOnError(viper.BindEnv("NODE_TLS_REJECT_UNAUTHORIZED"))

	viper.SetDefault("codefresh-host", defaultCodefreshHost)
	viper.SetDefault("port", "8080")
	viper.SetDefault("NODE_TLS_REJECT_UNAUTHORIZED", "1")

	startCmd.Flags().BoolVar(&startCmdOptions.verbose, "verbose", false, "Show more logs")
	startCmd.Flags().BoolVar(&startCmdOptions.rejectTLSUnauthorized, "tls-reject-unauthorized", viper.GetBool("NODE_TLS_REJECT_UNAUTHORIZED"), "Disable certificate validation for TLS connections")
	startCmd.Flags().StringVar(&startCmdOptions.agentID, "agent-id", viper.GetString("agent-id"), "ID of the agent [$AGENT_ID]")
	startCmd.Flags().StringVar(&startCmdOptions.configDir, "config-dir", viper.GetString("config-dir"), "path to configuration folder [$CONFIG_DIR]")
	startCmd.Flags().StringVar(&startCmdOptions.codefreshToken, "codefresh-token", viper.GetString("codefresh-token"), "Codefresh API token [$CODEFRESH_TOKEN]")
	startCmd.Flags().StringVar(&startCmdOptions.serverPort, "port", viper.GetString("port"), "The port to start the server [$PORT]")
	startCmd.Flags().StringVar(&startCmdOptions.codefreshHost, "codefresh-host", viper.GetString("codefresh-host"), "Codefresh API host default [$CODEFRESH_HOST]")
	startCmd.Flags().Int64Var(&startCmdOptions.taskPullingSecondsInterval, "task-pulling-interval", 3, "The interval (seconds) to pull new tasks from Codefresh")
	startCmd.Flags().Int64Var(&startCmdOptions.statusReportingSecondsInterval, "status-reporting-interval", 10, "The interval (seconds) to report status back to Codefresh")

	startCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			dieOnError(startCmd.Flags().Set(f.Name, viper.GetString(f.Name)))
		}
	})

	dieOnError(startCmd.MarkFlagRequired("codefresh-token"))
	dieOnError(startCmd.MarkFlagRequired("agent-id"))
	dieOnError(startCmd.MarkFlagRequired("port"))

	rootCmd.AddCommand(startCmd)
}

func run(options startOptions) {
	log := logger.New(logger.Options{
		Verbose: options.verbose,
	})

	log.Debug("Starting", "pid", os.Getpid())
	if !options.rejectTLSUnauthorized {
		log.Warn("Running in insecure mode", "NODE_TLS_REJECT_UNAUTHORIZED", options.rejectTLSUnauthorized)
	}

	configs, err := config.Load(options.configDir, ".*.runtime.yaml", log.New("module", "config-loader"))
	dieOnError(err)
	runtimes := map[string]runtime.Runtime{}
	{
		for name, config := range configs {
			k, err := kubernetes.New(kubernetes.Options{
				Token:    config.Token,
				Type:     config.Type,
				Host:     config.Host,
				Cert:     config.Cert,
				Insecure: !options.rejectTLSUnauthorized,
			})
			if err != nil {
				log.Error("Failed to load kubernetes", "error", err.Error(), "file", name, "name", config.Name)
				continue
			}
			re := runtime.New(runtime.Options{
				Kubernetes: k,
			})
			runtimes[config.Name] = re
		}
	}
	var cf codefresh.Codefresh
	{

		var httpClient http.Client
		if !options.rejectTLSUnauthorized {
			// #nosec
			tr := http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			httpClient = http.Client{
				Transport: &tr,
			}
		}
		cf = codefresh.New(codefresh.Options{
			Host:       options.codefreshHost,
			Token:      options.codefreshToken,
			AgentID:    options.agentID,
			Logger:     log.New("module", "service", "service", "codefresh"),
			HTTPClient: &httpClient,
		})
	}

	agent, err := agent.New(&agent.Options{
		Codefresh:                      cf,
		Logger:                         log.New("module", "agent"),
		Runtimes:                       runtimes,
		ID:                             options.agentID,
		TaskPullingSecondsInterval:     time.Duration(options.taskPullingSecondsInterval) * time.Second,
		StatusReportingSecondsInterval: time.Duration(options.statusReportingSecondsInterval) * time.Second,
	})
	dieOnError(err)

	serverMode := server.Release
	if options.verbose {
		serverMode = server.Debug
	}
	server, err := server.New(&server.Options{
		Port:   fmt.Sprintf(":%s", options.serverPort),
		Logger: log.New("module", "server"),
		Mode:   serverMode,
	})
	dieOnError(err)

	go handleSignals(server.Stop, agent.Stop, log)
	dieOnError(agent.Start())
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		dieOnError(err)
	}
}

func handleSignals(stopServer, stopAgent func() error, log logger.Logger) {
	sigChan := make(chan os.Signal, 10)
	receivedTerminationReq := false
	receivedTerminationReqMux := sync.Mutex{}

	handleSignal(sigChan, syscall.SIGTERM, syscall.SIGINT)

	for {
		switch <-sigChan {
		case syscall.SIGTERM, syscall.SIGINT:
			go func(stopServer, stopAgent func() error, log logger.Logger) {
				// check if should perform force shutdown
				shouldExit := false
				receivedTerminationReqMux.Lock()
				if receivedTerminationReq {
					shouldExit = true
				}
				receivedTerminationReq = true
				receivedTerminationReqMux.Unlock()

				if shouldExit { // perform force shutdown
					log.Warn("forcing termination!")
					exit(1)
				}

				log.Warn("Received shutdown request, stopping agent and server...")
				// order matters, the process will exit as soon as server is stopped
				if err := stopAgent(); err != nil {
					log.Error(err.Error())
				}
				if err := stopServer(); err != nil {
					log.Error(err.Error())
				}
				return
			}(stopServer, stopAgent, log)
		}
	}
}
