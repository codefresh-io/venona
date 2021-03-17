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

package newrelic

import (
	"context"
	"net/http"

	gorillamux "github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	nr "github.com/newrelic/go-agent/v3/newrelic"

	"github.com/codefresh-io/go/venona/pkg/monitoring"
)

type (
	monitor struct {
		app *nr.Application
	}

	transaction struct {
		t *nr.Transaction
	}

	segment struct {
		s *nr.Segment
	}

	externalSegment struct {
		s *nr.ExternalSegment
	}
)

// New creates a new newrelic monitor
func New(conf ...nr.ConfigOption) (monitoring.Monitor, error) {
	app, err := nr.NewApplication(conf...)
	if err != nil {
		return nil, err
	}

	return &monitor{app}, nil
}

// Monitor
func (m *monitor) NewTransaction(name string) monitoring.Transaction {
	return &transaction{m.app.StartTransaction(name)}
}

func (m *monitor) NewTransactionFromContext(ctx context.Context) monitoring.Transaction {
	return &transaction{nr.FromContext(ctx)}
}

func (m *monitor) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return nr.NewRoundTripper(rt)
}

func (m *monitor) NewGorillaMiddleware() gorillamux.MiddlewareFunc {
	return nrgorilla.Middleware(m.app)
}

// Transaction
func (t *transaction) NewSegment(r *http.Request) monitoring.Segment {
	return &externalSegment{nr.StartExternalSegment(t.t, r)}
}

func (t *transaction) NewSegmentByName(name string) monitoring.Segment {
	return &segment{nr.StartSegment(t.t, name)}
}

func (t *transaction) AddAttribute(key string, val interface{}) {
	t.t.AddAttribute(key, val)
}

func (t *transaction) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return t.NewRoundTripper(rt)
}

func (t *transaction) End() {
	t.t.End()
}

func (t *transaction) NoticeError(err error) {
	t.t.NoticeError(err)
}

// Segment
func (s *segment) End() {
	s.s.End()
}

func (s *externalSegment) End() {
	s.s.End()
}
