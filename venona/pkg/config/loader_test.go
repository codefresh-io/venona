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
	"os"
	"path/filepath"
	"testing"

	"github.com/codefresh-io/go/venona/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	type args struct {
		dir     string
		pattern string
	}
	tests := map[string]struct {
		args         args
		want         map[string]Config
		wantErr      string
		fileReadFunc func(string) ([]byte, error)
		walkFileFunc func(string, filepath.WalkFunc) error
	}{
		"Success and return empty list when file name does not match": {
			args: args{
				dir:     "location",
				pattern: "some-pattern",
			},
			want: map[string]Config{},
			walkFileFunc: func(root string, fn filepath.WalkFunc) error {
				return fn("some-path", &info{
					name:  "file",
					isDir: false,
				}, nil)
			},
			fileReadFunc: func(string) ([]byte, error) {
				return []byte{}, nil
			},
		},
		"return config map from matching file": {
			args: args{
				dir:     "location",
				pattern: ".*",
			},
			want: map[string]Config{
				"location/file.a.yaml": {},
			},
			walkFileFunc: func(root string, fn filepath.WalkFunc) error {
				return fn("location/file.a.yaml", &info{
					name:  "file.a.yaml",
					isDir: false,
				}, nil)
			},
			fileReadFunc: func(string) ([]byte, error) {
				return []byte{}, nil
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				readfile = os.ReadFile
				walkFilePath = filepath.Walk
			}()
			readfile = tt.fileReadFunc
			walkFilePath = tt.walkFileFunc
			got, err := Load(tt.args.dir, tt.args.pattern, logger.New(logger.Options{}))
			if err != nil || tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
