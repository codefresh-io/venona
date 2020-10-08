package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
	"github.com/stretchr/objx"
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

	// make sure the ingress is ready and has an IP
	ticker := time.NewTicker(5 * time.Second)
Loop:
	for {
		select {
		case <-ticker.C:
			u.logger.Debug("Checking for app-proxy-service")
			ingress, err := cs.NetworkingV1beta1().Ingresses(opt.ClusterNamespace).Get("app-proxy", v1.GetOptions{})
			if err == nil {
				ips := ingress.Status.LoadBalancer.Ingress
				if len(ips) > 0 {
					// ingress has an IP
					break Loop
				}
			}
		case <-time.After(600 * time.Second):
			u.logger.Error("Failed to get app-proxy ingress ip")
			return v, fmt.Errorf("Failed to get app-proxy ingress ip")
		}
	}

	host := objx.New(v["AppProxy"]).Get("Host").Str()
	pathPrefix := objx.New(v["AppProxy"]).Get("PathPrefix").Str()
	appProxyURL := fmt.Sprintf("https://%v%v", host, pathPrefix)
	u.logger.Info(fmt.Sprintf("\napp proxy is running at: %v", appProxyURL))

	// update IPC
	file := os.NewFile(3, "pipe")
	if file == nil {
		return v, nil
	}
	data := map[string]interface{}{
		"ingressIP": appProxyURL,
	}
	var jsonData []byte
	jsonData, err = json.Marshal(data)
	n, err := file.Write(jsonData)
	if err != nil {
		u.logger.Error("Failed to write to stream", err)
		return v, fmt.Errorf("Failed to write to stream")
	}
	u.logger.Debug(fmt.Sprintf("%v bytes were written to stream\n", n))
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
