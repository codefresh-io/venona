package config

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

type (
	mockLoggerOpts struct {
		method  string
		args    []interface{}
		returns []interface{}
	}
)

func mockLogger(opt ...mockLoggerOpts) *mocks.Logger {
	m := &mocks.Logger{}
	for _, o := range opt {
		m.On(o.method, o.args...).Return(o.returns...)
	}
	return m
}

func TestLoad(t *testing.T) {
	type args struct {
		dir     string
		pattern string
		logger  logger.Logger
	}
	tests := []struct {
		name         string
		args         args
		want         map[string]Config
		wantErr      bool
		fileReadFunc func(string) ([]byte, error)
		walkFileFunc func(string, filepath.WalkFunc) error
	}{
		{
			name: "Success and return empty list when file name does not match",
			args: args{
				dir: "location",
				logger: mockLogger(
					mockLoggerOpts{
						method: "Debug",
						args: []interface{}{
							"File ignored, regexp does not match",
							"regexp",
							"some-pattern",
							"file",
							"file",
						},
					},
				),
				pattern: "some-pattern",
			},
			wantErr: false,
			want:    map[string]Config{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				readfile = ioutil.ReadFile
				walkFilePath = filepath.Walk
			}()
			readfile = tt.fileReadFunc
			walkFilePath = tt.walkFileFunc
			got, err := Load(tt.args.dir, tt.args.pattern, tt.args.logger)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
