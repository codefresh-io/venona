
// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at 2018-12-28 10:43:47.156562836 +0200 IST m=+0.000982863
// using data from templates/kubernetes_dind
package kubernetes_dind

func TemplatesMap() map[string]string {
    templatesMap := make(map[string]string)

templatesMap["dind-daemon-conf.yml"] = `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: codefresh-dind-config
data:
  daemon.json: |
    {
      "hosts": [ "unix:///var/run/docker.sock",
                 "unix:///var/run/codefresh/docker.sock",
                 "tcp://0.0.0.0:1300"],
      "storage-driver": "overlay",
      "storage-opts": ["overlay.override_kernel_check=1"],
      "tlsverify": true,  
      "tls": true,
      "tlscacert": "/etc/ssl/cf-client/ca.pem",
      "tlscert": "/etc/ssl/cf/server-cert.pem",
      "tlskey": "/etc/ssl/cf/server-key.pem",
      "insecure-registries" : ["192.168.99.100:5000"],
      "metrics-addr" : "0.0.0.0:9323",
      "experimental" : true
    }
` 

templatesMap["dind-headless-service.yml"] = `---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: dind
  name: dind
spec:
  ports:
  - name: "dind-port"
    port: 1300
    protocol: TCP

  # This is a headless service, Kubernetes won't assign a VIP for it.
  # *.dind.default.svc.cluster.local
  clusterIP: None
  selector:
    app: dind

` 

templatesMap["runtime-conf.yml"] = `##
## RUNTIME CONFIGURATION
##
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: codefresh-resource-limitations
data:
  resource-limitations.json: |
    {}
---
apiVersion: v1
kind: Secret
metadata:
  name: codefresh-internal-registry
stringData:
  internal-registry.json: |
    {
      "kind": "standard" ,
      "domain": "{{CF_REGISTRY_DOMAIN}}",
      "username": "{{CF_REGISTRY_USER}}",
      "password": "{{REGISTRY_TOKEN}}",
      "repositoryPrefix": "internal",
      "connection": {
        "protocol": "{{CF_REGISTRY_PROTOCOL}}"
      }
    }
  additional-internal-registries.json: |
    []` 

    return  templatesMap
}
