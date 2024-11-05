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
    quay.io/codefresh/codefresh-shell:0.0.20 \
    -c \
    cd /venona && ./scripts/update_values_with_digests.sh