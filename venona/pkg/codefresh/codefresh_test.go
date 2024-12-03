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

package codefresh

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustURL(u string) *url.URL {
	r, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return r
}

func TestNew(t *testing.T) {
	tests := map[string]struct {
		opts Options
		want Codefresh
	}{
		"Build client with default host": {
			want: &cf{
				host:       defaultHost,
				httpClient: &http.Client{},
			},
		},
		"Build client with given host": {
			opts: Options{
				Host: "http://host.com",
			},
			want: &cf{
				host:       "http://host.com",
				httpClient: &http.Client{},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := New(tt.opts)
			assert.Equal(t, tt.want.Host(), got.Host())
		})
	}
}

func Test_cf_prepareURL(t *testing.T) {
	type args struct {
		host       string
		token      string
		agentID    string
		httpClient RequestDoer
	}
	tests := map[string]struct {
		fields  args
		paths   []string
		query   map[string]string
		want    *url.URL
		wantErr bool
	}{
		"Reject when parsing the URL faile": {
			fields: args{
				host: "123://sdd",
			},
			wantErr: true,
		},
		"Append path to the host": {
			paths: []string{"123", "123"},
			fields: args{
				host: "http://url",
			},
			wantErr: false,
			want:    mustURL("http://url/123/123"),
		},
		"Escape paths": {
			paths: []string{"docker:desktop/server"},
			fields: args{
				host: "http://url",
			},
			wantErr: false,
			want:    mustURL("http://url/docker:desktop%2Fserver"),
		},
		"Append query": {
			query: map[string]string{
				"key":    "value",
				"keyTwo": "valueTwo",
			},
			paths: []string{"docker:desktop/server"},
			fields: args{
				host: "http://url",
			},
			wantErr: false,
			want:    mustURL("http://url/docker:desktop%2Fserver?key=value&keyTwo=valueTwo"),
		},
		"Escape query": {
			query: map[string]string{
				"ke+y": "va+lu=e",
			},
			paths: []string{"docker:desktop/server"},
			fields: args{
				host: "http://url",
			},
			wantErr: false,
			want:    mustURL("http://url/docker:desktop%2Fserver?ke%2By=va%2Blu%3De"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := cf{
				host:       tt.fields.host,
				token:      tt.fields.token,
				agentID:    tt.fields.agentID,
				httpClient: tt.fields.httpClient,
			}
			url, err := c.prepareURL(tt.query, tt.paths...)
			if tt.wantErr {
				assert.Error(t, err)
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.String(), url.String())
			}
		})
	}
}
