#!/bin/bash

echo "-----"
echo "API_HOST:         ${API_HOST}"
echo "AGENT_NAME:       ${AGENT_NAME}"
echo "RUNTIME_NAME:     ${RUNTIME_NAME}"
echo "-----"

auth() {
  codefresh auth create-context --api-key ${USER_CODEFRESH_TOKEN} --url ${API_HOST}
}

remove_runtime() {
  codefresh delete re ${RUNTIME_NAME}
}

remove_agent() {
  codefresh delete agent ${AGENT_NAME}
}

auth
remove_runtime
remove_agent