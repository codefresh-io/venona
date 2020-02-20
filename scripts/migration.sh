# This script comes to migrate old versions of Venona installation ( version < 1.x.x ) to new version (version >= 1.0.0 )
# Please read carefully what the script does.
# There will be a "downtime" in terms of your builds targeted to this runtime environment
# Once the script is finished, all the builds during the downtime will start
# The script will:
# 1. Create new agent entity in Codefresh using Codefresh CLI - give it a name $CODEFRESH_AGENT_NAME, default is "codefresh"
# 2. Install the agent on you cluster pass variables:
#   a. $VENONA_KUBE_NAMESPACE - required 
#   b. $VENONA_KUBE_CONTEXT - default is current-context
#   c. $VENONA_KUBECONFIG_PATH - default is $HOME/.kube/config
# 3. Delete the old Venona pod - pass variables: (downtime starting here)
#   a. $RUNTIME_KUBE_NAMESPACE - required 
#   b. $RUNTIME_KUBE_CONTEXT - default is current-context
#   c. $RUNTIME_KUBECONFIG_PATH - default is $HOME/.kube/config
# 4. Attach runtime to the new agent (downtime ends) - pass $CODEFRESH_RUNTIME_NAME - required

set -e

echoAndRun() {
    info "Running command: $1"
    eval $1
}

info() { echo  "INFO [$(date)] ---> $1"; }
fatal() { echo  "ERROR [$(date)] ---> $1" ; exit 1; }


DEFAULT_KUBECONFIG="$HOME/.kube/config"


if [ -z "$CODEFRESH_AGENT_NAME" ]
then
    info "CODEFRESH_AGENT_NAME is not set, using default name: codefresh"
    CODEFRESH_AGENT_NAME="codefresh"
else
    info "CODEFRESH_AGENT_NAME is set to $CODEFRESH_AGENT_NAME"
fi

if [ -z "$CODEFRESH_RUNTIME_NAME" ]
then
    fatal "CODEFRESH_RUNTIME_NAME is not set, exiting"
fi

if [ -z "$VENONA_KUBE_NAMESPACE" ]
then
    fatal "VENONA_KUBE_NAMESPACE is not set, exiting"
fi

if [ -z "$VENONA_KUBECONFIG_PATH" ]
then
    info "VENONA_KUBECONFIG_PATH is not set, using \$KUBECONFIG if exist or $DEFAULT_KUBECONFIG"
    VENONA_KUBECONFIG_PATH=${KUBECONFIG:=$DEFAULT_KUBECONFIG}
    info "VENONA_KUBECONFIG_PATH=$VENONA_KUBECONFIG_PATH"
else
    info "VENONA_KUBECONFIG_PATH is set to $VENONA_KUBECONFIG_PATH"
fi
if [ -z "$VENONA_KUBE_CONTEXT" ]
then
    info "VENONA_KUBE_CONTEXT is not set, using current-context"
    VENONA_KUBE_CONTEXT=$(kubectl --kubeconfig $VENONA_KUBECONFIG_PATH config current-context)
    info "VENONA_KUBE_CONTEXT=$VENONA_KUBE_CONTEXT"
else
    info "VENONA_KUBE_CONTEXT is set to $VENONA_KUBE_CONTEXT"
fi


if [ -z "$RUNTIME_KUBE_NAMESPACE" ]
then
    fatal "RUNTIME_KUBE_NAMESPACE is not set, exiting"
    exit 1
fi

if [ -z "$RUNTIME_KUBECONFIG_PATH" ]
then
    info "RUNTIME_KUBECONFIG_PATH is not set, using \$KUBECONFIG if exist or $DEFAULT_KUBECONFIG"
    RUNTIME_KUBECONFIG_PATH=${KUBECONFIG:=$DEFAULT_KUBECONFIG}
    info "RUNTIME_KUBECONFIG_PATH=$RUNTIME_KUBECONFIG_PATH"
else
    info "RUNTIME_KUBECONFIG_PATH is set to $RUNTIME_KUBECONFIG_PATH"
fi

if [ -z "$RUNTIME_KUBE_CONTEXT" ]
then
    info "RUNTIME_KUBE_CONTEXT is not set, using current-context"
    RUNTIME_KUBE_CONTEXT=$(kubectl --kubeconfig $RUNTIME_KUBECONFIG_PATH config current-context)
    info "RUNTIME_KUBE_CONTEXT=$RUNTIME_KUBE_CONTEXT"
else
    info "RUNTIME_KUBE_CONTEXT is set to $RUNTIME_KUBE_CONTEXT"
fi

runtimekube="kubectl --kubeconfig $RUNTIME_KUBECONFIG_PATH --cluster $RUNTIME_KUBE_CONTEXT -n $RUNTIME_KUBE_NAMESPACE"
agentkube="kubectl --kubeconfig $RUNTIME_KUBECONFIG_PATH --cluster $RUNTIME_KUBE_CONTEXT -n $RUNTIME_KUBE_NAMESPACE"

info "Testing connection to runtime cluster"
runtimeTestCmd="$runtimekube get deploy venona"
echoAndRun "$runtimeTestCmd"

info "Crating new agent in Codefresh"
token=$(echoAndRun "codefresh create agent $CODEFRESH_AGENT_NAME" | awk 'FNR==3')

info "Creating new namespace $VENONA_KUBE_NAMESPACE"
createNsCmd="$agentkube create namespace $VENONA_KUBE_NAMESPACE"
echoAndRun "$createNsCmd" || true

info "Installing agent on namespace using token $token"
echoAndRun "codefresh install agent --token $token --kube-namespace $VENONA_KUBE_NAMESPACE --kube-context-name $VENONA_KUBE_CONTEXT --kube-config-path $VENONA_KUBECONFIG_PATH --verbose"

info "Deleting current Venona process"
echoAndRun "$runtimekube delete deploy venona"


info "Attaching old runtime to new agent"
echoAndRun "codefresh attach runtime --runtime-name $CODEFRESH_RUNTIME_NAME --agent-name $CODEFRESH_AGENT_NAME --runtime-kube-context-name $RUNTIME_KUBE_CONTEXT --agent-kube-context-name $VENONA_KUBE_CONTEXT --runtime-kube-namespace $RUNTIME_KUBE_NAMESPACE --agent-kube-namespace $VENONA_KUBE_NAMESPACE --agent-kube-config-path $VENONA_KUBECONFIG_PATH --runtime-kube-config-path $RUNTIME_KUBECONFIG_PATH --restart-agent --verbose"

pod=$(eval "$agentkube get pods -l app=venona -o go-template='{{range .items }}{{ .metadata.name}}{{end}}'")
echoAndRun "$agentkube wait --for=condition=Ready pod/$pod --timeout 60s"

info "Migration finished"
