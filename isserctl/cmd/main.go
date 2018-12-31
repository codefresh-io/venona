/*
Copyright 2019 The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/codefresh-io/Isser/isserctl/pkg/codefresh"
	"github.com/codefresh-io/Isser/isserctl/pkg/runtimectl"
	"github.com/codefresh-io/Isser/isserctl/pkg/certs"
	"github.com/golang/glog"
)

const (
	cmdInstall = "install"
	cmdStatus = "status"
	cmdDelete = "delete"
)
var (
	runtimectlType = runtimectl.TypeKubernetesDind
	codefreshAPIKey string
	codefreshURL    string

	kubeconfig  string
	kubecontext string
	namespace   string
	clusterName string

	v string // glog debug
)

func dieIfError(err error) {
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)		
	}
}

func _stringInList(list []string, s string) bool {
	for _, a := range list {
        if a == s {
            return true
        }
    }
    return false
}

func getruntimectlConfig() (*runtimectl.Config, error) {

	var clientConfig runtimectl.ClientConfig
	if kubeconfig == "" {
		currentUser, _ := user.Current()
		kubeconfig = path.Join(currentUser.HomeDir, ".kube", "config")
	}
	if runtimectlType == runtimectl.TypeKubernetesDind {
		clientConfig = runtimectl.ClientConfig{
			KubeClient: runtimectl.KubernetesClientConfig{
				Kubeconfig: &kubeconfig,
				Context:    &kubecontext,
				Namespace:  &namespace,
			},
		}
	} else {
		return nil, fmt.Errorf("Unknown runtimectl type %s", runtimectlType)
	}

	runtimectlConfig := &runtimectl.Config{
		Type:   runtimectlType,
		Name:   clusterName,
		Client: clientConfig,
		ServerCert: &certs.ServerCert{},
	}
	return runtimectlConfig, nil
}

func setCommonFlags(flagset *flag.FlagSet){
	if runtimectlType == runtimectl.TypeKubernetesDind {
		flagset.StringVar(&kubeconfig, "kubeconfig", "", "Absolute path to the kubeconfig")
		flagset.StringVar(&kubecontext, "kubecontext", "", "Kubeconfig context name")
		flagset.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	}
	flagset.Set("alsologtostderr", "true")
	flagset.StringVar(&v, "v", "", "glog debug flag - set -v4 for debug")
}

func doInstall(runtimectlConfig *runtimectl.Config) {
	cfAPI, err := codefresh.NewCfAPI(codefreshURL, codefreshAPIKey)
	dieIfError(err)
	
	err = cfAPI.Validate(runtimectlConfig)
    dieIfError(err)

	err = cfAPI.Sign(runtimectlConfig)
    dieIfError(err)

	ctl, err := runtimectl.GetCtl(runtimectlConfig)
    dieIfError(err)

	err = ctl.Install(runtimectlConfig)
    dieIfError(err)

	err = cfAPI.Register(runtimectlConfig)
    dieIfError(err)

	fmt.Printf("Installation completed Successfully")
}

func printStatus(runtimectlConfig *runtimectl.Config) {
	ctl, err := runtimectl.GetCtl(runtimectlConfig)
    dieIfError(err)
	
	status, err := ctl.GetStatus(runtimectlConfig)
	dieIfError(err)

	fmt.Printf(status.StatusMessage)
    fmt.Printf("\nStatus: %s\n", status.Status)
}

func doDelete(runtimectlConfig *runtimectl.Config) {
	ctl, err := runtimectl.GetCtl(runtimectlConfig)
    dieIfError(err)
	
	err = ctl.Delete(runtimectlConfig)
	dieIfError(err)
}

func main() {

	usage := `
Usage: isserctl <command> [options]

Commands:
	install --api-key <codefresh api-key> --cluster-name <cluster-name> [--url <codefresh url>] [kube params] 
	
	status [kube params]

	delete [kube params]

Options:
   kubeconfig
   kubecontext
   namespace
`
	installCommand := flag.NewFlagSet(cmdInstall, flag.ExitOnError)
    installCommand.StringVar(&codefreshAPIKey, "api-key", "", "Codefresh api key (token)")	
	installCommand.StringVar(&codefreshURL, "url", codefresh.DefaultURL, "Codefresh url")
	installCommand.StringVar(&clusterName, "cluster-name", "", "cluster name")
	setCommonFlags(installCommand)

	statusCommand := flag.NewFlagSet(cmdStatus, flag.ExitOnError)
	setCommonFlags(statusCommand)

	deleteCommand := flag.NewFlagSet(cmdDelete, flag.ExitOnError)
	setCommonFlags(deleteCommand)

	validCommands := []string{cmdInstall, cmdStatus, cmdDelete}
    if len(os.Args) < 2 {
		fmt.Printf("%s\n", usage)
		os.Exit(0)
	} else if !_stringInList(validCommands,os.Args[1]) {
		fmt.Printf("Invalid command %s\n%s", os.Args[1], usage)
		os.Exit(2)
	}

	runtimectlConfig, err := getruntimectlConfig()
	dieIfError(err)

	switch os.Args[1] {
	case cmdInstall:
		installCommand.Parse(os.Args[2:])
		doInstall(runtimectlConfig)
	case cmdStatus:
		statusCommand.Parse(os.Args[2:])
		printStatus(runtimectlConfig)
	case cmdDelete:
		deleteCommand.Parse(os.Args[2:])
		doDelete(runtimectlConfig)	
	default:
		glog.Errorf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
}
