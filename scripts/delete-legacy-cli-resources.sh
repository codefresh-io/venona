#!/bin/bash

NAMESPACE=${1}

if [ -z "$NAMESPACE" ]; then
    echo "Usage: $0 <namespace>"
    exit 1
fi

KINDS=(
    "deployment"
    "daemonset"
    "configmap"
    "secret"
    "serviceaccount"
    "role"
    "rolebinding"
    "clusterrole"
    "clusterrolebinding"
    "storageclass"
)

for kind in "${KINDS[@]}"; do
    echo "Deleting $kind resources in namespace: $NAMESPACE"
    kubectl delete "$kind" -n "$NAMESPACE" -l 'app in (runner, venona, dind-volume-provisioner, dind-lv-monitor, app-proxy)' --ignore-not-found
done

# Delete unlabeled resources
kubectl -n $NAMESPACE delete secret $(kubectl get sa runner -o json | jq -r '.secrets.[].name')
kubectl -n $NAMESPACE delete runnerconf
kubectl -n $NAMESPACE delete sa runner
kubectl -n $NAMESPACE delete role codefresh-engine runner
kubectl -n $NAMESPACE delete rolebinding codefresh-engine runner
