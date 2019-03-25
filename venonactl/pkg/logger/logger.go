package logger

import (
	log "github.com/inconshreveable/log15"
)

type (
	Logger interface {
		log.Logger
	}

	Options struct {
		Command   string
		Verbose   bool
		LogToFile string
	}
)

func New(o *Options) Logger {
	l := log.New(log.Ctx{
		"Command": o.Command,
	})
	handlers := []log.Handler{}
	lvl := log.LvlInfo
	if o.Verbose {
		lvl = log.LvlDebug
	}
	verboseHandler := log.LvlFilterHandler(lvl, log.StdoutHandler)
	handlers = append(handlers, verboseHandler)
	if o.LogToFile != "" {
		fileHandler := log.LvlFilterHandler(log.LvlDebug, log.Must.FileHandler(o.LogToFile, log.JsonFormat()))
		callerHandler := log.CallerFuncHandler(fileHandler)
		fileLineHandler := log.CallerFileHandler(callerHandler)
		handlers = append(handlers, fileLineHandler)

	}
	l.SetHandler(log.MultiHandler(handlers...))
	return l
}
