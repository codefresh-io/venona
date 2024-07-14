package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	sdkUtils "github.com/codefresh-io/go-sdk/pkg/utils"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
	"github.com/codefresh-io/venona/venonactl/pkg/kube"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/olekukonko/tablewriter"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	verbose      bool
	logFormatter string

	configPath string
	cfAPIHost  string
	cfAPIToken string
	cfContext  string

	kubeConfigPath string
)

func buildBasicStore(logger logger.Logger) {
	s := store.GetStore()
	s.Version = &store.Version{
		Current: &store.CurrentVersion{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}

	s.Image = &store.Image{
		Name: "codefresh/venona",
	}

	s.Mode = store.ModeInCluster

	s.ServerCert = &certs.ServerCert{}

	s.AppName = store.ApplicationName
	s.Image.Tag = s.Version.Current.Version
}

func extendStoreWithCodefershClient(logger logger.Logger) error {
	s := store.GetStore()
	if configPath == "" {
		currentUser, err := user.Current()
		if err != nil {
			return err
		}

		configPath = filepath.Join(currentUser.HomeDir, ".cfconfig")
		logger.Debug(fmt.Sprint("cfconfig path not set, using: ", configPath))
	}

	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf(".cfconfig file not found")
	}

	if cfAPIHost == "" && cfAPIToken == "" {
		context, err := sdkUtils.ReadAuthContext(configPath, cfContext)
		if err != nil {
			return err
		}
		cfAPIHost = context.URL
		cfAPIToken = context.Token
		logger.Debug("Using codefresh context", "Context-Name", context.Name, "Host", cfAPIHost)
	} else {
		logger.Debug("Reading credentials from environment variables")
		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
	}

	logger.Debug("Creating codefresh client", "host", cfAPIHost, "token", cfAPIToken)

	client := codefresh.New(&codefresh.ClientOptions{
		Auth: codefresh.AuthOptions{
			Token: cfAPIToken,
		},
		Host: cfAPIHost,
	})
	s.CodefreshAPI = &store.CodefreshAPI{
		Host:   cfAPIHost,
		Token:  cfAPIToken,
		Client: client,
	}

	return nil
}

func extendStoreWithKubeClient(logger logger.Logger) {
	s := store.GetStore()
	if kubeConfigPath == "" {
		currentUser, _ := user.Current()
		if currentUser != nil {
			kubeConfigPath = filepath.Join(currentUser.HomeDir, ".kube", "config")
			logger.Debug("Path to kubeconfig not set, using:", "kubeconfig", kubeConfigPath)
		}
	}

	s.KubernetesAPI = &store.KubernetesAPI{
		ConfigPath: kubeConfigPath,
	}
}

func setVerbosity(verbose bool) {
	s := store.GetStore()
	s.Verbose = verbose
}

func setInsecure(insecure bool) {
	s := store.GetStore()
	s.Insecure = insecure
}

func isUsingDefaultStorageClass(sc string) bool {
	if sc == "" {
		return true
	}
	return strings.HasPrefix(sc, plugins.DefaultStorageClassNamePrefix)
}

func dieOnError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func createTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator(" ")
	table.SetColWidth(100)
	return table
}

func getKubeClientBuilder(context string, namespace string, path string, inCluster bool, dryRun bool) kube.Kube {
	return kube.New(&kube.Options{
		ContextName:      context,
		Namespace:        namespace,
		PathToKubeConfig: path,
		InCluster:        inCluster,
		DryRun:           dryRun,
	})
}

func createLogger(command string, verbose bool, logFormatter string) logger.Logger {
	logFile := "venonalog.json"
	os.Remove(logFile)
	return logger.New(&logger.Options{
		Command:      command,
		Verbose:      verbose,
		LogToFile:    logFile,
		LogFormatter: logFormatter,
	})
}
