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
	"time"
)

type (
	info struct {
		name    string
		size    int64
		mode    os.FileMode
		modTime time.Time
		isDir   bool
		sys     interface{}
	}
)

func (i info) Name() string {
	return i.name
}

func (i info) Size() int64 {
	return i.size
}

func (i info) Mode() os.FileMode {
	return i.mode
}

func (i info) ModTime() time.Time {
	return i.modTime
}

func (i info) IsDir() bool {
	return i.isDir
}

func (i info) Sys() interface{} {
	return i.sys
}
