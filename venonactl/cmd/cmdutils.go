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
	runtimectl "github.com/codefresh-io/venona/venonactl/pkg/operators"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/sirupsen/logrus"
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

func buildBasicStore() {
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
		if !res {
			logrus.WithFields(logrus.Fields{
				"Local-Version":  s.Version.Current.Version,
				"Latest-Version": s.Version.Latest.Version,
			}).Info("New version is avaliable, please update")
		}
	}
}

func extendStoreWithCodefershClient() error {
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

		logrus.WithFields(logrus.Fields{
			"Context-Name":   context.Name,
			"Codefresh-Host": cfAPIHost,
		}).Debug("Using codefresh context")
	} else {
		logrus.Debug("Using creentials from environment variables")
	}

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

func extendStoreWithKubeClient() {
	s := store.GetStore()
	if kubeConfigPath == "" {
		currentUser, _ := user.Current()
		if currentUser != nil {
			kubeConfigPath = path.Join(currentUser.HomeDir, ".kube", "config")
			logrus.WithFields(logrus.Fields{
				"Kube-Config-Path": kubeConfigPath,
			}).Debug("Path to kubeconfig not set, using default")
		}
	}

	s.KubernetesAPI = &store.KubernetesAPI{
		ConfigPath: kubeConfigPath,
	}
}

func prepareLogger() {
	if verbose == true {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func isUsingDefaultStorageClass(sc string) bool {
	if sc == "" {
		return true
	}
	return strings.HasPrefix(sc, runtimectl.DefaultStorageClassNamePrefix)
}
