
#!/bin/bash

set -e

VERSION=$(cat VERSION)

echo "Building version $VERSION"
go build -ldflags "-X github.com/codefresh-io/go/venona/cmd.version=$VERSION" -o venona *.go