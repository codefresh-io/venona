package logger

import (
	log "github.com/inconshreveable/log15"
)

type (
	Logger interface {
		log.Logger
	}

	Options struct {
		Verbose bool
	}
)

// New creates new logger
func New(o Options) Logger {
	l := log.New(log.Ctx{})
	handlers := []log.Handler{}
	lvl := log.LvlInfo
	if o.Verbose {
		lvl = log.LvlDebug
	}
	verboseHandler := log.LvlFilterHandler(lvl, log.StdoutHandler)
	handlers = append(handlers, verboseHandler)
	l.SetHandler(log.MultiHandler(handlers...))
	return l
}
