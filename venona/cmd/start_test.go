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
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/mocks"
	"github.com/stretchr/testify/mock"
)

func Test_handleSignals(t *testing.T) {
	type args struct {
		log        logger.Logger
		fakeSigs   []os.Signal
		expectExit bool
		stopDelay  time.Duration
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
				false,
				time.Duration(0),
			},
		},
		{
			"should start graceful termination on SIGINT",
			args{
				createMockLogger(),
				[]os.Signal{syscall.SIGINT},
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
				time.Millisecond * 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare mocks
			readyChan := make(chan struct{})
			var sigChan chan<- os.Signal
			exitChan := make(chan struct{})
			handleSignal = func(c chan<- os.Signal, sig ...os.Signal) {
				sigChan = c
				readyChan <- struct{}{}
			}
			exit = func(code int) {
				exitChan <- struct{}{}
			}
			serverStopChan := make(chan struct{})
			serverStopFunc := func() error {
				serverStopChan <- struct{}{}
				return nil
			}
			agentStopChan := make(chan struct{})
			agentStopFunc := func() error {
				agentStopChan <- struct{}{}
				time.Sleep(tt.args.stopDelay) // delay the termination
				return nil
			}

			go handleSignals(serverStopFunc, agentStopFunc, tt.args.log)
			<-readyChan // mocks are all in place
			for _, sig := range tt.args.fakeSigs {
				sigChan <- sig
			}

			if tt.args.expectExit {
				<-exitChan
			} else {
				<-agentStopChan
				<-serverStopChan
			}
		})
	}

	// cleanup
	handleSignal = signal.Notify
	exit = os.Exit
}

func createMockStopFunc() func() error {
	return func() error {
		time.Sleep(time.Millisecond * 50)
		return nil
	}
}

func createMockLogger() *mocks.Logger {
	l := &mocks.Logger{}
	l.On("Warn", mock.Anything)
	return l
}
