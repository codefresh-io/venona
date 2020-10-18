package monitoring

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Monitor controlls monitoring for the application
type Monitor interface {
	NewTransaction(name string, w http.ResponseWriter, r *http.Request) Transaction
	NewTransactionFromContext(ctx context.Context) Transaction
	NewRoundTripper(rt http.RoundTripper) http.RoundTripper
	NewGinMiddleware() gin.HandlerFunc
}

// Transaction instruments one logical unit of work: either an inbound web request
// or background task. Start a new Transaction with the Monitor.NewTransaction() method
type Transaction interface {
	// End finishes the Transaction.  After that, subsequent calls to End or
	// other Transaction methods have no effect.  All segments and
	// instrumentation must be completed before End is called.
	End() error

	// AddAttribute adds a key value pair to the transaction event, errors,
	// and traces.
	//
	// The key must contain fewer than than 255 bytes.  The value must be a
	// number, string, or boolean.
	AddAttribute(key string, value interface{}) error

	NewRoundTripper(rt http.RoundTripper) http.RoundTripper

	NewSegment(r *http.Request) Segment

	NewSegmentByName(name string) Segment

	NoticeError(err error) error
}

// Segment is used to instrument functions, methods, and blocks of code
type Segment interface {
	End() error
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
func (m *monitor) NewTransaction(name string, w http.ResponseWriter, r *http.Request) Transaction {
	return &transaction{}
}

func (m *monitor) NewTransactionFromContext(ctx context.Context) Transaction {
	return &transaction{}
}

func (m *monitor) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return rt
}

func (*monitor) NewGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// Transaction
func (t *transaction) NewSegment(r *http.Request) Segment {
	return &segment{}
}

func (t *transaction) NewSegmentByName(name string) Segment {
	return &segment{}
}

func (t *transaction) AddAttribute(key string, val interface{}) error {
	return nil
}

func (t *transaction) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return rt
}

func (t *transaction) End() error {
	return nil
}

func (t *transaction) NoticeError(err error) error {
	return nil
}

// Segment
func (s *segment) End() error {
	return nil
}
