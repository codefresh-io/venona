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

package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/gin-gonic/gin"
)

var (
	errAlreadyStarted = errors.New("Already started")
	errAlreadyStopped = errors.New("Already stopped")
)

type (
	// Event an event that can be sent to the server events channel
	Event int

	// Server is an HTTP server that expose API
	Server struct {
		Port    string
		Logger  logger.Logger
		EventsC chan Event
		started bool
		srv     *http.Server
	}
)

const (
	// Shutdown send this event through a server's event channel
	// to start graceful termination
	Shutdown Event = iota
)

// Start starts the server and blocks indefinitely unless an error happens
func (s Server) Start() error {
	if s.started {
		return errAlreadyStarted
	}
	s.started = true
	s.Logger.Debug("Starting HTTP server", "port", s.Port)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	s.srv = &http.Server{
		Addr:    s.Port,
		Handler: r,
	}

	go s.handleExternalEvents()

	return s.srv.ListenAndServe()
}

func (s Server) handleExternalEvents() {
	for {
		switch <-s.EventsC {
		case Shutdown:
			s.shutdown()
			return
		}
	}
}

func (s Server) shutdown() {
	ctx := context.Background()
	err := s.srv.Shutdown(ctx)
	if err != nil {
		s.Logger.Error("failed to gracefully terminate server, cause: ", err)
	}
	s.started = false
}
