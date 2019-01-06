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

	codefreshSDK "github.com/codefresh-io/go-sdk/pkg/codefresh"
	"github.com/codefresh-io/isser/isserctl/pkg/certs"
	"github.com/codefresh-io/isser/isserctl/pkg/codefresh"
	"github.com/codefresh-io/isser/isserctl/pkg/runtimectl"
	"github.com/codefresh-io/isser/isserctl/pkg/store"
	"github.com/golang/glog"
)

const (
	cmdInstall = "install"
	cmdStatus  = "status"
	cmdDelete  = "delete"
)

var (
	runtimectlType  = runtimectl.TypeKubernetesDind
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
		fmt.Printf("Error: %v\n", err)
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

func buildDataStore() {

	if kubeconfig == "" {
		currentUser, _ := user.Current()
		kubeconfig = path.Join(currentUser.HomeDir, ".kube", "config")
	}

	s := store.GetStore()
	s.AppName = store.ApplicationName
	s.KubernetesAPI = &store.KubernetesAPI{
		Namespace:   namespace,
		ConfigPath:  kubeconfig,
		ContextName: kubecontext,
	}
	s.ClusterInCodefresh = clusterName
	s.CodefreshAPI = &store.CodefreshAPI{
		Host:  codefreshURL,
		Token: codefreshAPIKey,
		Client: codefreshSDK.New(&codefreshSDK.ClientOptions{
			Auth: codefreshSDK.AuthOptions{
				Token: codefreshAPIKey,
			},
			Host: codefreshURL,
		}),
	}
	s.Mode = store.ModeInCluster
	s.Image = &store.Image{
		Name: "codefresh/isser",
		Tag:  store.GetLatestVersion(),
	}
	s.ServerCert = &certs.ServerCert{}
}

func defineCommonFlags(flagset *flag.FlagSet) {
	if runtimectlType == runtimectl.TypeKubernetesDind {
		flagset.StringVar(&kubeconfig, "kubeconfig", "", "Absolute path to the kubeconfig")
		flagset.StringVar(&kubecontext, "kubecontext", "", "Kubeconfig context name")
		flagset.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	}
	flagset.StringVar(&clusterName, "cluster-name", "", "cluster name (if not passed runtime-environment will be created cluster-less)")
	flagset.StringVar(&v, "v", "2", "glog debug flag - set -v4 for debug")
}

func processFlags() {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", v)
}

func doInstall() {
	cfAPI := codefresh.New()

	err := cfAPI.Validate()
	dieIfError(err)

	err = cfAPI.Sign()
	dieIfError(err)

	err = cfAPI.Register()
	dieIfError(err)

	err = runtimectl.GetCtl().Install()
	dieIfError(err)

	fmt.Printf("Installation completed Successfully\n")
}

// func printStatus(runtimectlConfig *runtimectl.Config) {
// 	ctl, err := runtimectl.GetCtl(runtimectlConfig)
// 	dieIfError(err)

// 	status, err := ctl.GetStatus(runtimectlConfig)
// 	dieIfError(err)

// 	fmt.Printf(status.StatusMessage)
// 	fmt.Printf("\nStatus: %s\n", status.Status)
// }

// func doDelete(runtimectlConfig *runtimectl.Config) {
// 	ctl, err := runtimectl.GetCtl(runtimectlConfig)
// 	dieIfError(err)

// 	err = ctl.Delete(runtimectlConfig)
// 	dieIfError(err)
// }

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
	flag.Parse()

	installCommandFlagset := flag.NewFlagSet(cmdInstall, flag.ExitOnError)
	installCommandFlagset.StringVar(&codefreshAPIKey, "api-key", "", "Codefresh api key (token)")
	installCommandFlagset.StringVar(&codefreshURL, "url", codefresh.DefaultURL, "Codefresh url")

	// statusCommandFlagset := flag.NewFlagSet(cmdStatus, flag.ExitOnError)

	// deleteCommandFlagset := flag.NewFlagSet(cmdDelete, flag.ExitOnError)

	validCommands := []string{cmdInstall, cmdStatus, cmdDelete}
	if len(os.Args) < 2 {
		fmt.Printf("%s\n", usage)
		os.Exit(0)
	} else if !_stringInList(validCommands, os.Args[1]) {
		fmt.Printf("Invalid command %s\n%s", os.Args[1], usage)
		os.Exit(2)
	}

	var flagset *flag.FlagSet
	var cmdFunction func()

	switch os.Args[1] {
	case cmdInstall:
		flagset = installCommandFlagset
		cmdFunction = doInstall
	// case cmdStatus:
	// 	flagset = statusCommandFlagset
	// 	cmdFunction = printStatus
	// case cmdDelete:
	// 	flagset = deleteCommandFlagset
	// 	cmdFunction = doDelete
	default:
		glog.Errorf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	defineCommonFlags(flagset)
	flagset.Parse(os.Args[2:])
	processFlags()
	buildDataStore()
	cmdFunction()

}
