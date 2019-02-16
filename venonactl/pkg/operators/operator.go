package operators

import (
	"regexp"

	"github.com/codefresh-io/venona/venonactl/pkg/obj/kubeobj"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	RuntimeEnvironmentOperatorType = "runtime-environment"
	VenonaOperatorType             = "venona"
	VolumeProvisionerOperatorType  = "volume-provisioner"
	DefaultStorageClassNamePrefix  = "dind-local-volumes-venona"
)

type (
	Operator interface {
		Install() error
		Status() ([][]string, error)
		Delete() error
		Upgrade() error
	}

	installOptions struct {
		templates      map[string]string
		templateValues map[string]interface{}
		kubeClientSet  *kubernetes.Clientset
		namespace      string
		matchPattern   string
		operatorType   string
		dryRun         bool
	}

	statusOptions struct {
		templates      map[string]string
		templateValues map[string]interface{}
		kubeClientSet  *kubernetes.Clientset
		namespace      string
		matchPattern   string
		operatorType   string
	}

	deleteOptions struct {
		templates      map[string]string
		templateValues map[string]interface{}
		kubeClientSet  *kubernetes.Clientset
		namespace      string
		matchPattern   string
		operatorType   string
	}
)

func GetOperator(t string) Operator {
	if t == VenonaOperatorType {
		return &venonaOperator{}
	}

	if t == RuntimeEnvironmentOperatorType {
		return &RuntimeEnvironmentOperator{}
	}

	if t == VolumeProvisionerOperatorType {
		return &VolumeProvisionerOperator{}
	}

	return nil
}

func install(opt *installOptions) error {

	kubeObjects, err := KubeObjectsFromTemplates(opt.templates, opt.templateValues)
	if err != nil {
		return err
	}

	for fileName, obj := range kubeObjects {
		match, _ := regexp.MatchString(opt.matchPattern, fileName)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Operator": opt.operatorType,
				"Pattern":  opt.matchPattern,
			}).Debugf("Skipping installation of %s: pattern not match", fileName)
			continue
		}
		if opt.dryRun == true {
			logrus.WithFields(logrus.Fields{
				"File-Name": fileName,
				"Operator":  opt.operatorType,
			}).Debugf("%v", obj)
			continue
		}
		var createErr error
		var kind, name string
		name, kind, createErr = kubeobj.CreateObject(opt.kubeClientSet, obj, opt.namespace)

		if createErr == nil {
			logrus.Debugf("%s \"%s\" created\n ", kind, name)
		} else if statusError, errIsStatusError := createErr.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logrus.Debugf("%s \"%s\" already exists\n", kind, name)
			} else {
				logrus.Debugf("%s \"%s\" failed: %v ", kind, name, statusError)
				return statusError
			}
		} else {
			logrus.Debugf("%s \"%s\" failed: %v ", kind, name, createErr)
			return createErr
		}
	}

	return nil
}

func status(opt *statusOptions) ([][]string, error) {
	kubeObjects, err := KubeObjectsFromTemplates(opt.templates, opt.templateValues)
	if err != nil {
		return nil, err
	}
	var getErr error
	var kind, name string
	var rows [][]string
	for fileName, obj := range kubeObjects {
		match, _ := regexp.MatchString(opt.operatorType, fileName)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Operator": opt.operatorType,
				"Pattern":  opt.matchPattern,
			}).Debugf("Skipping status check of %s: pattern not match", fileName)
			continue
		}
		name, kind, getErr = kubeobj.CheckObject(opt.kubeClientSet, obj, opt.namespace)
		if getErr == nil {
			rows = append(rows, []string{kind, name, StatusInstalled})
		} else if statusError, errIsStatusError := getErr.(*errors.StatusError); errIsStatusError {
			rows = append(rows, []string{kind, name, StatusNotInstalled, statusError.ErrStatus.Message})
		} else {
			logrus.Debugf("%s \"%s\" failed: %v ", kind, name, getErr)
			return nil, getErr
		}
	}
	return rows, nil
}

func delete(opt *deleteOptions) error {
	kubeObjects, err := KubeObjectsFromTemplates(opt.templates, opt.templateValues)
	if err != nil {
		return err
	}
	var kind, name string
	var deleteError error
	for fileName, obj := range kubeObjects {
		match, _ := regexp.MatchString(opt.matchPattern, fileName)
		if match != true {
			logrus.WithFields(logrus.Fields{
				"Operator": opt.operatorType,
				"Pattern":  opt.matchPattern,
			}).Debugf("Skipping deletion of %s: pattern not match", fileName)
			continue
		}
		kind, name, deleteError = kubeobj.DeleteObject(opt.kubeClientSet, obj, opt.namespace)
		if deleteError == nil {
			logrus.Debugf("%s \"%s\" deleted\n ", kind, name)
		} else if statusError, errIsStatusError := deleteError.(*errors.StatusError); errIsStatusError {
			if statusError.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
				logrus.Debugf("%s \"%s\" already exists\n", kind, name)
			} else if statusError.ErrStatus.Reason == metav1.StatusReasonNotFound {
				logrus.Debugf("%s \"%s\" not found\n", kind, name)
			} else {
				logrus.Errorf("%s \"%s\" failed: %v ", kind, name, statusError)
				return statusError
			}
		} else {
			logrus.Errorf("%s \"%s\" failed: %v ", kind, name, deleteError)
			return deleteError
		}
	}
	return nil
}
