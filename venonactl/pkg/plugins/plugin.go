package plugins

import (
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/obj/kubeobj"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	RuntimeEnvironmentPluginType  = "runtime-environment"
	VenonaPluginType              = "venona"
	MonitorAgentPluginType        = "monitor-agent"
	VolumeProvisionerPluginType   = "volume-provisioner"
	EnginePluginType              = "engine"
	DefaultStorageClassNamePrefix = "dind-local-volumes-runner"
	RuntimeAttachType             = "runtime-attach"
	AppProxyPluginType            = "app-proxy"
)

type (
	Plugin interface {
		Install(*InstallOptions, Values) (Values, error)
		Status(*StatusOptions, Values) ([][]string, error)
		Delete(*DeleteOptions, Values) error
		Upgrade(*UpgradeOptions, Values) (Values, error)
		Migrate(*MigrateOptions, Values) error
		Test(TestOptions) error
		Name() string
	}

	PluginBuilder interface {
		Add(string) PluginBuilder
		Get() []Plugin
	}

	KubeClientBuilder interface {
		BuildClient() (*kubernetes.Clientset, error)
	}

	pb struct {
		logger  logger.Logger
		plugins []Plugin
	}

	Values map[string]interface{}

	InstallOptions struct {
		CodefreshHost         string
		CodefreshToken        string
		ClusterName           string
		ClusterNamespace      string
		RegisterWithAgent     bool
		MarkAsDefault         bool
		StorageClass          string
		DockerRegistry        string
		IsDefaultStorageClass bool
		KubeBuilder           interface {
			BuildClient() (*kubernetes.Clientset, error)
			BuildConfig() (*rest.Config, error)
			EnsureNamespaceExists(cs *kubernetes.Clientset) error
		}
		AgentKubeBuilder interface {
			BuildClient() (*kubernetes.Clientset, error)
			EnsureNamespaceExists(cs *kubernetes.Clientset) error
		}
		DryRun                bool
		KubernetesRunnerType  bool
		BuildNodeSelector     map[string]string
		Annotations           map[string]string
		RuntimeEnvironment    string
		RuntimeClusterName    string
		RuntimeServiceAccount string
		RestartAgent          bool
		Insecure              bool
	}

	DeleteOptions struct {
		KubeBuilder interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		AgentKubeBuilder interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		ClusterNamespace   string // runtime
		AgentNamespace     string // agent
		RuntimeEnvironment string
		RestartAgent       bool
	}

	UpgradeOptions struct {
		ClusterName      string
		ClusterNamespace string
		Name             string
		KubeBuilder      interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
	}

	MigrateOptions struct {
		ClusterName      string
		ClusterNamespace string
		KubeBuilder      interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
	}

	TestOptions struct {
		KubeBuilder interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		ClusterNamespace string
	}

	StatusOptions struct {
		KubeBuilder interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		ClusterNamespace string
	}

	installOptions struct {
		templates      map[string]string
		templateValues map[string]interface{}
		kubeClientSet  *kubernetes.Clientset
		namespace      string
		matchPattern   string
		operatorType   string
		dryRun         bool
		kubeBuilder    interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		logger logger.Logger
	}

	statusOptions struct {
		templates      map[string]string
		templateValues map[string]interface{}
		kubeClientSet  *kubernetes.Clientset
		namespace      string
		matchPattern   string
		operatorType   string
		logger         logger.Logger
	}

	deleteOptions struct {
		templates      map[string]string
		templateValues map[string]interface{}
		kubeClientSet  *kubernetes.Clientset
		namespace      string
		matchPattern   string
		operatorType   string
		logger         logger.Logger
	}

	testOptions struct {
		logger      logger.Logger
		kubeBuilder interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		namespace         string
		validationRequest validationRequest
	}
)

func NewBuilder(logger logger.Logger) PluginBuilder {
	return &pb{
		logger:  logger,
		plugins: []Plugin{},
	}
}

func (p *pb) Add(name string) PluginBuilder {
	p.plugins = append(p.plugins, build(name, p.logger))
	return p
}

func (p *pb) Get() []Plugin {
	return p.plugins
}

