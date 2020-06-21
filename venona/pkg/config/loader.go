// Copyright 2020 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"gopkg.in/yaml.v2"
)

var (
	readfile     = ioutil.ReadFile
	walkFilePath = filepath.Walk
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

	// Options to load the config
	Options struct {
		Logger logger.Logger
		Dir    string
	}
)

// Load read the dir and load all the matching files matchig to the config
// In case of conflict, the first matching is used
func Load(dir string, pattern string, logger logger.Logger) (map[string]Config, error) {
	regexp, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var files []string
	if err := walkFilePath(dir, visit(&files, regexp, logger)); err != nil {
		return nil, err
	}
	return buildConfigMap(files, logger)
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

func unmarshalConfig(data []byte) (Config, error) {
	cnf := Config{}
	if err := yaml.Unmarshal(data, &cnf); err != nil {
		return cnf, err
	}
	return cnf, nil
}

func buildConfigMap(files []string, logger logger.Logger) (map[string]Config, error) {
	result := map[string]Config{}
	for _, file := range files {
		b, err := readfile(file)
		if err != nil {
			logger.Error("Failed to read file content", "file", file, "err", err.Error())
			continue
		}
		cnf, err := unmarshalConfig(b)
		if err != nil {
			logger.Error("Failed to unmarshal file content into struct", "file", file, "err", err.Error())
			continue

		}
		result[file] = cnf
	}
	return result, nil
}
