package plugins

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	"gopkg.in/yaml.v2"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type runtimeAttachPlugin struct {
	logger logger.Logger
}

type RuntimeConfiguration struct {
	Crt   string `yaml:"crt"`
	Token string `yaml:"token"`
	Host  string `yaml:"host"`
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
}

type venonaConf struct {
	Runtimes map[string]RuntimeConfiguration `yaml:"runtimes,omitempty"`
}

const (
	runtimeAttachFilesPattern = ".*.runtime-attach.yaml"
	runtimeSecretName         = "runnerconf"
)

func buildRuntimeConfig(opt *InstallOptions, v Values) (RuntimeConfiguration, error) {

	config, err := opt.KubeBuilder.BuildConfig()
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to get client config on runtime cluster: %v", err)
	}

	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to create client on runtime cluster: %v", err)
	}
	err = opt.KubeBuilder.EnsureNamespaceExists(cs)
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to ensure namespace on runtime cluster: %v", err)
	}

	// get default service account for the namespace
	var getOpt metav1.GetOptions
	sa, err := cs.CoreV1().ServiceAccounts(opt.RuntimeClusterName).Get(opt.RuntimeServiceAccount, getOpt)
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to read service account runtime cluster: %v", err)
	}

	var saSecretName string
	saSecretPattern := fmt.Sprintf("%s-token-", opt.RuntimeServiceAccount)
	for _, secretRef := range sa.Secrets {
		if strings.Contains(secretRef.Name, saSecretPattern) {
			saSecretName = secretRef.Name
			break
		}
	}
	if saSecretName == "" {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to get secret %s from service account %s", saSecretPattern, opt.RuntimeServiceAccount)
	}
	secret, err := cs.CoreV1().Secrets(opt.RuntimeClusterName).Get(saSecretName, getOpt)
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to get secret from service account on runtime cluster: %v", err)
	}

	crt := secret.Data["ca.crt"]
	token := secret.Data["token"]

	rc := RuntimeConfiguration{
		Crt:   string(crt),
		Token: string(token),
		Host:  config.Host,
		Name:  opt.RuntimeEnvironment,
		Type:  "runtime",
	}

	return rc, nil
}

func readCurrentVenonaConf(agentKubeBuilder KubeClientBuilder, clusterNamespace string) (venonaConf, error) {

	cs, err := agentKubeBuilder.BuildClient()
	if err != nil {
		return venonaConf{}, fmt.Errorf("Failed to create client on venona cluster: %v", err)
	}
	secret, err := cs.CoreV1().Secrets(clusterNamespace).Get(runtimeSecretName, metav1.GetOptions{})

	conf := &venonaConf{
		Runtimes: make(map[string]RuntimeConfiguration),
	}
	for k, v := range secret.Data {
		cnf := RuntimeConfiguration{}
		if err := yaml.Unmarshal(v, &cnf); err != nil {
			return venonaConf{}, fmt.Errorf("Failed to unmarshal yaml with error: %s", err.Error())
		}
		conf.Runtimes[k] = cnf
	}
	return *conf, nil

}

