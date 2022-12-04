package plugins

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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

func (u *runtimeAttachPlugin) buildRuntimeConfig(ctx context.Context, opt *InstallOptions, v Values) (RuntimeConfiguration, error) {
	config, err := opt.KubeBuilder.BuildConfig()
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to get client config on runtime cluster: %v", err)
	}

	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to create client on runtime cluster: %v", err)
	}
	err = opt.KubeBuilder.EnsureNamespaceExists(ctx, cs)
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to ensure namespace on runtime cluster: %v", err)
	}

	secret, err := u.generateServiceAccountSecret(ctx, cs, opt.RuntimeNamespace, opt.RuntimeServiceAccount)
	if err != nil {
		return RuntimeConfiguration{}, fmt.Errorf("Failed to get secret from service account %s on runtime cluster: %v",
			opt.RuntimeServiceAccount, err)
	}

	crt := secret.Data["ca.crt"]
	token := secret.Data["token"]

	host := config.Host
	if opt.ClusterHost != "" {
		host = opt.ClusterHost
	}

	rc := RuntimeConfiguration{
		Crt:   string(crt),
		Token: string(token),
		Host:  host,
		Name:  opt.RuntimeEnvironment,
		Type:  "runtime",
	}

	return rc, nil
}

func (u *runtimeAttachPlugin) generateServiceAccountSecret(ctx context.Context, client kubernetes.Interface, namespace, saName string) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-token-", saName),
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": saName,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
	}

	u.logger.Debug("Creating secret for service-account token", "service-account", saName)

	secret, err := client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create service-account token secret: %w", err)
	}
	secretName := secret.Name

	u.logger.Debug("Created secret for service-account token", "service-account", saName, "secret", secret.Name)

	patch := []byte(fmt.Sprintf("{\"secrets\": [{\"name\": \"%s\"}]}", secretName))
	_, err = client.CoreV1().ServiceAccounts(namespace).Patch(ctx, saName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to patch service-account with new secret: %w", err)
	}

	u.logger.Debug("Added secret to service-account secrets", "service-account", saName, "secret", secret.Name)

	// try to read the token from the secret
	ticker := time.NewTicker(time.Second)
	retries := 15
	defer ticker.Stop()

	for try := 0; try < retries; try++ {
		select {
		case <-ticker.C:
			secret, err = client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		u.logger.Debug("Checking secret for service-account token", "service-account", saName, "secret", secret.Name)

		if err != nil {
			return nil, fmt.Errorf("failed to get service-account secret: %w", err)
		}

		if secret.Data == nil || len(secret.Data["token"]) == 0 {
			u.logger.Debug("Secret is missing service-account token", "service-account", saName, "secret", secret.Name)
			continue
		}

		u.logger.Debug("Got service-account token from secret", "service-account", saName, "secret", secret.Name)

		return secret, nil
	}

	return nil, fmt.Errorf("timed out waiting for secret to contain token")
}

func readCurrentVenonaConf(ctx context.Context, agentKubeBuilder KubeClientBuilder, clusterNamespace string) (venonaConf, error) {

	cs, err := agentKubeBuilder.BuildClient()
	if err != nil {
		return venonaConf{}, fmt.Errorf("Failed to create client on venona cluster: %v", err)
	}
	secret, err := cs.CoreV1().Secrets(clusterNamespace).Get(ctx, runtimeSecretName, metav1.GetOptions{})
	if err != nil {
		return venonaConf{}, fmt.Errorf("Failed to get %s secret: %v", runtimeSecretName, err)
	}

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

func (u *runtimeAttachPlugin) Install(ctx context.Context, opt *InstallOptions, v Values) (Values, error) {
	if opt.DryRun {
		return v, nil
	}
	cs, err := opt.AgentKubeBuilder.BuildClient() // on the agent cluster
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	err = opt.AgentKubeBuilder.EnsureNamespaceExists(ctx, cs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot ensure namespace exists: %v", err))
		return nil, err
	}

	// read current venona conf
	currentVenonaConf, err := readCurrentVenonaConf(ctx, opt.AgentKubeBuilder, opt.ClusterNamespace)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot read runnerconf: %v ", err))
		return nil, err
	}

	// new runtime configuration
	rc, err := u.buildRuntimeConfig(ctx, opt, v)
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
	v["Namespace"] = opt.ClusterNamespace

	cs.CoreV1().Secrets(opt.ClusterNamespace).Delete(ctx, runtimeSecretName, metav1.DeleteOptions{})

	err = install(ctx, &installOptions{
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
		list, err := cs.CoreV1().Pods(opt.ClusterNamespace).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%v", v["AppName"])})
		if err != nil {
			u.logger.Error(fmt.Sprintf("Cannot find agent pod: %v ", err))
			return nil, err
		}
		podName := list.Items[0].ObjectMeta.Name
		err = cs.CoreV1().Pods(opt.ClusterNamespace).Delete(ctx, podName, metav1.DeleteOptions{})
		if err != nil {
			u.logger.Error(fmt.Sprintf("Cannot delete agent pod: %v ", err))
			return nil, err
		}

		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				u.logger.Debug("Validating old runner pod termination")
				_, err = cs.CoreV1().Pods(opt.ClusterNamespace).Get(ctx, podName, metav1.GetOptions{})
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

func (u *runtimeAttachPlugin) Status(ctx context.Context, statusOpt *StatusOptions, v Values) ([][]string, error) {

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
	return status(ctx, opt)

}

func (u *runtimeAttachPlugin) Delete(ctx context.Context, deleteOpt *DeleteOptions, v Values) error {
	cs, err := deleteOpt.AgentKubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	// Delete the entry from runnerconf - if this is the only , delete the secret

	// read current venona conf
	currentVenonaConf, err := readCurrentVenonaConf(ctx, deleteOpt.AgentKubeBuilder, deleteOpt.AgentNamespace)
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

		cs.CoreV1().Secrets(deleteOpt.AgentNamespace).Delete(ctx, runtimeSecretName, metav1.DeleteOptions{})

		err = install(ctx, &installOptions{
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
		return uninstall(ctx, opt)
	}
	return nil

}

func (u *runtimeAttachPlugin) Upgrade(_ context.Context, _ *UpgradeOptions, v Values) (Values, error) {
	return v, nil
}

func (u *runtimeAttachPlugin) Migrate(context.Context, *MigrateOptions, Values) error {
	return fmt.Errorf("not supported")
}

func (u *runtimeAttachPlugin) Test(context.Context, *TestOptions, Values) error {
	return nil
}

func (u *runtimeAttachPlugin) Name() string {
	return RuntimeAttachType
}
