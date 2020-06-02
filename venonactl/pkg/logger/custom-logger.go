package logger

import (
	"bytes"
	"fmt"
	"os"

	log "github.com/inconshreveable/log15"
)

const (
	Plain = "plain"
)

func PlainTextFormatter() log.Format {
	return log.FormatFunc(func(r *log.Record) []byte {
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "%s\n", r.Msg)
		return buf.Bytes()
	})
}

func getStdoutHanlder(o *Options) log.Handler {
	if o.LogFormatter == Plain {
		return log.StreamHandler(os.Stdout, PlainTextFormatter())
	}
	return log.StdoutHandler
}
