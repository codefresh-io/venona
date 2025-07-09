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
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codefresh-io/go/venona/pkg/agent"
	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/config"
	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/metrics"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/codefresh-io/go/venona/pkg/monitoring/newrelic"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/codefresh-io/go/venona/pkg/server"

	nr "github.com/newrelic/go-agent/v3/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type startOptions struct {
	codefreshToken                 string
	codefreshHost                  string
	verbose                        bool
	rejectTLSUnauthorized          bool
	agentID                        string
	taskPullingSecondsInterval     int64
	statusReportingSecondsInterval int64
	concurrency                    int
	bufferSize                     int
	configDir                      string
	serverPort                     string
	newrelicLicenseKey             string
	newrelicAppname                string
	inClusterRuntime               string
	qps                            float32
	burst                          int
	forceDeletePvc                 bool
}

const (
	defaultCodefreshHost           = "https://g.codefresh.io"
	defaultTaskPullingInterval     = 3
	defaultStatusReportingInterval = 10
	defaultWorkflowConcurrency     = 50
	defaultWorkflowBufferSize      = 1000
	defaultK8sClientQPS            = 50
	defaultK8sClientBurst          = 100
	defaultForceDeletePvc          = false
)

var (
	startCmdOptions startOptions
	handleSignal    = signal.Notify
)

var startCmd = &cobra.Command{
	Use:  "start",
	Long: "Start venona process",
	PreRunE: func(_ *cobra.Command, _ []string) error {
		if startCmdOptions.taskPullingSecondsInterval <= 0 {
			return errors.New("--task-pulling-interval must be a positive number")
		}

		if startCmdOptions.statusReportingSecondsInterval <= 0 {
			return errors.New("--status-reporting-interval must be a positive number")
		}

		if startCmdOptions.concurrency <= 0 {
			return errors.New("--workflow-concurrency must be a positive number")
		}

		if startCmdOptions.bufferSize <= 0 {
			return errors.New("--workflow-buffer-size must be a positive number")
		}

		if startCmdOptions.qps <= 0 {
			return errors.New("--k8s-client-qps must be a positive number")
		}

		if startCmdOptions.burst <= 0 {
			return errors.New("--k8s-client-burst must be a positive number")
		}

		return nil
	},
	Run: func(_ *cobra.Command, _ []string) {
		run(startCmdOptions)
	},
}

