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

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		opt Options
	}
	tests := []struct {
		name        string
		args        args
		want        kube
		wantErr     bool
		errorString string
	}{
		{
			name: "on valid input retun kube",
			args: args{
				opt: Options{
					Type: "runtime",
				},
			},
			want:    kube{},
			wantErr: false,
		},
		{
			name: "on non valid type return errNotValidType",
			args: args{
				opt: Options{
					Type: "secret",
				},
			},
			wantErr:     true,
			errorString: "not a valid type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.opt)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errorString)
			}
		})
	}
}
