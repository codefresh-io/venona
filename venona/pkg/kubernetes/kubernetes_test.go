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
		name    string
		args    args
		want    kube
		wantErr bool
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
			wantErr: true,
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