func init() {
	dieOnError(viper.BindEnv("codefresh-token", "CODEFRESH_TOKEN"))
	dieOnError(viper.BindEnv("codefresh-host", "CODEFRESH_HOST"))
	dieOnError(viper.BindEnv("in-cluster-runtime", "CODEFRESH_IN_CLUSTER_RUNTIME"))
	dieOnError(viper.BindEnv("agent-id", "AGENT_ID"))
	dieOnError(viper.BindEnv("config-dir", "VENONA_CONFIG_DIR"))
	dieOnError(viper.BindEnv("port", "PORT"))
	dieOnError(viper.BindEnv("NODE_TLS_REJECT_UNAUTHORIZED"))
	dieOnError(viper.BindEnv("verbose", "VERBOSE"))
	dieOnError(viper.BindEnv("newrelic-license-key", "NEWRELIC_LICENSE_KEY"))
	dieOnError(viper.BindEnv("newrelic-appname", "NEWRELIC_APPNAME"))
	dieOnError(viper.BindEnv("task-pulling-interval", "TASK_PULLING_INTERVAL"))
	dieOnError(viper.BindEnv("status-reporting-interval", "STATUS_REPORTING_INTERVAL"))
	dieOnError(viper.BindEnv("workflow-concurrency", "WORKFLOW_CONCURRENCY"))
	dieOnError(viper.BindEnv("workflow-buffer-size", "WORKFLOW_BUFFER_SIZE"))
	dieOnError(viper.BindEnv("k8s-client-qps", "K8S_CLIENT_QPS"))
	dieOnError(viper.BindEnv("k8s-client-burst", "K8S_CLIENT_BURST"))
	dieOnError(viper.BindEnv("force-delete-pvc", "FORCE_DELETE_PVC"))

	viper.SetDefault("codefresh-host", defaultCodefreshHost)
	viper.SetDefault("port", "8080")
	viper.SetDefault("NODE_TLS_REJECT_UNAUTHORIZED", "1")
	viper.SetDefault("in-cluster-runtime", "")
	viper.SetDefault("newrelic-appname", AppName)
	viper.SetDefault("task-pulling-interval", defaultTaskPullingInterval)
	viper.SetDefault("status-reporting-interval", defaultStatusReportingInterval)
	viper.SetDefault("workflow-concurrency", defaultWorkflowConcurrency)
	viper.SetDefault("workflow-buffer-size", defaultWorkflowBufferSize)
	viper.SetDefault("k8s-client-qps", defaultK8sClientQPS)
	viper.SetDefault("k8s-client-burst", defaultK8sClientBurst)
	viper.SetDefault("force-delete-pvc", defaultForceDeletePvc)

	startCmd.Flags().BoolVar(&startCmdOptions.verbose, "verbose", viper.GetBool("verbose"), "Show more logs")
	startCmd.Flags().BoolVar(&startCmdOptions.rejectTLSUnauthorized, "tls-reject-unauthorized", viper.GetBool("NODE_TLS_REJECT_UNAUTHORIZED"), "Disable certificate validation for TLS connections")
	startCmd.Flags().StringVar(&startCmdOptions.inClusterRuntime, "in-cluster-runtime", viper.GetString("in-cluster-runtime"), "Runtime name to run agent in cluster mode [$CODEFRESH_IN_CLUSTER_RUNTIME]")
	startCmd.Flags().StringVar(&startCmdOptions.agentID, "agent-id", viper.GetString("agent-id"), "ID of the agent [$AGENT_ID]")
	startCmd.Flags().StringVar(&startCmdOptions.configDir, "config-dir", viper.GetString("config-dir"), "path to configuration folder [$CONFIG_DIR]")
	startCmd.Flags().StringVar(&startCmdOptions.codefreshToken, "codefresh-token", viper.GetString("codefresh-token"), "Codefresh API token [$CODEFRESH_TOKEN]")
	startCmd.Flags().StringVar(&startCmdOptions.serverPort, "port", viper.GetString("port"), "The port to start the server [$PORT]")
	startCmd.Flags().StringVar(&startCmdOptions.codefreshHost, "codefresh-host", viper.GetString("codefresh-host"), "Codefresh API host default [$CODEFRESH_HOST]")
	startCmd.Flags().Int64Var(&startCmdOptions.taskPullingSecondsInterval, "task-pulling-interval", viper.GetInt64("task-pulling-interval"), "The interval (seconds) to pull new tasks from Codefresh [$TASK_PULLING_INTERVAL]")
	startCmd.Flags().Int64Var(&startCmdOptions.statusReportingSecondsInterval, "status-reporting-interval", viper.GetInt64("status-reporting-interval"), "The interval (seconds) to report status back to Codefresh [$STATUS_REPORTING_INTERVAL]")
	startCmd.Flags().IntVar(&startCmdOptions.concurrency, "workflow-concurrency", viper.GetInt("workflow-concurrency"), "How many workflow tasks to handle concurrently [$WORKFLOW_CONCURRENCY]")
	startCmd.Flags().IntVar(&startCmdOptions.bufferSize, "workflow-buffer-size", viper.GetInt("workflow-cbuffer-sizeoncurrency"), "The size of the workflow channel buffer [$WORKFLOW_BUFFER_SIZE]")
	startCmd.Flags().StringVar(&startCmdOptions.newrelicLicenseKey, "newrelic-license-key", viper.GetString("newrelic-license-key"), "New-Relic license key [$NEWRELIC_LICENSE_KEY]")
	startCmd.Flags().StringVar(&startCmdOptions.newrelicAppname, "newrelic-appname", viper.GetString("newrelic-appname"), "New-Relic application name [$NEWRELIC_APPNAME]")
	startCmd.Flags().Float32Var(&startCmdOptions.qps, "k8s-client-qps", float32(viper.GetFloat64("k8s-client-qps")), "the maximum QPS to the master from this client [$K8S_CLIENT_QPS]")
	startCmd.Flags().IntVar(&startCmdOptions.burst, "k8s-client-burst", viper.GetInt("k8s-client-burst"), "k8s client maximum burst for throttle [$K8S_CLIENT_BURST]")
	startCmd.Flags().BoolVar(&startCmdOptions.forceDeletePvc, "force-delete-pvc", viper.GetBool("force-delete-pvc"), "set to true to disable PVC protection [$FORCE_DELETE_PVC]")

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

	log.Debug("Starting", "pid", os.Getpid(), "version", version)
	if !options.rejectTLSUnauthorized {
		log.Warn("Running in insecure mode", "NODE_TLS_REJECT_UNAUTHORIZED", options.rejectTLSUnauthorized)
	}

	reg := prometheus.NewRegistry()
	metrics.Register(reg)

	var runtimes map[string]runtime.Runtime
	k8sLog := log.New("module", "k8s")
	if options.inClusterRuntime != "" {
		runtimes = inClusterRuntimeConfiguration(options, k8sLog)
	} else {
		runtimes = remoteRuntimeConfiguration(options, k8sLog)
	}

	monitor := monitoring.NewEmpty()
	var err error

	if options.newrelicLicenseKey != "" {
		monitor, err = newrelic.New(
			nr.ConfigAppName(options.newrelicAppname),
			nr.ConfigLicense(options.newrelicLicenseKey),
		)
		if err != nil {
			log.Warn("Failed to create monitor", "error", err)
		} else {
			log.Info("Using New Relic monitor", "app-name", options.newrelicAppname, "license-key", options.newrelicLicenseKey)
		}
	} else {
		log.Warn("New Relic not starting without license key!")
	}

	var cf codefresh.Codefresh
	{
		var httpClient http.Client
		if !options.rejectTLSUnauthorized {
			customTransport := http.DefaultTransport.(*http.Transport).Clone()
			// #nosec
			customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			httpClient = http.Client{
				Transport: customTransport,
			}
		}

		httpClient.Transport = monitor.NewRoundTripper(httpClient.Transport)

		userAgent := fmt.Sprintf("cf-classic-runner/%s", version)
		if runtimeVersion := os.Getenv("RUNTIME_CHART_VERSION"); runtimeVersion != "" {
			userAgent += fmt.Sprintf(" cf-classic-runtime/%s", runtimeVersion)
		}
		httpHeaders := http.Header{}
		{
			httpHeaders.Add("User-Agent", userAgent)
		}

		cf = codefresh.New(codefresh.Options{
			Host:       options.codefreshHost,
			Token:      options.codefreshToken,
			AgentID:    options.agentID,
			HTTPClient: &httpClient,
			Headers:    httpHeaders,
		})
	}

	agent, err := agent.New(&agent.Options{
		Codefresh:                      cf,
		Logger:                         log.New("module", "agent"),
		Runtimes:                       runtimes,
		ID:                             options.agentID,
		TaskPullingSecondsInterval:     time.Duration(options.taskPullingSecondsInterval) * time.Second,
		StatusReportingSecondsInterval: time.Duration(options.statusReportingSecondsInterval) * time.Second,
		Monitor:                        monitor,
		Concurrency:                    options.concurrency,
		BufferSize:                     options.bufferSize,
	})
	dieOnError(err)

	server, err := server.New(&server.Options{
		Port:            fmt.Sprintf(":%s", options.serverPort),
		Logger:          log.New("module", "server"),
		Monitor:         monitor,
		MetricsRegistry: reg,
	})
	dieOnError(err)

	ctx := context.Background()

	ctx = withSignals(ctx, server.Stop, agent.Stop, log)
	go func() { dieOnError(agent.Start(ctx)) }()
	go func() { dieOnError(server.Start()) }()

	<-ctx.Done()
}

