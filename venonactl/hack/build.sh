#!/bin/bash
set -e
OUTFILE=${GOPATH}/bin/venonactl
go generate ${PWD}/hack/generate.go
go build -ldflags '-X github.com/codefresh-io/venona/venonactl/cmd.localDevFlow=true' -o $OUTFILE main.go

chmod +x $OUTFILE