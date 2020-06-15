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

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/mocks"
	"github.com/codefresh-io/go/venona/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_handleSignals(t *testing.T) {
	type args struct {
		sec           chan server.Event
		log           logger.Logger
		signals       []os.Signal
		expectExit    bool
		expectedEvent server.Event
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"should start graceful termination on SIGTERM",
			args{
				createMockServer().EventsChan,
				createMockLogger(),
				[]os.Signal{syscall.SIGTERM},
				false,
				server.Shutdown,
			},
		},
		{
			"should start graceful termination on SIGINT",
			args{
				createMockServer().EventsChan,
				createMockLogger(),
				[]os.Signal{syscall.SIGINT},
				false,
				server.Shutdown,
			},
		},
		{
			"should call exit if got SIGINT more than once",
			args{
				createMockServer().EventsChan,
				createMockLogger(),
				[]os.Signal{syscall.SIGINT, syscall.SIGINT},
				true,
				-1,
			},
		},
	}

	var sigChan chan<- os.Signal
	readyChan := make(chan struct{})
	handleSignal = func(c chan<- os.Signal, sig ...os.Signal) {
		sigChan = c
		readyChan <- struct{}{}
	}

	exitCalled := make(chan bool)
	exit = func(code int) {
		exitCalled <- true
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			go handleSignals(ctx, tt.args.sec, tt.args.log)
			<-readyChan // wait for sigChan to be swapped

			for _, sig := range tt.args.signals {
				sigChan <- sig
			}

			if tt.args.expectExit {
				assert.Equal(t, tt.args.expectExit, <-exitCalled)
			}
			if tt.args.expectedEvent != -1 {
				assert.Equal(t, tt.args.expectedEvent, <-tt.args.sec)
			} else {
				for range tt.args.signals {
					<-tt.args.sec
				}
			}
			cancel() // stops the handleSignals goroutine
		})
	}

	handleSignal = signal.Notify
	exit = os.Exit
}

func createMockLogger() *mocks.Logger {
	l := &mocks.Logger{}
	l.On("Crit", mock.Anything)
	return l
}

func createMockServer() *server.Server {
	s := server.Server{
		EventsChan: make(chan server.Event),
	}
	return &s
}
