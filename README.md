# VENONA
[![Go Report Card](https://goreportcard.com/badge/github.com/codefresh-io/venona)](https://goreportcard.com/report/github.com/codefresh-io/venona)
[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/codefresh-inc/codefresh-io%2Fvenona%2Fvenona?type=cf-1)]( https://g.codefresh.io/public/accounts/codefresh-inc/pipelines/codefresh-io/venona/venona)

## Version 1.x.x
Version 1.0.0 is released now, read more about migration from older version [here](#Migration)
We highly suggest to use Codefresh official CLI to install the agent:
1. `kubectl create namespace codefreh`
2. `codefresh install agent --kube-namespace codefresh --install-runtime`
The last command will:
1. Install the agent on the namespace `codefresh`
2. Install the runtime on the same namespace
3. Attach the runtime to the agent

It is still possible, for advanced users to install all manually, for example:
One process of Venona can manage multiple runtime environments
NOTE: Please make sure that the process where Venona is installed there is a network connection to the clusters where the runtimes will be installed
1. Create namespace for the agent: `kubectl create namespace codefresh-agent`
2. Install the agent on the namespace ( give your agent a unique):
  a. `codefresh create agent $NAME` - this command will print a token that the Venona process will be using.
  b. `codefresh install agent --token $TOKEN --kube-namespace codefresh-agent`
3. Create namespace for the first runtime: `kubectl create namespace codefresh-runtime-1`
4. Install the first runtime on the namespace: `codefresh install runtime --kube-namespace codefresh-runtime-1` - the runtime name is printted
5. Attach the first runtime to agent:
  a. `codefresh attach runtime --agent-name $AGENT_NAME --agent-kube-namespace codefresh-agent --runtime-name $RUNTIME_NAME --kube-namespace codefresh-runtime-1`
  b. restart the venona pod in namespace `codefresh-agent`
5. Create namespace for the second runtime: `kubectl create namespace codefresh-runtime-2`
6. Install the second runtime on the namespace: `codefresh install runtime --kube-namespace codefresh-runtime-2`
7. Attach the second runtime to agent:
  a. `codefresh attach runtime --agent-name $AGENT_NAME --agent-kube-namespace codefresh-agent --runtime-name $RUNTIME_NAME --runtime-kube-namespace codefresh-runtime-1`
  b. restart the venona pod in namespace `codefresh-agent`



  


## Installation

### Prerequisite:
* [Kubernetes](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - Used to create resource in your K8S cluster
  * Kube Version > 1.10:
    * [Instuction](#Install-on-cluster-version-<-1.10) to install on cluster version < 1.10
  * Disk size 50GB per node
* [Codefresh](https://codefresh-io.github.io/cli/) - Used to create resource in Codefresh
  * Authenticated context exist under `$HOME/.cfconfig` or authenticate with [Codefesh CLI](https://codefresh-io.github.io/cli/getting-started/#authenticate)

### Migration
Moving from Venona < 1.0.0 to > 1.0.0 is not done automatically atm, the fastest way to delete may cause "downtime", means that the pipeline that was configured to run on that runtime will not be able to execute.
1. Detele Venona `venona delete $NAME`
2. Install version 1.0.0 as described [here](#Version 1.x.x)

### Install venona

* Download [venona's](https://github.com/codefresh-io/venona/releases) binary
  * With homebrew:
    * `brew tap codefresh-io/venona`
    * `brew install venona`
* Create namespace where venona should run<br />
  > `kubectl create namespace codefresh-runtime`
* Create *new* runtime-environment with Venona's agents installed <br />
  > `venona install --kube-namespace codefresh-runtime`
* Get the status <br />
  > `venona status`
  > `kubectl get pods -n codefresh-runtime`


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
  * `service-account.venona.yaml` - The service account that the agent's pod will use at the end
  * `cluster-role-binding.venona.yaml` - The agent discovering K8S apis by calling to `openapi/v2`, this ClusterRoleBinding binds  bootstraped ClusterRole by Kubernetes `system:discovery` to `service-account.venona.yaml`. This role has only permissions to make a GET calls to non resources urls
  * `role.venona.yaml` - Allow to GET, CREATE and DELETE pods and persistentvolumeclaims
  * `role-binding.venona.yaml` - The agent is spinning up pods and pvc, this biniding binds `role.venona.yaml` to `service-account.venona.yaml`
* Runtime-environment (grouped by `/.*.re.yaml/`) Kubernetes controller that spins up all required resources to provide a good caching expirience during pipeline execution
  * `service-account.dind-volume-provisioner.re.yaml` - The service account that the controller will use
  * `cluster-role.dind-volume-provisioner.re.yaml` Defines all the permission needed for the controller to operate correctly
  * `cluster-role-binding.dind-volume-provisioner.yaml` - Binds the ClusterRole to `service-account.dind-volume-provisioner.re.yaml`

### Access the cluster from executed pipeline
After a successfull installation of Venona, you'll be able to run a Codefresh pipeline on the configured cluster.
However, the pipeline itself dosent have any permission to connect to the hosted cluster.
To make it work you need to add the cluster to Codefresh (make sure the service acount has all the permissions you need)
> codefresh create cluster --kube-context CONTEXT_NAME --namespace NAMESPACE --serviceaccount SERVICE_ACCOUNT --behind-firewall

#### Upgrade
To upgrade existing runtime-environment, a one that was created without Venona's agent, run:
* Find the name of the cluster was linked to that runtime environment <br />
Example: `codefresh get cluster`
* Install <br />
Example: `venona install --cluster-name CLUSTER`
* Get the status <br />
Example: `venona status RUNTIME-ENVIRONMENT`
Example: `kubectl get pods -n NAMESPACE`
