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

	"github.com/stretchr/testify/assert"
)

func Test_handleSignals(t *testing.T) {
	tests := map[string]struct {
		fakeSigs        []os.Signal
		expectExit      bool
		expectForceExit bool
		stopDelay       time.Duration
	}{
		"should start graceful termination on SIGTERM": {
			fakeSigs:        []os.Signal{syscall.SIGTERM},
			expectExit:      true,
			expectForceExit: false,
			stopDelay:       time.Duration(0),
		},
		"should start graceful termination on SIGINT": {
			fakeSigs:        []os.Signal{syscall.SIGINT},
			expectExit:      true,
			expectForceExit: false,
			stopDelay:       time.Duration(0),
		},
		"should do forced exit when received two SIGINT signals": {
			fakeSigs:        []os.Signal{syscall.SIGINT, syscall.SIGINT},
			expectExit:      true,
			expectForceExit: true,
			stopDelay:       time.Millisecond * 100,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
				case <-time.After(tt.stopDelay): // delay exit
				}

				serverExit <- struct{}{}
				return nil
			}
			agentStopFunc := func() error {
				time.Sleep(tt.stopDelay) // delay the termination

				agentExit <- struct{}{}
				return nil
			}

			ctx := context.Background()
			ctx = withSignals(ctx, serverStopFunc, agentStopFunc, logger.New(logger.Options{}))

			for _, sig := range tt.fakeSigs {
				sigChan <- sig
			}

			if tt.expectExit {
				// wait
				<-serverExit
				<-agentExit
			}

			<-time.After(time.Millisecond * 1000)
			select {
			case <-ctx.Done():
				assert.True(t, tt.expectExit)
				assert.Equal(t, tt.expectForceExit, forcedExit)
			default:
				assert.False(t, tt.expectExit)
				assert.False(t, tt.expectForceExit)
			}
		})
	}

	// cleanup
	handleSignal = signal.Notify
}
