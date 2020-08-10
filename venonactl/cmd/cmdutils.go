package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"encoding/json"

	"github.com/briandowns/spinner"
	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	sdkUtils "github.com/codefresh-io/go-sdk/pkg/utils"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
	"github.com/codefresh-io/venona/venonactl/pkg/kube"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
	k8sApi "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"helm.sh/helm/v3/pkg/strvals"
	cliValues "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"github.com/stretchr/objx"	
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

func getKubeClientBuilder(context string, namespace string, path string, inCluster bool) kube.Kube {
	return kube.New(&kube.Options{
		ContextName:      context,
		Namespace:        namespace,
		PathToKubeConfig: path,
		InCluster:        inCluster,
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

func createSpinner(prefix, suffix string) *spinner.Spinner {
	s := spinner.New([]string{"   ", ".  ", ".. ", "..."}, 520*time.Millisecond)
	s.Suffix = suffix
	s.Prefix = prefix
	return s
}

func loadTolerationsFromFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		dieOnError(err)
	}

	return string(data)
}

func parseTolerations(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	var data []k8sApi.Toleration
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		return "", fmt.Errorf("can not parse tolerations: %s", err)
	}
	y, err := yaml.Marshal(&data)
	if err != nil {
		return "", fmt.Errorf("can not marshel tolerations to yaml: %s", err)
	}
	d := fmt.Sprintf("\n%s", string(y))
	return d, nil
}

func fillKubernetesAPI(lgr logger.Logger, context string, namespace string, inCluster bool) {

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

func extendStoreWithAgentAPI(logger logger.Logger, token string, agentID string) {
	s := store.GetStore()
	logger.Debug("Using agent's token", "Token", token)
	s.AgentAPI = &store.AgentAPI{
		Token: base64.StdEncoding.EncodeToString([]byte(token)),
		Id:    agentID,
	}
}

// Parsing helpers --set-value , --set-file
// by https://github.com/helm/helm/blob/ec1d1a3d3eb672232f896f9d3b3d0797e4f519e3/pkg/cli/values/options.go#L41

// templateValuesToMap - processes cmd --values <values-file.yaml> --set-value k=v --set-file v=<context-of-file> 
// using helm libraries
func templateValuesToMap(templateValueFiles, templateValues, templateFileValues []string) (map[string]interface{}, error) {
	valueOpts := &cliValues.Options{}
	if len(templateValueFiles) > 0 {
		for _, v := range(templateValueFiles) {
			valueOpts.ValueFiles = append(valueOpts.ValueFiles, v)
		}
	}
	
	if len(templateValues) > 0 {
		for _, v := range(templateValues) {
			valueOpts.Values = append(valueOpts.Values, v)
		}			
	}
	
	if len(templateFileValues) > 0 {
		for _, v := range(templateFileValues) {
			valueOpts.FileValues = append(valueOpts.FileValues, v)
		}
	}
	valuesMap, err := valueOpts.MergeValues(getter.Providers{})
	return valuesMap, err
}


// parses --set-value options
func parseSetValues(setValuesOpts []string) (map[string]interface{}, error) {
	base := map[string]interface{}{}
	for _, value := range setValuesOpts {
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, fmt.Errorf("Cannot parse option --set-value %s", value)
		}
	}
	return base, nil
}

// parses --set-file options
func parseSetFiles(setFilesOpts []string) (map[string]interface{}, error) {
	base := map[string]interface{}{}
	for _, value := range setFilesOpts {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := ioutil.ReadFile(string(rs))
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return nil, fmt.Errorf("Cannot parse option --set-file %s", value)
		}
	}
	return base, nil
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// paramOrValueStr - for merging cli parameters with mapped parameters
func paramOrValueStr(valuesMap map[string]interface{}, key string, param *string) {
	mapX := objx.New(valuesMap)
	if param != nil && *param != "" {
		//mapX.Set(key)
	    return
    }
   
    val := mapX.Get(key).String()
    *param = val
}

// paramOrValueBool - for merging cli parameters with mapped parameters
func paramOrValueBool(valuesMap map[string]interface{}, key string, param *bool) {
	if param != nil ||*param == true {
		return
	}
	mapX := objx.New(valuesMap)
	val := mapX.Get(key).Bool()
	*param = val
 }