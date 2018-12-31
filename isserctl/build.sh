#!/bin/bash
set -e
DIR=$(dirname $0)
OUTFILE=${GOPATH}/bin/isserctl

go generate ${DIR}/pkg/runtimectl/types.go
dep ensure --vendor-only
go build -o $OUTFILE ${DIR}/cmd

chmod +x $OUTFILE