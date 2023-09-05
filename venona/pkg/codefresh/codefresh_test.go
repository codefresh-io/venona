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

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/mocks"

	"github.com/stretchr/testify/assert"
)

func buildFakeMock() *mocks.Logger {
	l := &mocks.Logger{}
	return l
}

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
	type fields struct {
		host       string
		token      string
		agentID    string
		logger     logger.Logger
		httpClient RequestDoer
	}
	tests := map[string]struct {
		fields  fields
		paths   []string
		want    *url.URL
		wantErr bool
	}{
		"Reject when parsing the URL faile": {
			fields: fields{
				host: "123://sdd",
			},
			wantErr: true,
		},
		"Append path to the host": {
			paths: []string{"123", "123"},
			fields: fields{
				host: "http://url",
			},
			wantErr: false,
			want:    mustURL("http://url/123/123"),
		},
		"Escape paths": {
			paths: []string{"docker:desktop/server"},
			fields: fields{
				host: "http://url",
			},
			wantErr: false,
			want:    mustURL("http://url/docker:desktop%2Fserver"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := cf{
				host:       tt.fields.host,
				token:      tt.fields.token,
				agentID:    tt.fields.agentID,
				logger:     tt.fields.logger,
				httpClient: tt.fields.httpClient,
			}
			url, err := c.prepareURL(tt.paths...)
			if tt.wantErr {
				assert.Error(t, err)
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.String(), url.String())
			}
		})
	}
}
