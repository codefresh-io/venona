package plugins

import (
	"encoding/base64"
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"sigs.k8s.io/yaml"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type runtimeAttachPlugin struct {
	logger logger.Logger
}

type runtimeConfiguration struct {
	Crt string `yaml:"crt"`
	Token string `yaml:"token"`
	Host string `yaml:"host"`
}

type venonaConf struct {
	Runtimes map[string] runtimeConfiguration `yaml:"runtimes,omitempty"`
}

const (
	runtimeAttachFilesPattern = ".*.runtime-attach.yaml"
	runtimeSecretName = "venonaruntimes"
)

func buildRuntimeConfig(opt *InstallOptions, v Values) (runtimeConfiguration, error) {
	
	config, err := opt.KubeBuilder.BuildConfig().ClientConfig()
	if err != nil {
		return runtimeConfiguration{}, fmt.Errorf("Failed to get client config on runtime cluster: %v", err)
	}
	
	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return runtimeConfiguration{}, fmt.Errorf("Failed to create client on runtime cluster: %v", err)
	}

	// get default service account for the namespace
	var getOpt metav1.GetOptions
	sa, err := cs.CoreV1().ServiceAccounts(opt.RuntimeClusterName).Get(opt.RuntimeServiceAccount, getOpt)
	if err != nil {
		return runtimeConfiguration{}, fmt.Errorf("Failed to read service account runtime cluster: %v", err)
	}

	secretRef := sa.Secrets[0]
	secret, err := cs.CoreV1().Secrets(opt.RuntimeClusterName).Get(secretRef.Name, getOpt)
	if err != nil {
		return runtimeConfiguration{}, fmt.Errorf("Failed to get secret from service account on runtime cluster: %v", err)
	}

	crt := secret.Data["ca.crt"]
	token := secret.Data["token"]

	rc := runtimeConfiguration{
		Crt: string(crt),
		Token: string(token),
		Host: config.Host,

	}


	return rc, nil
}

func readCurrentVenonaConf(opt *InstallOptions) (venonaConf, error) {

	cs, err := opt.AgentKubeBuilder.BuildClient()
	if err != nil {
		return venonaConf{}, fmt.Errorf("Failed to create client on venona cluster: %v", err)
	}
	secret, err := cs.CoreV1().Secrets(opt.ClusterNamespace).Get(runtimeSecretName, metav1.GetOptions{})
	
	conf := &venonaConf{}
	err = yaml.Unmarshal(secret.Data["venonaconf"], &conf)
	if err != nil {
		return venonaConf{}, fmt.Errorf("Failed to unmarshal secret  cluster: %v", err)
	}
	if conf == nil {
		return venonaConf{}, nil
	}
	return *conf, nil
	
	

}

func (u *runtimeAttachPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	cs, err := opt.AgentKubeBuilder.BuildClient() // on the agent cluster
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}

	// read current venona conf
	currentVenonaConf, err := readCurrentVenonaConf(opt)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot read venonaconf: %v ", err))
		return nil, err
	}

	// new runtime configuration
	rc, err := buildRuntimeConfig(opt, v)
	if err != nil {
		return nil, err
	}
	if currentVenonaConf.Runtimes == nil {
		currentVenonaConf.Runtimes = make(map[string]runtimeConfiguration)
	}
	currentVenonaConf.Runtimes[opt.RuntimeEnvironment] = rc

	// marshel prior persist
	d, err := yaml.Marshal(&currentVenonaConf)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot marshal merged venonaconf: %v ", err))
		return nil, err
	}


	v["venonaConf"] = base64.StdEncoding.EncodeToString([]byte(d))

	// TODO: High - make the secret deletation as a transaction (rename)

	cs.CoreV1().Secrets(opt.ClusterNamespace).Delete(runtimeSecretName, &metav1.DeleteOptions{})

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
