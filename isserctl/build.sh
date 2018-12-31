#!/bin/bash
set -e
DIR=$(dirname $0)
OUTFILE=${GOPATH}/bin/isserctl

dep ensure --vendor-only
go generate pkg/runtimectl/types.go
go build -o $OUTFILE ${DIR}/cmd

chmod +x $OUTFILE