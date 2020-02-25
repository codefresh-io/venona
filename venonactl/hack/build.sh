#!/bin/bash
set -e
OUTFILE=/usr/local/bin/venonactl
go generate ${PWD}/hack/generate.go
go fmt ${PWD}/pkg/obj/kubeobj/kubeobj.go
go fmt ${PWD}/pkg/templates/kubernetes/templates.go
VERSION="$(cat VERSION)-$(git rev-parse --short HEAD)"
echo "Setting up version $VERSION"
go build -ldflags "-X github.com/codefresh-io/venona/venonactl/cmd.version=$VERSION" -o $OUTFILE main.go

chmod +x $OUTFILE
