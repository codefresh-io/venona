# Venona


* venona - the agent process that is running on remote cluster
    * cmd - entrypoints to the application
    * pkg/agent - call Codefresh API every X ms to get new pipelines to run. Also, report status back to Codefresh
    * pkg/codefresh - Codefresh API client
    * pkg/config - Interface to load the attached runtimes from the filesystem
    * pkg/kubernetes - Interface to Kubernetes
    * pkg/logger - logger
    * pkg/runtime - Interface that uses Kubernetes API to start the pipeline