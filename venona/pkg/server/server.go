package server

import (
	"errors"

	"github.com/codefresh-io/go/venona/pkg/agent"
	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/gin-gonic/gin"
)

var (
	errAlreadyStarted = errors.New("Already started")
)

type (
	// Server is an HTTP server that expose API
	Server struct {
		Agent   agent.Agent
		Port    string
		Logger  logger.Logger
		started bool
	}
)

// Start starts the server
func (s Server) Start() error {
	if s.started {
		return errAlreadyStarted
	}
	r := gin.Default()
	s.Logger.Debug("Starting HTTP server", "port", s.Port)
	r.GET("/", s.status)
	return r.Run(s.Port)
}

func (s Server) status(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": s.Agent.Status(),
	})
}