func (u *runtimeAttachPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	cs, err := opt.AgentKubeBuilder.BuildClient() // on the agent cluster
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	err = opt.AgentKubeBuilder.EnsureNamespaceExists(cs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot ensure namespace exists: %v", err))
		return nil, err
	}

	// read current venona conf
	currentVenonaConf, err := readCurrentVenonaConf(opt.AgentKubeBuilder, opt.ClusterNamespace)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot read runnerconf: %v ", err))
		return nil, err
	}

	// new runtime configuration
	rc, err := buildRuntimeConfig(opt, v)
	if err != nil {
		return nil, err
	}
	if currentVenonaConf.Runtimes == nil {
		currentVenonaConf.Runtimes = make(map[string]RuntimeConfiguration)
	}
	// normalize the key in the secret to make sure we are not violating kube naming conventions
	name := strings.ReplaceAll(opt.RuntimeEnvironment, "/", ".")
	name = strings.ReplaceAll(name, "@", ".")
	name = strings.ReplaceAll(name, ":", ".")
	currentVenonaConf.Runtimes[fmt.Sprintf("%s.runtime.yaml", name)] = rc
	runtimes := map[string]string{}
	for name, runtime := range currentVenonaConf.Runtimes {
		// marshel prior persist
		d, err := yaml.Marshal(runtime)
		if err != nil {
			u.logger.Error(fmt.Sprintf("Cannot marshal merged runnerconf: %v ", err))
			return nil, err
		}

		runtimes[name] = base64.StdEncoding.EncodeToString([]byte(d))
	}
	v["runnerConf"] = runtimes

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

	if opt.RestartAgent {
		list, err := cs.CoreV1().Pods(opt.ClusterNamespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%v", v["AppName"])})
		if err != nil {
			u.logger.Error(fmt.Sprintf("Cannot find agent pod: %v ", err))
			return nil, err
		}
		podName := list.Items[0].ObjectMeta.Name
		err = cs.CoreV1().Pods(opt.ClusterNamespace).Delete(podName, &metav1.DeleteOptions{})
		if err != nil {
			u.logger.Error(fmt.Sprintf("Cannot delete agent pod: %v ", err))
			return nil, err
		}

		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				u.logger.Debug("Validating old runner pod termination")
				_, err = cs.CoreV1().Pods(opt.ClusterNamespace).Get(podName, metav1.GetOptions{})
				if err != nil {
					if statusError, errIsStatusError := err.(*kerrors.StatusError); errIsStatusError {
						if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
							return v, nil
						}
					}
				}
			case <-time.After(60 * time.Second):
				u.logger.Error("Failed to validate old venona pod termination")
				return v, fmt.Errorf("Failed to validate old venona pod termination")
			}
		}
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
	cs, err := deleteOpt.AgentKubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	// Delete the entry from runnerconf - if this is the only , delete the secret

	// read current venona conf
	currentVenonaConf, err := readCurrentVenonaConf(deleteOpt.AgentKubeBuilder, deleteOpt.AgentNamespace)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot read runnerconf: %v ", err))
		return err
	}
	name := strings.ReplaceAll(deleteOpt.RuntimeEnvironment, "/", ".")
	name = fmt.Sprintf("%s.runtime.yaml", name)
	if _, ok := currentVenonaConf.Runtimes[name]; ok {
		delete(currentVenonaConf.Runtimes, name)
	}

	// If only one runtime is defined, remove the secret , otherwise remove the entry and persist
	shouldDelete := true
	if len(currentVenonaConf.Runtimes) > 0 {

		runtimes := map[string]string{}
		for name, runtime := range currentVenonaConf.Runtimes {
			// marshel prior persist
			d, err := yaml.Marshal(runtime)
			if err != nil {
				u.logger.Error(fmt.Sprintf("Cannot marshal merged runnerconf: %v ", err))
				return err
			}

			runtimes[name] = base64.StdEncoding.EncodeToString([]byte(d))
		}

		shouldDelete = false
		v["runnerConf"] = runtimes

		cs.CoreV1().Secrets(deleteOpt.AgentNamespace).Delete(runtimeSecretName, &metav1.DeleteOptions{})

		err = install(&installOptions{
			logger:         u.logger,
			templates:      templates.TemplatesMap(),
			templateValues: v,
			kubeClientSet:  cs,
			namespace:      deleteOpt.AgentNamespace,
			matchPattern:   runtimeAttachFilesPattern,
			operatorType:   RuntimeAttachType,
		})
		return err

	}

	if shouldDelete {
		opt := &deleteOptions{
			templates:      templates.TemplatesMap(),
			templateValues: v,
			kubeClientSet:  cs,
			namespace:      deleteOpt.AgentNamespace,
			matchPattern:   runtimeAttachFilesPattern,
			operatorType:   RuntimeAttachType,
			logger:         u.logger,
		}
		return uninstall(opt)
	}
	return nil

}

func (u *runtimeAttachPlugin) Upgrade(_ *UpgradeOptions, v Values) (Values, error) {
	return v, nil
}

func (u *runtimeAttachPlugin) Migrate(*MigrateOptions, Values) error {
	return fmt.Errorf("not supported")
}

func (u *runtimeAttachPlugin) Test(opt TestOptions) error {
	return nil
}

func (u *runtimeAttachPlugin) Name() string {
	return RuntimeAttachType
}
