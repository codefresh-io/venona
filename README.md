# VENONA
[![Go Report Card](https://goreportcard.com/badge/github.com/codefresh-io/venona)](https://goreportcard.com/report/github.com/codefresh-io/venona) 
[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/codefresh-inc/codefresh-io%2Fvenona%2Fvenona?type=cf-1)]( https://g.codefresh.io/public/accounts/codefresh-inc/pipelines/codefresh-io/venona/venona)

## Installation

### Prerequisite:
* [Kubernetes](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - Used to create resource in your K8S cluster
* [Codefresh](https://codefresh-io.github.io/cli/) - Used to create resource in Codefresh


### Install venona
#### Fresh installation
* Download [venona's](https://github.com/codefresh-io/venona/releases) binary
* Create namespace where venona should run<br />
Example: `kubectl create namespace codefresh-runtime`
* Create *new* runtime-environment with Venona's agents installed <br />
Example: `venona install --kube-namespace codefresh-runtime`
* Get the status <br />
Example: `venona status`  
Example: `kubectl get pods -n codefresh-runtime`

#### Upgrade
To upgrade existing runtime-environment, a one that was created without Venona's agent, run:
* Find the name of the environment <br />
Example: `codefresh get re`
* Install <br />
Example: `venona install --skip-runtime-installation --runtime-environment RUNTIME-ENVIRONMENT`
* Get the status <br />
Example: `venona status RUNTIME-ENVIRONMENT`  
Example: `kubectl get pods -n NAMESPACE`