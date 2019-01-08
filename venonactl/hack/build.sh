#!/bin/bash
set -e
OUTFILE=${GOPATH}/bin/venonactl

go generate ${GOPATH}/src/github.com/codefresh-io/venona/venonactl/pkg/operators/types.go
go build -o $OUTFILE main.go

chmod +x $OUTFILE