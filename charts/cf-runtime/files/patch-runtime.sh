#!/bin/bash

set -x

AGENT=${AGENT:-true}
API_HOST=${API_HOST:-""}
API_KEY=${API_KEY:-""}
ACCOUNTS=${ACCOUNTS:-""}
RUNTIME_NAME_ENCODED=${RUNTIME_NAME_ENCODED:-""}

codefresh auth create-context --api-key $API_KEY --url $API_HOST
cat /usr/share/extras/runtime.yaml

if [[ $AGENT == "true" ]]; then
    codefresh patch re -f /usr/share/extras/runtime.yaml
else
    codefresh patch sys-re -f /usr/share/extras/runtime.yaml
    if [[ -n $ACCOUNTS ]]; then
        PAYLOAD=$(echo $ACCOUNTS | jq '{accounts: .}')
        curl -X PUT \
            -H "Content-Type: application/json" \
            -H "Authorization: $API_KEY" \
            -d "$PAYLOAD" \
            "$API_HOST/api/admin/runtime-environments/account/modify/$RUNTIME_NAME_ENCODED"
    else
        echo "No accounts to add"
    fi
fi
