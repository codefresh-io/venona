Version 1.0 brings a lot of changes, both in the way Codefresh addressing the underline resource and they way we expose the management of it.
* We highly suggest to no use `venona` binary directly from now on, use [Codefresh official CLI](https://cli.codefresh.io/).
* Introduction of agent entity in Codefresh, from now on, all the new runtime-environments will be managed by an agent, see `codefresh get agent --help`.
    * Account can have multiple agents.
    * Agent can manage multiple runtime environments.
    * Runtime environment can be managed by one agent.
    * `codefresh create agent` will print you a token (this is a limited scoped token), that will be used by the Venona pod to talk to Codefresh API.
    * Venona pod (or the process) must have a network connection the the runtime environmens that he is manages
* Seperation of commands
    * `venona install` is now deprecated, instead we have `codefresh install agent` and `codefresh install runtime` which will use Venona binary.
* `codefresh install agent` will not create a service-account,role and rolebinding, means the the Venona pod will start with the default service account of the namespace.
* `codefresh install runtime` will create additional service-account,role and rolebinding, this service-account data will be injected to Venona pod during `codefresh attach runtime` flow.
* `codefresh/venona:latest` image is now the latest of the 0.x.x versions