package plugins

import (
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/obj/kubeobj"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	RuntimeEnvironmentPluginType  = "runtime-environment"
	VenonaPluginType              = "venona"
	VolumeProvisionerPluginType   = "volume-provisioner"
	EnginePluginType              = "engine"
	DefaultStorageClassNamePrefix = "dind-local-volumes-venona"
	RuntimeAttachType			  = "runtime-attach"
)

type (
	Plugin interface {
		Install(*InstallOptions, Values) (Values, error)
		Status(*StatusOptions, Values) ([][]string, error)
		Delete(*DeleteOptions, Values) error
		Upgrade(*UpgradeOptions, Values) (Values, error)
	}

	PluginBuilder interface {
		Add(string) PluginBuilder
		Get() []Plugin
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
		IsDefaultStorageClass bool
		KubeBuilder           interface {
			BuildClient() (*kubernetes.Clientset, error)
			BuildConfig() clientcmd.ClientConfig
			
		}
		AgentKubeBuilder	  interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		DryRun               bool
		KubernetesRunnerType bool
		BuildNodeSelector    map[string]string
		Annotations          map[string]string
		RuntimeEnvironment   string
		RuntimeClusterName   string
		RuntimeServiceAccount string
		RestartAgent        bool
	}

	DeleteOptions struct {
		KubeBuilder interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		ClusterNamespace string
	}

	UpgradeOptions struct {
		CodefreshHost    string
		CodefreshToken   string
		ClusterName      string
		ClusterNamespace string
		Name             string
		KubeBuilder      interface {
			BuildClient() (*kubernetes.Clientset, error)
		}
		DryRun bool
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
			logger: logger.New("Plugin", VenonaPluginType),
		}
	}

	if t == RuntimeEnvironmentPluginType {
		return &runtimeEnvironmentPlugin{
			logger: logger.New("Plugin", RuntimeEnvironmentPluginType),
		}
	}

	if t == VolumeProvisionerPluginType {
		return &volumeProvisionerPlugin{
			logger: logger.New("Plugin", VolumeProvisionerPluginType),
		}
	}

	if t == EnginePluginType {
		return &enginePlugin{
			logger: logger.New("Plugin", EnginePluginType),
		}
	}

	if t == RuntimeAttachType {
		return &runtimeAttachPlugin{
			logger: logger.New("Plugin", RuntimeAttachType),
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

func delete(opt *deleteOptions) error {

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
