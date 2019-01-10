package cmd

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

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/spf13/viper"

	"github.com/codefresh-io/venona/venonactl/internal"

	"github.com/sirupsen/logrus"

	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	sdkUtils "github.com/codefresh-io/go-sdk/pkg/utils"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var verbose bool
var skipVerionCheck bool

// variables been set with ldflags flag
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	// set to false by default, when running hack/build.sh will change to true
	// to prevent version checking during development
	localDevFlow = "false"
)

var rootCmd = &cobra.Command{
	Use:   "venonactl",
	Short: "A command line application for Codefresh",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		fullPath := cmd.CommandPath()
		if verbose == true {
			logrus.SetLevel(logrus.DebugLevel)
		}

		configPath := cmd.Flag("cfconfig").Value.String()
		if configPath == "" {
			configPath = fmt.Sprintf("%s/.cfconfig", os.Getenv("HOME"))
		}
		context, err := sdkUtils.ReadAuthContext(configPath, cmd.Flag("context").Value.String())
		if err != nil {
			return err
		}
		logrus.WithFields(logrus.Fields{
			"Context-Name":   context.Name,
			"Codefresh-Host": context.URL,
		}).Debug("Using codefresh context")
		client := codefresh.New(&codefresh.ClientOptions{
			Auth: codefresh.AuthOptions{
				Token: context.Token,
			},
			Host: context.URL,
		})

		kubeContextName := cmd.Flag("kube-context-name").Value.String()
		kubeConfigPath := cmd.Flag("kube-config-path").Value.String()
		kubeNamespace := cmd.Flag("kube-namespace").Value.String()

		if kubeConfigPath == "" {
			currentUser, _ := user.Current()
			kubeConfigPath = path.Join(currentUser.HomeDir, ".kube", "config")
			logrus.WithFields(logrus.Fields{
				"Kube-Config-Path": kubeConfigPath,
			}).Debug("Path to kubeconfig not set, using default")
		}

		if kubeContextName == "" {
			config := clientcmd.GetConfigFromFileOrDie(kubeConfigPath)
			kubeContextName = config.CurrentContext
			logrus.WithFields(logrus.Fields{
				"Kube-Config-Path":  kubeConfigPath,
				"Kube-Context-Name": kubeContextName,
			}).Debug("Kube Context is not set, using current context")
		}

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
		if skipVerionCheck || localDevFlow == "true" {
			latestVersion := &store.LatestVersion{
				Version:   store.DefaultVersion,
				IsDefault: true,
			}
			s.Version.Latest = latestVersion
			logrus.WithFields(logrus.Fields{
				"Default-Version": store.DefaultVersion,
				"Image-Tag":       s.Version.Current.Version,
			}).Debug("Skipping version check")
		} else {
			latestVersion := &store.LatestVersion{
				Version:   store.GetLatestVersion(),
				IsDefault: false,
			}
			s.Image.Tag = latestVersion.Version
			s.Version.Latest = latestVersion
			res, _ := store.IsRunningLatestVersion()
			// the local version and the latest version not match
			// make sure the command is no venonactl version
			if !res && strings.Index(fullPath, "version") == -1 {
				logrus.WithFields(logrus.Fields{
					"Local-Version":  s.Version.Current.Version,
					"Latest-Version": s.Version.Latest.Version,
				}).Info("New version is avaliable, please update")
			}
		}
		s.AppName = store.ApplicationName
		s.KubernetesAPI = &store.KubernetesAPI{
			Namespace:   kubeNamespace,
			ConfigPath:  kubeConfigPath,
			ContextName: kubeContextName,
		}
		s.ClusterInCodefresh = clusterName
		s.CodefreshAPI = &store.CodefreshAPI{
			Host:   context.URL,
			Token:  context.Token,
			Client: client,
		}
		s.Mode = store.ModeInCluster

		s.ServerCert = &certs.ServerCert{}

		return nil
	},
}

// Execute - execute the root command
func Execute() {
	err := rootCmd.Execute()
	internal.DieOnError(err)
}

func init() {
	viper.AutomaticEnv()
	viper.BindEnv("kubeconfig", "KUBECONFIG")
	viper.BindEnv("cfconfig", "CFCONFIG")

	rootCmd.PersistentFlags().String("cfconfig", viper.GetString("cfconfig"), "Config file (default is $HOME/.cfconfig) [$CFCONFIG]")
	rootCmd.PersistentFlags().String("context", "", "Name of the context from --cfconfig (default is current-context)")
	rootCmd.PersistentFlags().String("kube-context-name", "", "Name of the kubernetes context (default is current-context)")
	rootCmd.PersistentFlags().String("kube-config-path", viper.GetString("kubeconfig"), "Path to kubeconfig file (default is $HOME/.kube/config) [$KUBECONFIG]")
	rootCmd.PersistentFlags().String("kube-namespace", "default", "Name of the namespace on which venona should be installed")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Print logs")
	rootCmd.PersistentFlags().BoolVar(&skipVerionCheck, "skip-version-check", false, "Do not compare current Venona's version with latest")

}
