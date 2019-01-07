#!/bin/bash
set -e
DIR=$(dirname $0)
OUTFILE=${GOPATH}/bin/venonactl

go generate ${DIR}/pkg/operators/types.go
go build -o $OUTFILE main.go

chmod +x $OUTFILE