# isserctl

Manages Codefresh runtime environment

```
Usage: isserctl <command> [options]

Commands:
        install --api-key <codefresh api-key> --cluster-name <cluster-name> [--url <codefresh url>] [kube params]

        status [kube params]

        delete [kube params]

Options:
   [kube params]
        kubeconfig
        kubecontext
        namespace
```

### Build
Set Go environment + dep and `build.sh`
`isserctl` will be in $GOPATH/bin

### Push for public downloading
To be done

### `isserctl install` Flow
- call Codefresh api to validate api-key and get some data
- generate Csr, submit it for signing to Codefresh 
- Create Config object
- Parse and execute all the templates in ./templates/<runtime-type>/ into map of k8s.io/apimachinery/pkg/runtime.Object 
- Post all the objects to kubernetes

### Templates
isserctl applies kubernetes objects generated from templates in ./templates/<runtime-type>/
These are go-templates with gomplate functions - see https://gomplate.hairyhenderson.ca/ 

The template values are provided from Config struct (from runtimectl/types.go)

###### Adding new templates
Just add valid template files of kubernetes yamls into ./templates/<runtime-type>/
and `build.sh`
we are using `go generate ` to create templates.go 



