#!/bin/bash
set -e
OUTFILE=$PWD/venonactl-linux
go generate ${PWD}/hack/generate.go
go fmt ${PWD}/pkg/obj/kubeobj/kubeobj.go
go fmt ${PWD}/pkg/templates/kubernetes/templates.go

GOOS=linux  go build -gcflags=all="-N -l" -ldflags '-X github.com/codefresh-io/venona/venonactl/cmd.localDevFlow=true'  -o $OUTFILE main.go

chmod +x $OUTFILE
