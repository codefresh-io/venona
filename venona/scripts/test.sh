
#!/bin/bash

set -e

rm -rf cover/
mkdir cover/

echo "running go test"
go test -v -race -coverprofile=cover/cover.out -covermode=atomic ./...
code=$?
echo "go test cmd exited with code $code"

echo "running go tool cover"
go tool cover -html=cover/cover.out -o=cover/coverage.html
echo "go tool cover exited with code $?"

exit $code
