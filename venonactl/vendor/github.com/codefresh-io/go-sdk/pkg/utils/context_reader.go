package utils

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type (
	CFContext struct {
		Type        string `yaml:"type"`
		Name        string `yaml:"name"`
		URL         string `yaml:"url"`
		Token       string `yaml:"token"`
		Beta        bool   `yaml:"beta"`
		OnPrem      bool   `yaml:"onPrem"`
		ACLType     string `yaml:"acl-type"`
		UserID      string `yaml:"user-id"`
		AccountID   string `yaml:"account-id"`
		Expires     int    `yaml:"expires"`
		UserName    string `yaml:"user-name"`
		AccountName string `yaml:"account-name"`
	}

	CFConfig struct {
		Contexts       map[string]*CFContext `yaml:"contexts"`
		CurrentContext string                `yaml:"current-context"`
	}
)

func ReadAuthContext(path string, name string) (*CFContext, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file\n")
		return nil, err
	}
	config := CFConfig{}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		fmt.Printf("Error unmarshaling content\n")
		fmt.Println(err.Error())
		return nil, err
	}
	var context *CFContext
	if name != "" {
		context = config.Contexts[name]
	} else {
		context = config.Contexts[config.CurrentContext]
	}
	return context, nil
}
