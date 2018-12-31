#!/bin/bash
#
# 
set -e 

USAGE="
usage: $0 <version> [isserctl-path] 
"
VERSION=$1
if [[ -z "${VERSION}" ]]; then
  echo $USAGE
  exit 1
fi
ISSERCTL=${2:-${GOPATH}/bin}

S3BUCKET=s3://download.codefresh.io

NAME=isserctl
SYSTEM=$(uname -s)
PLATFORM=$(uname -m)

S3PATH=${S3BUCKET}/${NAME}/${VERSION}/${SYSTEM}/${PLATFORM}/isserctl

ISSERCTL=${2:-${GOPATH}/bin/isserctl}
if [[ ! -f "${ISSERCTL}" ]]; then
   echo "cannot find isserctl to push. build it before"
   echo $USAGE
   exit 1
fi

echo "Uploading $ISSERCTL to $S3PATH "
aws s3 cp $ISSERCTL $S3PATH 
