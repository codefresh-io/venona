# ISSER

## Installation

### Prerequisite:
* [Kubernetes](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - Used to create resource in your K8S cluster
* [Codefresh](https://codefresh-io.github.io/cli/) - Used to create resource in Codefresh
* [gomplate](https://gomplate.hairyhenderson.ca/) - Used to render K8S resources


### Install Isser

* Create namespace where Isser should run (Example: `kubectl create namespace codefresh-runtime`)
* Create a cluster in Codefresh (Example: `codefresh create clusters --kube-context YOUR_KUBE_CONTEXT --behind-firewall --namespace codefresh-runtime`)
* Create runtime-environment in Codefresh [follow instructions](https://github.com/codefresh-io/k8s-dind-config)
* Create token for just created runtime-environment (Example: `codefresh create token --name TOKEN_NAME --subject YOUR_KUBE_CONTEXT/codefresh-runtime`)
    * Encode the token and export it as: `CODEFRESH_TOKEN_B64_ENCODED` (Example: `echo -n "TOKEN" | base64`)
* Render K8S resources (`gomplate -f kubernetes/template.tmpl --out kubernetes/resources.yaml`)
* Apply resources (`kubectl apply -f kubernetes/resources.yaml`)