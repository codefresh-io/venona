#!/bin/bash

set -x

AGENT=${AGENT:-true}
API_HOST=${API_HOST:-""}
API_KEY=${API_KEY:-""}

(set +x; codefresh auth create-context --api-key $API_KEY --url $API_HOST)

if [[ "$AGENT" == "true" ]]; then
    patch_type="re"
else
    patch_type="sys-re"
fi

codefresh patch $patch_type -f /opt/codefresh/runtime.yaml

for runtime in /opt/codefresh/runtime.d/system/*.yaml; do
    if [[ -f $runtime ]]; then
        codefresh patch sys-re -f $runtime
        ACCOUNTS=$(yq $runtime '.accounts')
        RUNTIME_NAME_ENCODED=$(yq '.metadata.name' $runtime | jq -Rr @uri)
        if [[ -n $ACCOUNTS ]]; then
            PAYLOAD=$(echo $ACCOUNTS | jq '{accounts: .}')
                set +x
                curl -X PUT \
                    -H "Content-Type: application/json" \
                    -H "Authorization: $API_KEY" \
                    -d "$PAYLOAD" \
                    "$API_HOST/api/admin/runtime-environments/account/modify/$RUNTIME_NAME_ENCODED"
        else
            echo "No accounts to add"
        fi
    fi
done
