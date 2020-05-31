package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/codefresh-io/venona/pkg/logger"
	"gopkg.in/yaml.v2"
)

type (
	// Config used to define the connectivity to remote clusters
	Config struct {
		Type  string `yaml:"type" json:"type"`
		Cert  string `yaml:"crt" json:"crt"`
		Token string `yaml:"token" json:"token"`
		Host  string `yaml:"host" json:"host"`
		Name  string `yaml:"name" json:"name"`
	}
)

// Load will read the dir and load all the matching files matchig to the config
// In case of conflict, the first matching is used
func Load(dir string, log logger.Logger) (map[string]Config, error) {
	regexp, err := regexp.Compile(".*.runtime.yaml")
	if err != nil {
		return nil, err
	}
	var files []string
	if err := filepath.Walk(dir, visit(&files, regexp, log)); err != nil {
		return nil, err
	}
	result := map[string]Config{}
	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			log.Error("Failed to read file content", "file", file, "err", err.Error())
			continue
		}
		cnf := Config{}
		if err := yaml.Unmarshal(b, &cnf); err != nil {
			log.Error("Failed to unmarshal file content into struct", "file", file, "err", err.Error())
			continue
		}
		result[file] = cnf
	}

	return result, nil
}

func visit(files *[]string, re *regexp.Regexp, log logger.Logger) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("Failed to visit", "path", path, "err", err.Error())
			return nil
		}
		if info.IsDir() {
			log.Debug("Directory ignored, Venona loading only files that are mached to regexp", "regexp", re.String(), "dir", info.Name())
			return nil
		}
		if !re.MatchString(info.Name()) {
			log.Debug("File ignored, regexp does not match", "regexp", re.String(), "file", info.Name())
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}