func build(t string, logger logger.Logger) Plugin {
	if t == VenonaPluginType {
		return &venonaPlugin{
			logger: logger.New("installer", VenonaPluginType),
		}
	}

	if t == RuntimeEnvironmentPluginType {
		return &runtimeEnvironmentPlugin{
			logger: logger.New("installer", RuntimeEnvironmentPluginType),
		}
	}

	if t == VolumeProvisionerPluginType {
		return &volumeProvisionerPlugin{
			logger: logger.New("installer", VolumeProvisionerPluginType),
		}
	}

	if t == EnginePluginType {
		return &enginePlugin{
			logger: logger.New("installer", EnginePluginType),
		}
	}

	if t == RuntimeAttachType {
		return &runtimeAttachPlugin{
			logger: logger.New("installer", RuntimeAttachType),
		}
	}

	if t == MonitorAgentPluginType {
		return &monitorAgentPlugin{
			logger: logger.New("installer", MonitorAgentPluginType),
		}
	}

	if t == AppProxyPluginType {
		return &appProxyPlugin{
			logger: logger.New("installer", AppProxyPluginType),
		}
	}

	return nil
}

func install(opt *installOptions) error {

	kubeObjects, err := KubeObjectsFromTemplates(opt.templates, opt.templateValues, opt.matchPattern, opt.logger)
	if err != nil {
		return err
	}

	for fileName, obj := range kubeObjects {
		if opt.dryRun == true {
			opt.logger.Debug(fmt.Sprintf("%v", obj), "File-Name", fileName)
			continue
		}
		var createErr error
		var kind, name string
		name, kind, createErr = kubeobj.CreateObject(opt.kubeClientSet, obj, opt.namespace)

		if createErr == nil {
			opt.logger.Debug(fmt.Sprintf("%s \"%s\" created", kind, name))
		} else if statusError, errIsStatusError := createErr.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				opt.logger.Debug(fmt.Sprintf("%s \"%s\" already exists", kind, name))
			} else {
				opt.logger.Debug(fmt.Sprintf("%s \"%s\" failed: %v ", kind, name, statusError))
				return statusError
			}
		} else {
			opt.logger.Debug(fmt.Sprintf("%s \"%s\" failed: %v ", kind, name, createErr))
			return createErr
		}
	}

	return nil
}

func status(opt *statusOptions) ([][]string, error) {
	kubeObjects, err := KubeObjectsFromTemplates(opt.templates, opt.templateValues, opt.matchPattern, opt.logger)
	if err != nil {
		return nil, err
	}
	var getErr error
	var kind, name string
	var rows [][]string
	for _, obj := range kubeObjects {
		name, kind, getErr = kubeobj.CheckObject(opt.kubeClientSet, obj, opt.namespace)
		if getErr == nil {
			rows = append(rows, []string{kind, name, StatusInstalled})
		} else if statusError, errIsStatusError := getErr.(*errors.StatusError); errIsStatusError {
			rows = append(rows, []string{kind, name, StatusNotInstalled, statusError.ErrStatus.Message})
		} else {
			opt.logger.Debug(fmt.Sprintf("%s \"%s\" failed: %v ", kind, name, getErr))
			return nil, getErr
		}
	}
	return rows, nil
}

func uninstall(opt *deleteOptions) error {

	kubeObjects, err := KubeObjectsFromTemplates(opt.templates, opt.templateValues, opt.matchPattern, opt.logger)
	if err != nil {
		return err
	}
	var kind, name string
	var deleteError error
	for _, obj := range kubeObjects {
		kind, name, deleteError = kubeobj.DeleteObject(opt.kubeClientSet, obj, opt.namespace)
		if deleteError == nil {
			opt.logger.Debug(fmt.Sprintf("%s \"%s\" deleted", kind, name))
		} else if statusError, errIsStatusError := deleteError.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				opt.logger.Debug(fmt.Sprintf("%s \"%s\" already exist", kind, name))
			} else if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
				opt.logger.Debug(fmt.Sprintf("%s \"%s\" not found", kind, name))
			} else {
				opt.logger.Error(fmt.Sprintf("%s \"%s\" failed: %v ", kind, name, statusError))
				return statusError
			}
		} else {
			opt.logger.Error(fmt.Sprintf("%s \"%s\" failed: %v ", kind, name, deleteError))
			return deleteError
		}
	}
	return nil
}

func test(opt testOptions) error {
	lgr := opt.logger
	cs, err := opt.kubeBuilder.BuildClient()
	if err != nil {
		lgr.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	lgr.Debug("Running acceptance tests")
	res, err := ensureClusterRequirements(cs, opt.validationRequest, lgr)
	if err != nil {
		return err
	}
	return handleValidationResult(res, lgr)
}
