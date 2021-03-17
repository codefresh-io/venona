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

package monitoring

import (
	"context"
	"net/http"

	gorillamux "github.com/gorilla/mux"
)

// Monitor controls monitoring for the application
type Monitor interface {
	NewTransaction(name string) Transaction
	NewTransactionFromContext(ctx context.Context) Transaction
	NewRoundTripper(rt http.RoundTripper) http.RoundTripper
	NewGorillaMiddleware() gorillamux.MiddlewareFunc
}

// Transaction instruments one logical unit of work: either an inbound web request
// or background task. Start a new Transaction with the Monitor.NewTransaction() method
type Transaction interface {
	// End finishes the Transaction.  After that, subsequent calls to End or
	// other Transaction methods have no effect.  All segments and
	// instrumentation must be completed before End is called.
	End()

	// AddAttribute adds a key value pair to the transaction event, errors,
	// and traces.
	//
	// The key must contain fewer than than 255 bytes.  The value must be a
	// number, string, or boolean.
	AddAttribute(key string, value interface{})

	NewRoundTripper(rt http.RoundTripper) http.RoundTripper

	NewSegment(r *http.Request) Segment

	NewSegmentByName(name string) Segment

	NoticeError(err error)
}

// Segment is used to instrument functions, methods, and blocks of code
type Segment interface {
	End()
}

// Empty implementation
type monitor struct{}
type transaction struct{}
type segment struct{}

// NewEmpty a noop monitor implementation
func NewEmpty() Monitor {
	return &monitor{}
}

// Monitor
func (m *monitor) NewTransaction(name string) Transaction {
	return &transaction{}
}

func (m *monitor) NewTransactionFromContext(ctx context.Context) Transaction {
	return &transaction{}
}

func (m *monitor) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return rt
}

func (m *monitor) NewGorillaMiddleware() gorillamux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return h
	}
}

// Transaction
func (t *transaction) NewSegment(r *http.Request) Segment {
	return &segment{}
}

func (t *transaction) NewSegmentByName(name string) Segment {
	return &segment{}
}

func (t *transaction) AddAttribute(key string, val interface{}) {}

func (t *transaction) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return rt
}

func (t *transaction) End() {}

func (t *transaction) NoticeError(err error) {}

// Segment
func (s *segment) End() {}
