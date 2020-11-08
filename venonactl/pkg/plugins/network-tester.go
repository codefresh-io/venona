/*
Copyright 2019 The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugins

import (
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/logger"
)

type networkTesterPlugin struct {
	logger logger.Logger
}

// Install venona agent
func (u *networkTesterPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	return nil, fmt.Errorf("not supported")
}

// Status of runtimectl environment
func (u *networkTesterPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
	return nil, fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
	return fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Upgrade(opt *UpgradeOptions, v Values) (Values, error) {
	return v, fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Migrate(*MigrateOptions, Values) error {
	return fmt.Errorf("not supported")
}

func (u *networkTesterPlugin) Test(opt TestOptions) error {
	return nil
}

func (u *networkTesterPlugin) Name() string {
	return NetworkTesterPluginType
}