func inClusterRuntimeConfiguration(options startOptions, log logger.Logger) map[string]runtime.Runtime {
	k, err := kubernetes.NewInCluster(log, options.qps, options.burst, options.forceDeletePvc)
	dieOnError(err)
	re := runtime.New(runtime.Options{
		Kubernetes: k,
	})
	return map[string]runtime.Runtime{options.inClusterRuntime: re}
}

func remoteRuntimeConfiguration(options startOptions, log logger.Logger) map[string]runtime.Runtime {
	configs, err := config.Load(options.configDir, ".*.runtime.yaml", log.New("module", "config-loader"))
	dieOnError(err)
	runtimes := map[string]runtime.Runtime{}
	for name, config := range configs {
		k, err := kubernetes.New(kubernetes.Options{
			Logger:         log,
			Token:          config.Token,
			Type:           config.Type,
			Host:           config.Host,
			Cert:           config.Cert,
			Insecure:       !options.rejectTLSUnauthorized,
			QPS:            options.qps,
			Burst:          options.burst,
			ForceDeletePvc: options.forceDeletePvc,
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

	return runtimes
}

func withSignals(
	ctx context.Context,
	stopServer func(context.Context) error,
	stopAgent func() error,
	log logger.Logger,
) context.Context {
	var terminationReq int32 = 0
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 10)

	handleSignal(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {
			<-sigChan
			if terminationReq++; terminationReq > 1 {
				// signal received more than once, forcing termination
				log.Warn("Forcing termination!")
				cancel()
				return
			}

			go func() {
				log.Warn("Received shutdown request, stopping agent and server...")
				// order matters, the process will exit as soon as server is stopped
				if err := stopAgent(); err != nil {
					log.Error(err.Error())
				}

				if err := stopServer(ctx); err != nil {
					log.Error(err.Error())
				}

				cancel() // done
			}()
		}
	}()

	return ctx
}
