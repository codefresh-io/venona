#!/usr/bin/env bash

MYDIR=$(dirname $0)
CHARTDIR="${MYDIR}/../charts/cf-runtime"

<<COMMENT
Script updates runtime images from system/root runtime on SAAS to codefresh/values.yaml
yq (https://github.com/mikefarah/yq/) version 4.23.1
COMMENT

set -ue

DEBUG=${DEBUG:-false}

if [ ${DEBUG} == "true" ]; then
    set -x
fi

msg() { echo -e "\e[32mINFO ---> $1\e[0m"; }
err() { echo -e "\e[31mERR ---> $1\e[0m" ; return 1; }

runtimeJson=$(mktemp)
codefresh get sys-re system/root --extend -o json > $runtimeJson

RUNTIME_IMAGES=(
    ENGINE_IMAGE
    DIND_IMAGE
    CONTAINER_LOGGER_IMAGE
    DOCKER_PUSHER_IMAGE
    DOCKER_TAG_PUSHER_IMAGE
    DOCKER_PULLER_IMAGE
    DOCKER_BUILDER_IMAGE
    GIT_CLONE_IMAGE
    COMPOSE_IMAGE
    KUBE_DEPLOY
    FS_OPS_IMAGE
    TEMPLATE_ENGINE
    PIPELINE_DEBUGGER_IMAGE
)

filename=$CHARTDIR/values.yaml

for k in ${RUNTIME_IMAGES[@]}; do
    image="$(jq -er .runtimeScheduler.envVars.$k $runtimeJson)"
    patch "$filename" <<< $(diff -U0 -w -b --ignore-blank-lines $filename <(yq eval ".runtime.engine.runtimeImages.\"$k\" = \"$image\"" $filename)) || true
done

msg "The list of updated runtime images:\n"
echo -e "\e[33m$(cat $CHARTDIR/values.yaml)\e[0m"