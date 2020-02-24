# VENONA
[![Go Report Card](https://goreportcard.com/badge/github.com/codefresh-io/venona)](https://goreportcard.com/report/github.com/codefresh-io/venona)
[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/codefresh-inc/codefresh-io%2Fvenona%2Fvenona?type=cf-1)]( https://g.codefresh.io/public/accounts/codefresh-inc/pipelines/codefresh-io/venona/venona)

## Version 1.x.x
Version 1.0.0 is released now, read more about migration from older version [here](#Migration)  
We highly suggest to use [Codefresh official CLI](https://codefresh-io.github.io/cli/) to install the agent:
```bash
kubectl create namespace codefresh
codefresh install agent --kube-namespace codefresh --install-runtime
```

The last command will:  
1. Install the agent on the namespace `codefresh`
2. Install the runtime on the same namespace
3. Attach the runtime to the agent

It is still possible, for advanced users to install all manually, for example:
One process of Venona can manage multiple runtime environments
NOTE: Please make sure that the process where Venona is installed there is a network connection to the clusters where the runtimes will be installed
```bash
# 1. Create namespace for the agent: 
kubectl create namespace codefresh-agent

# 2. Install the agent on the namespace ( give your agent a unique):
# Print a token that the Venona process will be using.
codefresh create agent $NAME
codefresh install agent --token $TOKEN --kube-namespace codefresh-agent

# 3. Create namespace for the first runtime:
kubectl create namespace codefresh-runtime-1

# 4. Install the first runtime on the namespace
# 5. the runtime name is printed
codefresh install runtime --kube-namespace codefresh-runtime-1

# 6. Attach the first runtime to agent:
codefresh attach runtime --agent-name $AGENT_NAME --agent-kube-namespace codefresh-agent --runtime-name $RUNTIME_NAME --kube-namespace codefresh-runtime-1

# 7. Restart the venona pod in namespace `codefresh-agent`
kubectl delete pods $VENONA_POD

# 8. Create namespace for the second runtime
kubectl create namespace codefresh-runtime-2

# 9. Install the second runtime on the namespace
codefresh install runtime --kube-namespace codefresh-runtime-2

# 10. Attach the second runtime to agent and restart the Venoa pod automatically
codefresh attach runtime --agent-name $AGENT_NAME --agent-kube-namespace codefresh-agent --runtime-name $RUNTIME_NAME --runtime-kube-namespace codefresh-runtime-1 --restart-agent

```

## Migration
Migrating from Venona `< 1.x.x` to `> 1.x.x` is not done automatically, please use the [migration script](https://github.com/codefresh-io/venona/blob/master/scripts/migration.sh) to do that, check out which environment variables are required to run it.
```bash
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
# 3. Attach runtime to the new agent (downtime ends) - pass $CODEFRESH_RUNTIME_NAME - required
```


## Installation

### Prerequisite:
* [Kubernetes](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - Used to create resource in your K8S cluster
  * Kube Version > 1.10:
    * [Instuction](#Install-on-cluster-version-<-1.10) to install on cluster version < 1.10
  * Disk size 50GB per node
* [Codefresh](https://codefresh-io.github.io/cli/) - Used to create resource in Codefresh
  * Authenticated context exist under `$HOME/.cfconfig` or authenticate with [Codefesh CLI](https://codefresh-io.github.io/cli/getting-started/#authenticate)

### Install venona

* Download [venona's](https://github.com/codefresh-io/venona/releases) binary
  * With homebrew:
    * `brew tap codefresh-io/venona`
    * `brew install venona`


#### Install on cluster version < 1.10
* Make sure the `PersistentLocalVolumes` [feature gate](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/) is turned on
* Venona's agent is trying to load avaliables apis using api `/openapi/v2` endpoint
Add this endpoint to ClusterRole `system:discovery` under `rules[0].nonResourceURLs`
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:discovery
rules:
- nonResourceURLs:
  - ...other_resources
  - /openapi
  - /openapi/*
  verbs:
  - get
```

#### Install on GCP
  * Make sure your user has `Kubernetes Engine Cluster Admin` role in google console
  * Bind your user with cluster-admin kubernetes clusterrole
    > `kubectl create clusterrolebinding NAME --clusterrole cluster-admin --user YOUR_USER`

#### Kubernetes RBAC
Installation of Venona on Kubernetes cluster installing 2 groups of objects,
Each one has own RBAC needs and therefore, created roles(and cluster-roles)
The resource descriptors are avaliable [here](https://github.com/codefresh-io/venona/tree/master/venonactl/templates/kubernetes)
List of the resources that will be created
* Agent (grouped by `/.*.venona.yaml/`)
  * `service-account.re.yaml` - The service account that the Venona pod will use to create the resource on the runtime namespace(the resoucre installed on the runtime namespace)
  * `role.re.yaml` - Allow to GET, CREATE and DELETE pods and persistentvolumeclaims
  * `role-binding.re.yaml` - The agent is spinning up pods and pvc, this biniding binds `role.venona.yaml` to `service-account.venona.yaml`
  * `cluster-role-binding.venona.yaml` - The agent discovering K8S apis by calling to `openapi/v2`, this ClusterRoleBinding binds  bootstraped ClusterRole by Kubernetes `system:discovery` to `service-account.venona.yaml`. This role has only permissions to make a GET calls to non resources urls
* Runtime-environment (grouped by `/.*.re.yaml/`) Kubernetes controller that spins up all required resources to provide a good caching expirience during pipeline execution
  * `service-account.dind-volume-provisioner.re.yaml` - The service account that the controller will use
  * `cluster-role.dind-volume-provisioner.re.yaml` Defines all the permission needed for the controller to operate correctly
  * `cluster-role-binding.dind-volume-provisioner.yaml` - Binds the ClusterRole to `service-account.dind-volume-provisioner.re.yaml`
