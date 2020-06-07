package cmd

import (
	"time"

	"github.com/codefresh-io/go/venona/pkg/agent"
	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/config"
	"github.com/codefresh-io/go/venona/pkg/kubernetes"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/runtime"
	"github.com/spf13/cobra"
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
				Type: config.Type,
				Host: config.Host,
				Name: config.Name,
				Cert: config.Cert,
			})
			if err != nil {
				log.Error("Failed to load kubernetes with error", "error", err.Error())
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
			Runtimes: 			runtimes,
			ID:                 startCmdOptions.agentID,
			TaskPullerTicker:   time.NewTicker(time.Duration(startCmdOptions.taskPullingSecondsInterval) * time.Second),
			ReportStatusTicker: time.NewTicker(time.Duration(startCmdOptions.statusReportingSecondsInterval) * time.Second),
		}
		dieOnError(agent.Start())

	},
	Long: "Start venona process",
}

func init() {
	viper.BindEnv("codefreshToken", "CODEFRESH_TOKEN")
	viper.BindEnv("codefreshHost", "CODEFRESH_HOST")
	viper.BindEnv("agentID", "AGENT_ID")

	viper.SetDefault("codefreshHost", defaultCodefreshHost)

	startCmd.PersistentFlags().BoolVar(&startCmdOptions.verbose, "verbose", true, "Show more logs")
	startCmd.PersistentFlags().StringVar(&startCmdOptions.agentID, "agent-id", viper.GetString("agentID"), "ID of the agent [$AGENT_ID]")
	startCmd.PersistentFlags().StringVar(&startCmdOptions.configDir, "config-dir", viper.GetString("configDir"), "path to configuration folder")
	startCmd.PersistentFlags().StringVar(&startCmdOptions.codefreshToken, "codefresh-token", viper.GetString("codefreshToken"), "Codefresh API token [$CODEFRESH_TOKEN]")
	startCmd.PersistentFlags().StringVar(&startCmdOptions.codefreshHost, "codefresh-host", viper.GetString("codefreshHost"), "Codefresh API host default [$CODEFRESH_HOST]")
	startCmd.PersistentFlags().Int64Var(&startCmdOptions.taskPullingSecondsInterval, "task-pulling-interval", 3, "The interval to pull new tasks from Codefresh")
	startCmd.PersistentFlags().Int64Var(&startCmdOptions.statusReportingSecondsInterval, "status-reporting-interval", 10, "The interval to report status back to Codefresh")

	startCmd.MarkFlagRequired("codefresh-token")
	startCmd.MarkFlagRequired("agent-id")

	rootCmd.AddCommand(startCmd)
}
