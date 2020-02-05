package plugins

import (
	"encoding/base64"
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type runtimeAttachPlugin struct {
	logger logger.Logger
}

const (
	runtimeAttachFilesPattern = ".*.runtime-attach.yaml"
)

func saveKubeConfig(opt *InstallOptions, v Values) (Values, error) {
	
	config, err := opt.KubeBuilder.BuildConfig().ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to get client config on runtime cluster: %v", err)
	}
	
	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return nil, fmt.Errorf("Failed to create client on runtime cluster: %v", err)
	}

	// get default service account for the namespace
	var getOpt metav1.GetOptions
	sa, err := cs.CoreV1().ServiceAccounts(opt.ClusterNamespace).Get("default", getOpt)

	secretRef := sa.Secrets[0]
	secret, err := cs.CoreV1().Secrets(opt.ClusterNamespace).Get(secretRef.Name, getOpt)
	if err != nil {
		return nil, fmt.Errorf("Failed to get secret from service account on runtime cluster: %v", err)
	}

	crt := secret.Data["ca.crt"]
	token := secret.Data["token"]

	v["RuntimeEnvironmentConfigCrt"] = fmt.Sprintf("%s", base64.StdEncoding.EncodeToString(crt))
	v["RuntimeEnvironmentConfigToken"] = fmt.Sprintf("%s", base64.StdEncoding.EncodeToString(token))
	v["RuntimeEnvrionmentConfigHost"] = fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte (config.Host)))
	return v, nil
}

func (u *runtimeAttachPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	cs, err := opt.AgentKubeBuilder.BuildClient() // on the agent cluster
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}

	v, err = saveKubeConfig(opt, v)
	if err != nil {
		return nil, err
	}

	v["RuntimeEnvironment"] = opt.RuntimeEnvironment

	err = install(&installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   runtimeAttachFilesPattern,
		operatorType:   RuntimeAttachType,
		dryRun:         opt.DryRun,
	})

	if err != nil {
		return nil, err
	}

	return v, nil

}

func (u *runtimeAttachPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {

	cs, err := statusOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	opt := &statusOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      statusOpt.ClusterNamespace,
		matchPattern:   runtimeAttachFilesPattern,
		operatorType:   RuntimeAttachType,
		logger:         u.logger,
	}
	return status(opt)

}

func (u *runtimeAttachPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
	cs, err := deleteOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil
	}
	opt := &deleteOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      deleteOpt.ClusterNamespace,
		matchPattern:   runtimeAttachFilesPattern,
		operatorType:   RuntimeAttachType,
		logger:         u.logger,
	}
	return delete(opt)
}

func (u *runtimeAttachPlugin) Upgrade(_ *UpgradeOptions, v Values) (Values, error) {
	return v, nil
}
