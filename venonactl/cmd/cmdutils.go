package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"
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
	// set to false by default, when running hack/build.sh will change to true
	// to prevent version checking during development
	localDevFlow = "false"

	verbose bool

	configPath string
	cfAPIHost  string
	cfAPIToken string
	cfContext  string

	kubeConfigPath string

	skipVerionCheck bool
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

	if skipVerionCheck || localDevFlow == "true" {
		latestVersion := &store.LatestVersion{
			Version:   store.DefaultVersion,
			IsDefault: true,
		}
		s.Version.Latest = latestVersion
		logger.Debug("Skipping version check")
	} else {
		latestVersion := &store.LatestVersion{
			Version:   store.GetLatestVersion(logger),
			IsDefault: false,
		}
		s.Image.Tag = latestVersion.Version
		s.Version.Latest = latestVersion
		res, _ := store.IsRunningLatestVersion()
		// the local version and the latest version not match
		// make sure the command is no venonactl version
		if !res {
			logger.Info("New version is avaliable, please update",
				"Local-Version", s.Version.Current.Version,
				"Latest-Version", s.Version.Latest.Version)
		}
	}
}

func extendStoreWithCodefershClient(logger logger.Logger) error {
	s := store.GetStore()
	if configPath == "" {
		configPath = fmt.Sprintf("%s/.cfconfig", os.Getenv("HOME"))
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
		logger.Debug("Reading creentials from environment variables")
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
			kubeConfigPath = path.Join(currentUser.HomeDir, ".kube", "config")
			logger.Debug("Path to kubeconfig not set, using default")
		}
	}

	s.KubernetesAPI = &store.KubernetesAPI{
		ConfigPath: kubeConfigPath,
	}
}

func isUsingDefaultStorageClass(sc string) bool {
	if sc == "" {
		return true
	}
	return strings.HasPrefix(sc, plugins.DefaultStorageClassNamePrefix)
}

func dieOnError(err error) {
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
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

func getKubeClientBuilder(context string, namespace string, path string, inCluster bool) kube.Kube {
	return kube.New(&kube.Options{
		ContextName:      context,
		Namespace:        namespace,
		PathToKubeConfig: path,
		InCluster:        inCluster,
	})
}

func createLogger(command string, verbose bool) logger.Logger {
	logFile := "venonalog.json"
	os.Remove(logFile)
	return logger.New(&logger.Options{
		Command:   command,
		Verbose:   verbose,
		LogToFile: logFile,
	})
}
