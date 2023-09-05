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
	"time"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/monitoring"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
)

var (
	errAlreadyRunning = errors.New("Server already running")
	errAlreadyStopped = errors.New("Server already stopped")
	errLoggerRequired = errors.New("Logger is required")
)

type (
	// Options for creating a new server instance
	Options struct {
		Port            string
		Logger          logger.Logger
		Monitor         monitoring.Monitor
		MetricsRegistry *prometheus.Registry
	}

	// Server is an HTTP server that expose API
	Server struct {
		log     logger.Logger
		running bool
		srv     *http.Server
	}
)

// New returns a new Server instance or an error
func New(opts *Options) (*Server, error) {
	if opts.Logger == nil {
		return nil, errLoggerRequired
	}
	log := opts.Logger

	r := mux.NewRouter()
	if opts.Monitor != nil {
		r.Use(opts.Monitor.NewGorillaMiddleware())
	}

	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Handle("/metrics", promhttp.HandlerFor(opts.MetricsRegistry, promhttp.HandlerOpts{Registry: opts.MetricsRegistry}))

	srv := &http.Server{
		Addr:              opts.Port,
		Handler:           r,
		ReadHeaderTimeout: 60 * time.Second,
	}

	return &Server{
		log,
		false,
		srv,
	}, nil
}

// Start starts the server and blocks indefinitely unless an error happens
func (s *Server) Start() error {
	if s.running {
		return errAlreadyRunning
	}

	s.running = true
	s.log.Info("Starting HTTP server", "addr", s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if !s.running {
		return errAlreadyStopped
	}

	s.running = false
	s.log.Warn("Received graceful termination request, shutting down...")
	err := s.srv.Shutdown(ctx)
	if err != nil {
		s.log.Error("failed to gracefully terminate server, cause: ", err)
	}

	return nil
}
