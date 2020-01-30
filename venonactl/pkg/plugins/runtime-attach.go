package plugins

import (
	"encoding/base64"
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"k8s.io/client-go/tools/clientcmd"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"

)

type runtimeAttachPlugin struct {
	logger logger.Logger
}

const (
	runtimeAttachFilesPattern = ".*.runtime-attach.yaml"
)

func saveKubeConfig(opt *InstallOptions, v Values) (Values, error) {
	config := opt.KubeBuilder.BuildConfig() // on the runtime cluster
	rc, err := config.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to get raw kubernetes config: %v", err)
	}
	d, err := clientcmd.Write(rc)
	if err != nil {
		return nil, fmt.Errorf("Failed to persist raw kubernetes config: %v", err)
	}

	v["RuntimeEnvironmentConfig"] = fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte(string(d))))
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

func (u *runtimeAttachPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error)  {

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
