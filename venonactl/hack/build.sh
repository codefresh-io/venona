#!/bin/bash
set -e
OUTFILE=/usr/local/bin/venonactl
go generate ${PWD}/hack/generate.go
go fmt ${PWD}/pkg/obj/kubeobj/kubeobj.go
go fmt ${PWD}/pkg/templates/kubernetes/templates.go
go build -ldflags '-X github.com/codefresh-io/venona/venonactl/cmd.localDevFlow=true' -o $OUTFILE main.go

chmod +x $OUTFILE
