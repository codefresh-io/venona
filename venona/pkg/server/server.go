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
	errAlreadyRunning = errors.New("Already running")
	errAlreadyStopped = errors.New("Already stopped")
	errLoggerRequired = errors.New("Logger is required")
)

type (
	// Event an event that can be sent to the server events channel
	Event int

	// Options for creating a new server instance
	Options struct {
		Port   string
		Logger logger.Logger
		Mode   string
	}

	// Server is an HTTP server that expose API
	Server struct {
		Logger     logger.Logger
		EventsChan chan Event
		running    bool
		srv        *http.Server
	}
)

const (
	// Shutdown send this event through a server's event channel
	// to start graceful termination
	Shutdown Event = iota
)

const (
	// Release mode
	Release = gin.ReleaseMode
	// Debug mode (more logs)
	Debug = gin.DebugMode
)

// New returns a new Server instance or an error
func New(opt *Options) (Server, error) {
	s := Server{}
	if opt.Logger == nil {
		return s, errLoggerRequired
	}

	s.Logger = opt.Logger
	s.EventsChan = make(chan Event)
	gin.SetMode(opt.Mode)
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	s.srv = &http.Server{
		Addr:    opt.Port,
		Handler: r,
	}

	return s, nil
}

// Start starts the server and blocks indefinitely unless an error happens
func (s Server) Start() error {
	if s.running {
		return errAlreadyStarted
	}
	s.running = true
	s.Logger.Info("Starting HTTP server", "addr", s.srv.Addr)

	go s.handleExternalEvents()

	return s.srv.ListenAndServe()
}

func (s Server) handleExternalEvents() {
	for {
		switch <-s.EventsChan {
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
	s.running = false
}
