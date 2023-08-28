#!/bin/bash

echo "-----"
echo "API_HOST:         ${API_HOST}"
echo "AGENT_NAME:       ${AGENT_NAME}"
echo "RUNTIME_NAME:     ${RUNTIME_NAME}"
echo "AGENT:            ${AGENT}"
echo "-----"

auth() {
  codefresh auth create-context --api-key ${API_TOKEN} --url ${API_HOST}
}

remove_runtime() {
  if [ "$AGENT" == "true" ]; then
    codefresh delete re ${RUNTIME_NAME} || true
  else
    codefresh delete sys-re ${RUNTIME_NAME} || true
  fi
}

remove_agent() {
  codefresh delete agent ${AGENT_NAME} || true
}

remove_finalizers() {
  kubectl patch secret $(kubectl get secret -l codefresh.io/internal=true | awk 'NR>1{print $1}' | xargs) -p '{"metadata":{"finalizers":null}}' --type=merge || true
}

auth
remove_runtime
remove_agent
remove_finalizers