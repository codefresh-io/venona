package cmd

import (
	"time"

	"github.com/codefresh-io/go/venona/pkg/agent"
	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/config"
	"github.com/codefresh-io/go/venona/pkg/logger"
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

	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New(logger.Options{
			Verbose: startCmdOptions.verbose,
		})

		_, err := config.Load(startCmdOptions.configDir, ".*.runtime.yaml", log.New("module", "config-loader"))
		dieOnError(err)

		cf := codefresh.New(codefresh.Options{
			Host:    startCmdOptions.codefreshHost,
			Token:   startCmdOptions.codefreshToken,
			AgentID: startCmdOptions.agentID,
			Logger:  log.New("module", "service", "service", "codefresh"),
		})
		agent := agent.Agent{
			Codefresh:          cf,
			Logger:             log.New("module", "agent"),
			ID:                 startCmdOptions.agentID,
			TaskPullerTicker:   time.NewTicker(time.Duration(startCmdOptions.taskPullingSecondsInterval) * time.Second),
			ReportStatusTicker: time.NewTicker(time.Duration(startCmdOptions.statusReportingSecondsInterval) * time.Second),
		}
		dieOnError(agent.Start())
		return nil

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
	startCmd.PersistentFlags().StringVar(&startCmdOptions.codefreshToken, "codefresh-token", viper.GetString("codefreshToken"), "Codefresh API token [$CODEFRESH_TOKEN]")
	startCmd.PersistentFlags().StringVar(&startCmdOptions.codefreshHost, "codefresh-host", viper.GetString("codefreshHost"), "Codefresh API host default [$CODEFRESH_HOST]")
	startCmd.PersistentFlags().Int64Var(&startCmdOptions.taskPullingSecondsInterval, "task-pulling-interval", 3, "The interval to pull new tasks from Codefresh")
	startCmd.PersistentFlags().Int64Var(&startCmdOptions.statusReportingSecondsInterval, "status-reporting-interval", 10, "The interval to report status back to Codefresh")

	startCmd.MarkFlagRequired("codefresh-token")
	startCmd.MarkFlagRequired("agent-id")

	rootCmd.AddCommand(startCmd)
}
