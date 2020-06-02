
#!/bin/bash

set -e

rm -rf .cover/ .test/
mkdir .cover/ .test/
trap "rm -rf .test/" EXIT

for pkg in `go list ./... | grep -v /vendor/`; do
    go test -race -v -covermode=atomic \
        -coverprofile=".cover/$(echo $pkg | sed 's/\//_/g').cover.out" $pkg
done
echo "mode: set" > .cover/cover.out && cat .cover/*.cover.out | grep -v mode: | sort -r | \
   awk '{if($1 != last) {print $0;last=$1}}' >> .cover/cover.out

code=$?
go tool cover -html=.cover/cover.out -o=.cover/coverage.html
echo "go test cmd exited with code $code"
exit $code