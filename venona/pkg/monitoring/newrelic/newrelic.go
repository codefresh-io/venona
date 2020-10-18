package newrelic

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	nr "github.com/newrelic/go-agent"
	nrgin "github.com/newrelic/go-agent/_integrations/nrgin/v1"

	"github.com/codefresh-io/go/venona/pkg/monitoring"
)

type (
	monitor struct {
		app nr.Application
	}

	transaction struct {
		t nr.Transaction
	}

	segment struct {
		s *nr.Segment
	}

	externalSegment struct {
		s *nr.ExternalSegment
	}
)

// New creates a new newrelic monitor
func New(conf nr.Config) (monitoring.Monitor, error) {
	app, err := nr.NewApplication(conf)
	if err != nil {
		return nil, err
	}

	return &monitor{app}, nil
}

// Monitor
func (m *monitor) NewTransaction(name string, w http.ResponseWriter, r *http.Request) monitoring.Transaction {
	return &transaction{m.app.StartTransaction(name, w, r)}
}

func (m *monitor) NewTransactionFromContext(ctx context.Context) monitoring.Transaction {
	return &transaction{nr.FromContext(ctx)}
}

func (m *monitor) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return nr.NewRoundTripper(nil, rt)
}

func (m *monitor) NewGinMiddleware() gin.HandlerFunc {
	return nrgin.Middleware(m.app)
}

// Transaction
func (t *transaction) NewSegment(r *http.Request) monitoring.Segment {
	return &externalSegment{nr.StartExternalSegment(t.t, r)}
}

func (t *transaction) NewSegmentByName(name string) monitoring.Segment {
	return &segment{nr.StartSegment(t.t, name)}
}

func (t *transaction) AddAttribute(key string, val interface{}) error {
	return t.t.AddAttribute(key, val)
}

func (t *transaction) NewRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return nr.NewRoundTripper(t.t, rt)
}

func (t *transaction) End() error {
	return t.t.End()
}

func (t *transaction) NoticeError(err error) error {
	return t.t.NoticeError(err)
}

// Segment
func (s *segment) End() error {
	return s.s.End()
}

func (s *externalSegment) End() error {
	return s.s.End()
}
