# ISSER

## Installation

### Prerequisite:
* [Kubernetes](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - Used to create resource in your K8S cluster
* [Codefresh](https://codefresh-io.github.io/cli/) - Used to create resource in Codefresh
* [gomplate](https://gomplate.hairyhenderson.ca/) - Used to render K8S resources


### Install Isser

* Create namespace where Isser should run 
Example: `kubectl create namespace codefresh-runtime`
* Create a cluster in Codefresh 
Example: `codefresh create clusters --kube-context YOUR_KUBE_CONTEXT --behind-firewall --namespace codefresh-runtime`
* Create runtime-environment in Codefresh 
Example: `codefresh create re --cluster YOUR_KUBE_CONTEXT --namespace codefresh-runtime --kube-context YOUR_KUBE_CONTEXT`
* Create token for just created runtime-environment 
Example: `codefresh create token --name TOKEN_NAME --type runtime-environment --subject YOUR_KUBE_CONTEXT/codefresh-runtime`
* Encode the token and export it as `CODEFRESH_TOKEN_B64_ENCODED` environment variable
Example: `echo -n "TOKEN" | base64`
* export environment variables: 
Example: `export AGENT_NAME=codefresh-runtime AGENT_VERSION=1 APP_NAME=isser AGENT_NAMESPACE=codefresh-runtime CODEFRESH_HOST=https://g.codefresh.io AGENT_MODE=InCluster AGENT_IMAGE_NAME=codefresh/isser AGENT_IMAGE_TAG=master`
* Render K8S resources 
`gomplate -f kubernetes/template.tmpl --out kubernetes/resources.yaml`
* Apply resources 
`kubectl apply -f kubernetes/resources.yaml`
