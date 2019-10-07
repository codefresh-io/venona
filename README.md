# VENONA
[![Go Report Card](https://goreportcard.com/badge/github.com/codefresh-io/venona)](https://goreportcard.com/report/github.com/codefresh-io/venona) 
[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/codefresh-inc/codefresh-io%2Fvenona%2Fvenona?type=cf-1)]( https://g.codefresh.io/public/accounts/codefresh-inc/pipelines/codefresh-io/venona/venona)

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

#### Install on GKE using local SSD's
  * This setup assumes you've created a node pool with locally attached SSD's. [Terraform node pool](https://www.terraform.io/docs/providers/google/r/container_cluster.html#node_config) options for GKE.
  * Ensure you've configured your cluster with the steps detailed in the Install on GCP section.
  * Delete and recreate the dind-local-volumes-venona-<namespace> StorageClass, ensuring that the volumeParentDir field is set:
```
parameters:
  volumeBackend: local
  volumeParentDir: /mnt/disks/ssd0/codefresh/dind-volumes
```
  * Dump your runtime-environment config to your local disk via: `codefresh get runtime-environments <runtime env  name> -o yaml > mycustom-runtime.yaml`
  * Edit the nodeSelector param for the `runtimeScheduler` to ensure the dind pods / volumes only exist on local-ssd nodes.
```
nodeSelector:
  cloud.google.com/gke-local-ssd: true
```
  * Save the file and patch `codefresh patch runtime-environments -f ./mycustom-runtime.yaml`
  * Edit dind-lv-monitor-venona DaemonSet and set env var `VOLUME_PARENT_DIR: /mnt/disks/ssd0/codefresh/dind-volumes`. Once it's done delete old pods of DS dind-lv-monitor-venona to let the new changes take effect.

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

