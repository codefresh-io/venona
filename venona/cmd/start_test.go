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

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_handleSignals(t *testing.T) {
	type args struct {
		log             logger.Logger
		fakeSigs        []os.Signal
		expectExit      bool
		expectForceExit bool
		stopDelay       time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"should start graceful termination on SIGTERM",
			args{
				createMockLogger(),
				[]os.Signal{syscall.SIGTERM},
				true,
				false,
				time.Duration(0),
			},
		},
		{
			"should start graceful termination on SIGINT",
			args{
				createMockLogger(),
				[]os.Signal{syscall.SIGINT},
				true,
				false,
				time.Duration(0),
			},
		},
		{
			"should do forced exit when received two SIGINT signals",
			args{
				createMockLogger(),
				[]os.Signal{syscall.SIGINT, syscall.SIGINT},
				true,
				true,
				time.Millisecond * 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare mocks
			var sigChan chan<- os.Signal
			forcedExit := false
			serverExit := make(chan struct{}, 1)
			agentExit := make(chan struct{}, 1)

			handleSignal = func(c chan<- os.Signal, _ ...os.Signal) {
				sigChan = c
			}

			serverStopFunc := func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					forcedExit = true
				case <-time.After(tt.args.stopDelay): // delay exit
				}

				serverExit <- struct{}{}
				return nil
			}
			agentStopFunc := func() error {
				time.Sleep(tt.args.stopDelay) // delay the termination

				agentExit <- struct{}{}
				return nil
			}

			ctx := context.Background()
			ctx = withSignals(ctx, serverStopFunc, agentStopFunc, tt.args.log)

			for _, sig := range tt.args.fakeSigs {
				sigChan <- sig
			}

			if tt.args.expectExit {
				// wait
				<-serverExit
				<-agentExit
			}

			<-time.After(time.Millisecond * 1000)
			select {
			case <-ctx.Done():
				assert.True(t, tt.args.expectExit)
				assert.Equal(t, tt.args.expectForceExit, forcedExit)
			default:
				assert.False(t, tt.args.expectExit)
				assert.False(t, tt.args.expectForceExit)
			}
		})
	}

	// cleanup
	handleSignal = signal.Notify
}

func createMockLogger() *mocks.Logger {
	l := &mocks.Logger{}
	l.On("Warn", mock.Anything)
	return l
}
