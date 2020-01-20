package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
	"errors"
	"io/ioutil"
	"encoding/base64"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	sdkUtils "github.com/codefresh-io/go-sdk/pkg/utils"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
	"github.com/codefresh-io/venona/venonactl/pkg/kube"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/olekukonko/tablewriter"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
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

type nodeSelector map[string]string

func parseNodeSelector(s string) (nodeSelector, error) {
	if s == "" {
		return nodeSelector{}, nil
	}
	v := strings.Split(s, "=")
	if len(v) != 2 {
		return nil, errors.New("node selector must be in form \"key=value\"")
	}
	return nodeSelector{v[0]: v[1]}, nil
}
type toleration struct {
	Key      string  `json:key`
    Operator string  `json:operator`
    Effect   string  `json:effect`
} 

func getTolerationFromPath(s string) (string, error)  {
	if s == "" {
		return "", nil
	}
	rawData, err := ioutil.ReadFile(s)
	if err != nil {
		return "", err
	}

	data := []toleration{}
	err = json.Unmarshal([]byte(rawData), &data);
	if (err != nil) {
		return "", errors.New("can not parse tolerations")
	}
	y, err := yaml.Marshal(&data)
	if (err != nil) {
		return "", errors.New("can not marshel tolerations to yaml")
	}
	d := fmt.Sprintf("\n%s", string(y))
	return d, nil
}

func fillKubernetesAPI(lgr logger.Logger, context string, namespace string, inCluster bool)  {

	s := store.GetStore()
	if context == "" {
		config := clientcmd.GetConfigFromFileOrDie(s.KubernetesAPI.ConfigPath)
		context = config.CurrentContext
		lgr.Debug("Kube Context is not set, using current context", "Kube-Context-Name", context)
	}
	if namespace == "" {
		namespace = "default"
	}

	s.KubernetesAPI.InCluster = inCluster
	s.KubernetesAPI.ContextName = context
	s.KubernetesAPI.Namespace = namespace
	
}

func extendStoreWithAgentAPI(logger logger.Logger)  {
	s := store.GetStore()
	logger.Debug("Using agent's token", "Token", installAgentCmdOptions.agentToken)
	s.AgentAPI = &store.AgentAPI{
		Token:  base64.StdEncoding.EncodeToString([]byte(installAgentCmdOptions.agentToken)),
	}
}
