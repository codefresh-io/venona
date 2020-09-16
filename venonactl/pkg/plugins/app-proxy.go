package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type appProxyPlugin struct {
	logger logger.Logger
}

const (
	appProxyFilesPattern = ".*.app-proxy.yaml"
)

func (u *appProxyPlugin) Install(opt *InstallOptions, v Values) (Values, error) {

	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	err = opt.KubeBuilder.EnsureNamespaceExists(cs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot ensure namespace exists: %v", err))
		return nil, err
	}
	err = install(&installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   appProxyFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   AppProxyPluginType,
	})
	if err != nil {
		u.logger.Error(fmt.Sprintf("AppProxy installation failed: %v", err))
		return nil, err
	}
	// locating the serivce and get it's internal api

	var ingressIP string
	ticker := time.NewTicker(5 * time.Second)
Loop:
	for {
		select {
		case <-ticker.C:
			u.logger.Debug("Checking for app-proxy-service")
			service, err := cs.CoreV1().Services(opt.ClusterNamespace).Get("app-proxy-service", v1.GetOptions{})
			if err == nil {
				ips := service.Status.LoadBalancer.Ingress
				if len(ips) > 0 {
					ingressIP = ips[0].IP
					break Loop
				}
			}
		case <-time.After(600 * time.Second):
			u.logger.Error("Failed to get app-proxy-service internal ip")
			return v, fmt.Errorf("Failed to get app-proxy-service internal ip")
		}
	}
	// update IPC
	file := os.NewFile(3, "pipe")
	data := map[string]interface{}{
		"ingressIP": ingressIP,
	}
	var jsonData []byte
	jsonData, err = json.Marshal(data)
	n, err := file.Write(jsonData)
	if err != nil {
		u.logger.Error("Failed to write to stream", err)
		return v, fmt.Errorf("Failed to write to stream")
	}
	u.logger.Info(fmt.Sprintf("%s bytes were written to stream", n))
	fmt.Sprintf(" ip : %s", ingressIP)
	return v, err

}

func (u *appProxyPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
	return [][]string{}, nil
}

func (u *appProxyPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
	cs, err := deleteOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return err
	}
	opt := &deleteOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      deleteOpt.ClusterNamespace,
		matchPattern:   appProxyFilesPattern,
		operatorType:   AppProxyPluginType,
		logger:         u.logger,
	}
	return uninstall(opt)
}

func (u *appProxyPlugin) Upgrade(opt *UpgradeOptions, v Values) (Values, error) {
	return nil, nil
}
func (u *appProxyPlugin) Migrate(*MigrateOptions, Values) error {
	return fmt.Errorf("not supported")
}

func (u *appProxyPlugin) Test(opt TestOptions) error {
	return nil
}

func (u *appProxyPlugin) Name() string {
	return AppProxyPluginType
}
