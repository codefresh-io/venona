#!/bin/bash
set -eux
MYDIR=$(dirname $0)
REPO_ROOT="${MYDIR}/../../.."

echo "Update image digests"
docker run \
    -v "$REPO_ROOT:/venona" \
    -u $(id -u) \
    --rm \
    quay.io/codefresh/codefresh-shell:0.0.20 \
    /bin/bash /venona/scripts/update_values_with_digests.sh