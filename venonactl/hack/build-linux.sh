#!/bin/bash
set -e
OUTFILE=$PWD/venonactl-linux
go generate ${PWD}/hack/generate.go
GOOS=linux GOARCH=386 go build -ldflags '-X github.com/codefresh-io/venona/venonactl/cmd.localDevFlow=true' -o $OUTFILE main.go

chmod +x $OUTFILE