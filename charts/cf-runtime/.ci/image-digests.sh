#!/bin/bash
set -eux
MYDIR=$(dirname $0)
REPO_ROOT="${MYDIR}/../../.."

echo "Update image digests"
docker run \
    -v "$REPO_ROOT:/venona" \
    -u $(id -u) \
    --rm \
    --entrypoint /bin/sh \
    regclient/regctl:v0.7.2-alpine \
    -c \
    cd /venona && ./scripts/update_values_with_digests.sh