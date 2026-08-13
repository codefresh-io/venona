#!/bin/bash
## Reference: https://github.com/norwoodj/helm-docs
set -eux
REPO_ROOT="${1:-$(cd "$(dirname "$0")/../../.." && pwd)}"
WORKDIR="${2:-/helm-docs}"

echo "$REPO_ROOT"
echo "$WORKDIR"

echo "Running Helm-Docs"
docker run \
    -v "$REPO_ROOT:$WORKDIR" \
    -w "$WORKDIR" \
    -u $(id -u) \
    --rm \
    --entrypoint /bin/sh \
    jnorwood/helm-docs:v1.9.1 \
    -c \
    "helm-docs --chart-search-root=charts --template-files=./_templates.gotmpl --template-files=README.md.gotmpl"
