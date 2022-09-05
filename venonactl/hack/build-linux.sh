#!/bin/bash
set -e
DIR=$(realpath $(dirname $0)/..)
OUTFILE=${DIR}/venonactl-linux
go generate ${DIR}/hack/generate.go
go fmt ${DIR}/pkg/obj/kubeobj/kubeobj.go
go fmt ${DIR}/pkg/templates/kubernetes/templates.go

CGO_ENABLED=0 GOOS=linux  go build -gcflags=all="-N -l" -ldflags '-X github.com/codefresh-io/venona/venonactl/cmd.localDevFlow=true'  -o $OUTFILE ${DIR}

chmod +x $OUTFILE
