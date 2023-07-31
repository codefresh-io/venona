## Codefresh Runner

![Version: 3.0.7](https://img.shields.io/badge/Version-3.0.7-informational?style=flat-square)

Helm chart for deploying [Codefresh Runner](https://codefresh.io/docs/docs/installation/codefresh-runner/) to Kubernetes.

## Table of Content

- [Prerequisites](#prerequisites)
- [Get Repo Info](#get-repo-info)
- [Install Chart](#install-chart)
- [Upgrade Chart](#upgrade-chart)
  - [To 2.x](#to-2x)
  - [To 3.x](#to-3x)
- [Architecture](#architecture)
- [Configuration](#configuration)
  - [EBS backend volume configuration](#ebs-backend-volume-configuration)
  - [Custom volume mounts](#custom-volume-mounts)
  - [Custom global environment variables](#custom-global-environment-variables)
  - [Volume reuse policy](#volume-reuse-policy)
  - [Volume cleaners](#volume-cleaners)

## Prerequisites

- Kubernetes **1.19+**
- Helm **3.8.0+**

## Get Repo Info

```console
helm repo add cf-runtime http://chartmuseum.codefresh.io/cf-runtime
helm repo update
```

## Install Chart

**Important:** only helm3 is supported

1. Download the Codefresh CLI and authenticate it with your Codefresh account. Follow [here](https://codefresh-io.github.io/cli/getting-started/) for more detailed instructions.
2. Run the following command to create mandatory values for Codefresh Runner:

    ```console
    codefresh runner init --generate-helm-values-file
    ```

   * This will not install anything on your cluster, except for running cluster acceptance tests, which may be skipped using the `--skip-cluster-test` option.
   * This command will also generate a `generated_values.yaml` file in your current directory, which you will need to provide to the `helm upgrade` command later.
3. Run the following to complete the installation:

    ```console
    helm repo add cf-runtime https://chartmuseum.codefresh.io/cf-runtime

    helm upgrade --install cf-runtime cf-runtime/cf-runtime -f ./generated_values.yaml --create-namespace --namespace codefresh
    ```

    *Install from OCI-based registry*
    ```console
    helm upgrade --install cf-runtime oci://quay.io/codefresh/cf-runtime -f ./generated_values.yaml --create-namespace --namespace codefresh
    ```
4. At this point you should have a working Codefresh Runner. You can verify the installation by running:
    ```console
    codefresh runner execute-test-pipeline --runtime-name <runtime-name>
    ```

## Upgrade Chart

### To 2.x

This major release renames and deprecated several values in the chart. Most of the workload templates have been refactored.

Affected values:
- `dockerRegistry` is deprecated. Replaced with `global.imageRegistry`
- `re` is renamed to `runtime`
- `storage.localVolumeMonitor` is replaced with `volumeProvisioner.dind-lv-monitor`
- `volumeProvisioner.volume-cleanup` is replaced with `volumeProvisioner.dind-volume-cleanup`
- `image` values structure has been updated. Split to `image.registry` `image.repository` `image.tag`
- pod's `annotations` is renamed to `podAnnotations`

### To 3.x

⚠️⚠️⚠️
### Please, READ this before the upgrade!

This major release adds [runtime-environment](https://codefresh.io/docs/docs/installation/codefresh-runner/#runtime-environment-specification) spec into chart templates.
That means it is possible to set parametes for `dind` and `engine` pods via [values.yaml](./values.yaml).

**If you had any overrides (i.e. tolerations/nodeSelector/environment variables/etc) added in runtime spec via [codefresh CLI](https://codefresh-io.github.io/cli/) (for example, you did use [get](https://codefresh-io.github.io/cli/runtime-environments/get-runtime-environments/) and [patch](https://codefresh-io.github.io/cli/runtime-environments/apply-runtime-environments/) commands to modify the runtime-environment), you MUST add these into chart's [values.yaml](./values.yaml) for `.Values.runtime.dind` or(and) .`Values.runtime.engine`**

**For backward compatibility, you can disable updating runtime-environment spec via** `.Values.runtime.patch.enabled=false`

Affected values:
- added **mandatory** `global.codefreshToken`/`global.codefreshTokenSecretKeyRef` **You must specify it before the upgrade!**
- `runtime.engine` is added
- `runtime.dind` is added
- `global.existingAgentToken` is replaced with `global.agentTokenSecretKeyRef`
- `global.existingDindCertsSecret` is replaced with `global.dindCertsSecretRef`

## Architecture

[Codefresh Runner architecture](https://codefresh.io/docs/docs/installation/codefresh-runner/#codefresh-runner-architecture)

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values cf-runtime/cf-runtime
```

### EBS backend volume configuration

`dind-volume-provisioner` should have permissions to create/attach/detach/delete/get EBS volumes

Minimal IAM policy for `dind-volume-provisioner`

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:AttachVolume",
        "ec2:CreateSnapshot",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:DeleteSnapshot",
        "ec2:DeleteTags",
        "ec2:DeleteVolume",
        "ec2:DescribeInstances",
        "ec2:DescribeSnapshots",
        "ec2:DescribeTags",
        "ec2:DescribeVolumes",
        "ec2:DetachVolume"
      ],
      "Resource": "*"
    }
  ]
}
```

There are three options:

1. Run `dind-volume-provisioner` pod on the node/node-group with IAM role

```yaml
storage:
  # -- Set backend volume type (`local`/`ebs`/`ebs-csi`/`gcedisk`/`azuredisk`)
  backend: ebs-csi

  ebs:
    availabilityZone: "us-east-1a"

volumeProvisioner:
  # -- Set node selector
  nodeSelector: {}
  # -- Set tolerations
  tolerations: []
```

2. Pass static credentials in `.Values.storage.ebs.accessKeyId/accessKeyIdSecretKeyRef` and `.Values.storage.ebs.secretAccessKey/secretAccessKeySecretKeyRef`

```yaml
storage:
  # -- Set backend volume type (`local`/`ebs`/`ebs-csi`/`gcedisk`/`azuredisk`)
  backend: ebs-csi

  ebs:
    availabilityZone: "us-east-1a"

    # -- Set AWS_ACCESS_KEY_ID for volume-provisioner (optional)
    accessKeyId: ""
    # -- Existing secret containing AWS_ACCESS_KEY_ID.
    accessKeyIdSecretKeyRef: {}
    # E.g.
    # accessKeyIdSecretKeyRef:
    #   name:
    #   key:

    # -- Set AWS_SECRET_ACCESS_KEY for volume-provisioner (optional)
    secretAccessKey: ""
    # -- Existing secret containing AWS_SECRET_ACCESS_KEY
    secretAccessKeySecretKeyRef: {}
    # E.g.
    # secretAccessKeySecretKeyRef:
    #   name:
    #   key:
```

3. Assign IAM role to `dind-volume-provisioner` service account

```yaml
storage:
  # -- Set backend volume type (`local`/`ebs`/`ebs-csi`/`gcedisk`/`azuredisk`)
  backend: ebs-csi

  ebs:
    availabilityZone: "us-east-1a"

volumeProvisioner:
  # -- Service Account parameters
  serviceAccount:
    # -- Create service account
    create: true
    # -- Additional service account annotations
    serviceAccount:
      annotations:
        eks.amazonaws.com/role-arn: "arn:aws:iam::<ACCOUNT_ID>:role/<IAM_ROLE_NAME>"
```

### Custom volume mounts

You can add your own volumes and volume mounts in the runtime environment, so that all pipeline steps will have access to the same set of external files.

```yaml
runtime:
  dind:
    userVolumes:
      regctl-docker-registry:
        name: regctl-docker-registry
        secret:
          items:
            - key: .dockerconfigjson
              path: config.json
          secretName: regctl-docker-registry
          optional: true
    userVolumeMounts:
      regctl-docker-registry:
        name: regctl-docker-registry
        mountPath: /home/appuser/.docker/
        readOnly: true

```

### Custom global environment variables

You can add your own environment variables to the runtime environment. All pipeline steps have access to the global variables.

```yaml
runtime:
  engine:
    userEnvVars:
    - name: GITHUB_TOKEN
      valueFrom:
        secretKeyRef:
          name: github-token
          key: token
```

### Volume reuse policy

Volume reuse behavior depends on the configuration for `reuseVolumeSelector` in the runtime environment spec.

```yaml
runtime:
  dind:
    pvcs:
      - name: dind
        ...
        reuseVolumeSelector: 'codefresh-app,io.codefresh.accountName'
        reuseVolumeSortOrder: pipeline_id
```

The following options are available:
- `reuseVolumeSelector: 'codefresh-app,io.codefresh.accountName'` - PV can be used by ANY pipeline in the specified account (default).
Benefit: Fewer PVs, resulting in lower costs. Since any PV can be used by any pipeline, the cluster needs to maintain/reserve fewer PVs in its PV pool for Codefresh.
Downside: Since the PV can be used by any pipeline, the PVs could have assets and info from different pipelines, reducing the probability of cache.

- `reuseVolumeSelector: 'codefresh-app,io.codefresh.accountName,project_id'` - PV can be used by ALL pipelines in your account, assigned to the same project.

- `reuseVolumeSelector: 'codefresh-app,io.codefresh.accountName,pipeline_id'` - PV can be used only by a single pipeline.
Benefit: More probability of cache without “spam” from other pipelines.
Downside: More PVs to maintain and therefore higher costs.

- `reuseVolumeSelector: 'codefresh-app,io.codefresh.accountName,pipeline_id,io.codefresh.branch_name'` - PV can be used only by single pipeline AND single branch.

- `reuseVolumeSelector: 'codefresh-app,io.codefresh.accountName,pipeline_id,trigger'` - PV can be used only by single pipeline AND single trigger.

### Volume cleaners

Codefresh pipelines require disk space for:
  * [Pipeline Shared Volume](https://codefresh.io/docs/docs/pipelines/introduction-to-codefresh-pipelines/#sharing-the-workspace-between-build-steps) (`/codefresh/volume`, implemented as [docker volume](https://docs.docker.com/storage/volumes/))
  * Docker containers, both running and stopped
  * Docker images and cached layers

Codefresh offers two options to manage disk space and prevent out-of-space errors:
* Use runtime cleaners on Docker images and volumes
* [Set the minimum disk space per pipeline build volume](https://codefresh.io/docs/docs/pipelines/pipelines/#set-minimum-disk-space-for-a-pipeline-build)

To improve performance by using Docker cache, Codefresh `volume-provisioner` can provision previously used disks with Docker images and pipeline volumes from previously run builds.

### Types of runtime volume cleaners

Docker images and volumes must be cleaned on a regular basis.

* [IN-DIND cleaner](https://github.com/codefresh-io/dind/tree/master/cleaner): Deletes extra Docker containers, volumes, and images in **DIND pod**.
* [External volume cleaner](https://github.com/codefresh-io/dind-volume-cleanup): Deletes unused **external** PVs (EBS, GCE/Azure disks).
* [Local volume cleaner](https://github.com/codefresh-io/dind-volume-utils/blob/master/local-volumes/lv-cleaner.sh): Deletes **local** volumes if node disk space is close to the threshold.

### IN-DIND cleaner

**Purpose:** Removes unneeded *docker containers, images, volumes* inside Kubernetes volume mounted on the DIND pod

**How it runs:** Inside each DIND pod as script

**Triggered by:** SIGTERM and also during the run when disk usage > 90% (configurable)

**Configured by:**  Environment Variables which can be set in Runtime Environment spec

**Configuration/Logic:** [README.md](https://github.com/codefresh-io/dind/tree/master/cleaner#readme)

Override `.Values.runtime.dind.env` if necessary (the following are **defaults**):

```yaml
runtime:
  dind:
    env:
      CLEAN_PERIOD_SECONDS: '21600' # launch clean if last clean was more than CLEAN_PERIOD_SECONDS seconds ago
      CLEAN_PERIOD_BUILDS: '5' # launch clean if last clean was more CLEAN_PERIOD_BUILDS builds since last build
      IMAGE_RETAIN_PERIOD: '14400' # do not delete docker images if they have events since current_timestamp - IMAGE_RETAIN_PERIOD
      VOLUMES_RETAIN_PERIOD: '14400' # do not delete docker volumes if they have events since current_timestamp - VOLUMES_RETAIN_PERIOD
      DISK_USAGE_THRESHOLD: '0.8' # launch clean based on current disk usage DISK_USAGE_THRESHOLD
      INODES_USAGE_THRESHOLD: '0.8' # launch clean based on current inodes usage INODES_USAGE_THRESHOLD
```

### External volumes cleaner

**Purpose:** Removes unused *kubernetes volumes and related backend volumes*

**How it runs:** Runs as `dind-volume-cleanup` CronJob. Installed in case the Runner uses non-local volumes `.Values.storage.backend != local`

**Triggered by:** CronJob every 10min (configurable)

**Configuration:**

Set `codefresh.io/volume-retention` for dinds' PVCs:

```yaml
runtime:
  dind:
    pvcs:
      - name: dind
        ...
        annotations:
          codefresh.io/volume-retention: 7d
```

Or override environment variables for `dind-volume-cleanup` cronjob:

```yaml
volumeProvisioner:
  dind-volume-cleanup:
    env:
      RETENTION_DAYS: 7   # clean volumes that were last used more than `RETENTION_DAYS` (default is 4) ago
```

### Local volumes cleaner

**Purpose:** Deletes local volumes when node disk space is close to the threshold

**How it runs:** Runs as `dind-lv-monitor` DaemonSet. Installed in case the Runner uses local volumes `.Values.storage.backend == local`

**Triggered by:** Disk space usage or inode usage that exceeds thresholds (configurable)

**Configuration:**

Override environment variables for `dind-lv-monitor` daemonset:

```yaml
volumeProvisioner:
  dind-lv-monitor:
    env:
      KB_USAGE_THRESHOLD: 60  # default 80 (percentage)
      INODE_USAGE_THRESHOLD: 60  # default 80
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://chartmuseum.codefresh.io/cf-common | cf-common | 0.11.1 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| appProxy.affinity | object | `{}` | Set affinity |
| appProxy.enabled | bool | `false` | Enable app-proxy |
| appProxy.env | object | `{}` | Add additional env vars |
| appProxy.image | object | `{"registry":"quay.io","repository":"codefresh/cf-app-proxy","tag":"latest"}` | Set image |
| appProxy.ingress.annotations | object | `{}` | Set extra annotations for ingress object |
| appProxy.ingress.class | string | `""` | Set ingress class |
| appProxy.ingress.host | string | `""` | Set DNS hostname the ingress will use |
| appProxy.ingress.pathPrefix | string | `"/"` | Set path prefix for ingress |
| appProxy.ingress.tlsSecret | string | `""` | Set k8s tls secret for the ingress object |
| appProxy.nodeSelector | object | `{}` | Set node selector |
| appProxy.podAnnotations | object | `{}` | Set pod annotations |
| appProxy.podSecurityContext | object | `{}` | Set security context for the pod |
| appProxy.rbac | object | `{"create":true,"namespaced":true,"rules":[]}` | RBAC parameters |
| appProxy.rbac.create | bool | `true` | Create RBAC resources |
| appProxy.rbac.namespaced | bool | `true` | Use Role(true)/ClusterRole(true) |
| appProxy.rbac.rules | list | `[]` | Add custom rule to the role |
| appProxy.readinessProbe | object | See below | Readiness probe configuration |
| appProxy.replicasCount | int | `1` | Set number of pods |
| appProxy.resources | object | `{}` | Set requests and limits |
| appProxy.serviceAccount | object | `{"annotations":{},"create":true,"name":"","namespaced":true}` | Service Account parameters |
| appProxy.serviceAccount.annotations | object | `{}` | Additional service account annotations |
| appProxy.serviceAccount.create | bool | `true` | Create service account |
| appProxy.serviceAccount.name | string | `""` | Override service account name |
| appProxy.serviceAccount.namespaced | bool | `true` | Use Role(true)/ClusterRole(true) |
| appProxy.tolerations | list | `[]` | Set tolerations |
| appProxy.updateStrategy | object | `{"type":"RollingUpdate"}` | Upgrade strategy |
| dockerRegistry | string | `""` |  |
| extraResources | list | `[]` | Array of extra objects to deploy with the release |
| global | object | See below | Global parameters Global values are in generated_values.yaml. Run `codefresh runner init --generate-helm-values-file`! |
| global.accountId | string | `""` | Account ID |
| global.agentId | string | `""` | Agent ID |
| global.agentName | string | `""` | Agent Name |
| global.agentToken | string | `""` | Agent token in plain text. |
| global.agentTokenSecretKeyRef | object | `{}` | Agent token that references an existing secret containing API key. |
| global.codefreshHost | string | `""` | URL of Codefresh Platform |
| global.codefreshToken | string | `""` | User token in plain text. Ref: https://g.codefresh.io/user/settings (see API Keys) |
| global.codefreshTokenSecretKeyRef | object | `{}` | User token that references an existing secret containing API key. |
| global.dindCertsSecretRef | object | `{}` | Certs for Dind docker daemon that references an existing secret |
| global.imagePullSecrets | list | `[]` | Global Docker registry secret names as array |
| global.imageRegistry | string | `""` | Global Docker image registry |
| global.keys | object | `{"ca":"","key":"","serverCert":""}` | Certs for Dind docker daemon |
| global.namespace | string | `""` | Runner namespace |
| global.runtimeName | string | `""` | Runtime name |
| monitor.affinity | object | `{}` | Set affinity |
| monitor.clusterId | string | `""` | Cluster name as it registered in account Generated from `codefresh runner init --generate-helm-values-file` output |
| monitor.enabled | bool | `false` | Enable monitor Ref: https://codefresh.io/docs/docs/installation/codefresh-runner/#install-monitoring-component |
| monitor.env | object | `{}` | Add additional env vars |
| monitor.existingMonitorToken | string | `""` | Set Existing secret (name-of-existing-secret) with API token from Codefresh supersedes value of monitor.token; secret must contain `codefresh.token` key |
| monitor.image | object | `{"registry":"quay.io","repository":"codefresh/agent","tag":"stable"}` | Set image |
| monitor.nodeSelector | object | `{}` | Set node selector |
| monitor.podAnnotations | object | `{}` | Set pod annotations |
| monitor.podSecurityContext | object | `{}` |  |
| monitor.rbac | object | `{"create":true,"namespaced":true,"rules":[]}` | RBAC parameters |
| monitor.rbac.create | bool | `true` | Create RBAC resources |
| monitor.rbac.namespaced | bool | `true` | Use Role(true)/ClusterRole(true) |
| monitor.rbac.rules | list | `[]` | Add custom rule to the role |
| monitor.readinessProbe | object | See below | Readiness probe configuration |
| monitor.replicasCount | int | `1` | Set number of pods |
| monitor.resources | object | `{}` | Set resources |
| monitor.serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | Service Account parameters |
| monitor.serviceAccount.annotations | object | `{}` | Additional service account annotations |
| monitor.serviceAccount.create | bool | `true` | Create service account |
| monitor.serviceAccount.name | string | `""` | Override service account name |
| monitor.token | string | `""` | API token from Codefresh Generated from `codefresh runner init --generate-helm-values-file` output |
| monitor.tolerations | list | `[]` | Set tolerations |
| monitor.updateStrategy | object | `{"type":"RollingUpdate"}` | Upgrade strategy |
| re | object | `{}` |  |
| runner | object | See below | Runner parameters |
| runner.affinity | object | `{}` | Set affinity |
| runner.env | object | `{}` | Add additional env vars |
| runner.image | object | `{"registry":"quay.io","repository":"codefresh/venona","tag":"1.9.16"}` | Set image |
| runner.nodeSelector | object | `{}` | Set node selector |
| runner.podAnnotations | object | `{}` | Set pod annotations |
| runner.podSecurityContext | object | See below | Set security context for the pod |
| runner.rbac | object | `{"create":true,"rules":[]}` | RBAC parameters |
| runner.rbac.create | bool | `true` | Create RBAC resources |
| runner.rbac.rules | list | `[]` | Add custom rule to the role |
| runner.readinessProbe | object | See below | Readiness probe configuration |
| runner.replicasCount | int | `1` | Set number of pods |
| runner.resources | object | `{}` | Set requests and limits |
| runner.serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | Service Account parameters |
| runner.serviceAccount.annotations | object | `{}` | Additional service account annotations |
| runner.serviceAccount.create | bool | `true` | Create service account |
| runner.serviceAccount.name | string | `""` | Override service account name |
| runner.tolerations | list | `[]` | Set tolerations |
| runner.updateStrategy | object | `{"type":"RollingUpdate"}` | Upgrade strategy |
| runtime | object | See below | Set runtime parameters |
| runtime.dind | object | `{"affinity":{},"env":{},"image":{"registry":"quay.io","repository":"codefresh/dind","tag":"20.10.18-1.25.7"},"nodeSelector":{},"podAnnotations":{},"pvcs":[{"name":"dind","reuseVolumeSelector":"codefresh-app,io.codefresh.accountName","reuseVolumeSortOrder":"pipeline_id","storageClassName":"{{ include \"dind-volume-provisioner.storageClassName\" . }}","volumeSize":"16Gi"}],"resources":{"limits":{"cpu":"400m","memory":"800Mi"},"requests":null},"schedulerName":"","serviceAccount":"codefresh-engine","tolerations":[],"userAccess":true,"userVolumeMounts":{},"userVolumes":{}}` | Parameters for DinD (docker-in-docker) pod (aka "runtime" pod). |
| runtime.dind.affinity | object | `{}` | Set affinity |
| runtime.dind.env | object | `{}` | Set additional env vars. |
| runtime.dind.image | object | `{"registry":"quay.io","repository":"codefresh/dind","tag":"20.10.18-1.25.7"}` | Set dind image. |
| runtime.dind.nodeSelector | object | `{}` | Set node selector. |
| runtime.dind.podAnnotations | object | `{}` | Set pod annotations. |
| runtime.dind.pvcs | list | `[{"name":"dind","reuseVolumeSelector":"codefresh-app,io.codefresh.accountName","reuseVolumeSortOrder":"pipeline_id","storageClassName":"{{ include \"dind-volume-provisioner.storageClassName\" . }}","volumeSize":"16Gi"}]` | PV claim spec parametes. |
| runtime.dind.pvcs[0] | object | `{"name":"dind","reuseVolumeSelector":"codefresh-app,io.codefresh.accountName","reuseVolumeSortOrder":"pipeline_id","storageClassName":"{{ include \"dind-volume-provisioner.storageClassName\" . }}","volumeSize":"16Gi"}` | PVC name prefix. Keep `dind` as default! Don't change! |
| runtime.dind.pvcs[0].reuseVolumeSelector | string | `"codefresh-app,io.codefresh.accountName"` | PV reuse selector. Ref: https://codefresh.io/docs/docs/installation/codefresh-runner/#volume-reuse-policy |
| runtime.dind.pvcs[0].storageClassName | string | `"{{ include \"dind-volume-provisioner.storageClassName\" . }}"` | PVC storage class name. Change ONLY if you need to use storage class NOT from Codefresh volume-provisioner |
| runtime.dind.pvcs[0].volumeSize | string | `"16Gi"` | PVC size. |
| runtime.dind.resources | object | `{"limits":{"cpu":"400m","memory":"800Mi"},"requests":null}` | Set dind resources. |
| runtime.dind.schedulerName | string | `""` | Set scheduler name. |
| runtime.dind.serviceAccount | string | `"codefresh-engine"` | Set service account for pod. |
| runtime.dind.tolerations | list | `[]` | Set tolerations. |
| runtime.dind.userAccess | bool | `true` | Keep `true` as default! |
| runtime.dind.userVolumeMounts | object | `{}` | Add extra volume mounts |
| runtime.dind.userVolumes | object | `{}` | Add extra volumes |
| runtime.dindDaemon | object | See below | DinD pod daemon config |
| runtime.engine | object | `{"affinity":{},"command":["npm","run","start"],"env":{},"image":{"registry":"quay.io","repository":"codefresh/engine","tag":"1.164.9"},"nodeSelector":{},"podAnnotations":{},"resources":{"limits":{"cpu":"1000m","memory":"2048Mi"},"requests":{"cpu":"100m","memory":"128Mi"}},"runtimeImages":{"COMPOSE_IMAGE":"quay.io/codefresh/compose:1.3.0","CONTAINER_LOGGER_IMAGE":"quay.io/codefresh/cf-container-logger:1.10.2","DOCKER_BUILDER_IMAGE":"quay.io/codefresh/cf-docker-builder:1.3.5","DOCKER_PULLER_IMAGE":"quay.io/codefresh/cf-docker-puller:8.0.9","DOCKER_PUSHER_IMAGE":"quay.io/codefresh/cf-docker-pusher:6.0.12","DOCKER_TAG_PUSHER_IMAGE":"quay.io/codefresh/cf-docker-tag-pusher:1.3.9","FS_OPS_IMAGE":"quay.io/codefresh/fs-ops:1.2.3","GIT_CLONE_IMAGE":"quay.io/codefresh/cf-git-cloner:10.1.19","KUBE_DEPLOY":"quay.io/codefresh/cf-deploy-kubernetes:16.1.11","PIPELINE_DEBUGGER_IMAGE":"quay.io/codefresh/cf-debugger:1.3.0","TEMPLATE_ENGINE":"quay.io/codefresh/pikolo:0.13.8"},"schedulerName":"","serviceAccount":"codefresh-engine","tolerations":[],"userEnvVars":[]}` | Parameters for Engine pod (aka "pipeline" orchestrator). |
| runtime.engine.affinity | object | `{}` | Set affinity |
| runtime.engine.command | list | `["npm","run","start"]` | Set container command. |
| runtime.engine.env | object | `{}` | Set additional env vars. |
| runtime.engine.image | object | `{"registry":"quay.io","repository":"codefresh/engine","tag":"1.164.9"}` | Set image. |
| runtime.engine.nodeSelector | object | `{}` | Set node selector. |
| runtime.engine.podAnnotations | object | `{}` | Set pod annotations. |
| runtime.engine.resources | object | `{"limits":{"cpu":"1000m","memory":"2048Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Set resources. |
| runtime.engine.runtimeImages | object | See below. | Set system(base) runtime images. |
| runtime.engine.schedulerName | string | `""` | Set scheduler name. |
| runtime.engine.serviceAccount | string | `"codefresh-engine"` | Set service account for pod. |
| runtime.engine.tolerations | list | `[]` | Set tolerations. |
| runtime.engine.userEnvVars | list | `[]` | Set extra env vars |
| runtime.patch | object | See below | Parameters for `runtime-patch` post-upgrade/install hook |
| runtime.rbac | object | `{"create":true,"rules":[]}` | RBAC parameters |
| runtime.rbac.create | bool | `true` | Create RBAC resources |
| runtime.rbac.rules | list | `[]` | Add custom rule to the engine role |
| runtime.runtimeExtends | list | `["system/default/hybrid/k8s_low_limits"]` | Set parent runtime to inherit. Should not be changes. Parent runtime is controlled from Codefresh side. |
| runtime.serviceAccount | object | `{"annotations":{},"create":true}` | Set annotation on engine Service Account Ref: https://codefresh.io/docs/docs/administration/codefresh-runner/#injecting-aws-arn-roles-into-the-cluster |
| storage.azuredisk.cachingMode | string | `"None"` |  |
| storage.azuredisk.skuName | string | `"Premium_LRS"` | Set storage type (`Premium_LRS`) |
| storage.backend | string | `"local"` | Set backend volume type (`local`/`ebs`/`ebs-csi`/`gcedisk`/`azuredisk`) |
| storage.ebs.accessKeyId | string | `""` | Set AWS_ACCESS_KEY_ID for volume-provisioner (optional) Ref: https://codefresh.io/docs/docs/installation/codefresh-runner/#dind-volume-provisioner-permissions |
| storage.ebs.accessKeyIdSecretKeyRef | object | `{}` | Existing secret containing AWS_ACCESS_KEY_ID. |
| storage.ebs.availabilityZone | string | `"us-east-1a"` | Set EBS volumes availability zone (required) |
| storage.ebs.encrypted | string | `"false"` | Enable encryption (optional) |
| storage.ebs.kmsKeyId | string | `""` | Set KMS encryption key ID (optional) |
| storage.ebs.secretAccessKey | string | `""` | Set AWS_SECRET_ACCESS_KEY for volume-provisioner (optional) Ref: https://codefresh.io/docs/docs/installation/codefresh-runner/#dind-volume-provisioner-permissions |
| storage.ebs.secretAccessKeySecretKeyRef | object | `{}` | Existing secret containing AWS_SECRET_ACCESS_KEY |
| storage.ebs.volumeType | string | `"gp2"` | Set EBS volume type (`gp2`/`gp3`/`io1`) (required) |
| storage.fsType | string | `"ext4"` | Set filesystem type (`ext4`/`xfs`) |
| storage.gcedisk.availabilityZone | string | `"us-west1-a"` | Set GCP volume availability zone |
| storage.gcedisk.serviceAccountJson | string | `""` | Set Google SA JSON key for volume-provisioner (optional) |
| storage.gcedisk.serviceAccountJsonSecretKeyRef | object | `{}` | Existing secret containing containing Google SA JSON key for volume-provisioner (optional) |
| storage.gcedisk.volumeType | string | `"pd-ssd"` | Set GCP volume backend type (`pd-ssd`/`pd-standard`) |
| storage.local.volumeParentDir | string | `"/var/lib/codefresh/dind-volumes"` | Set volume path on the host filesystem |
| storage.mountAzureJson | bool | `false` |  |
| volumeProvisioner | object | See below | Volume Provisioner parameters |
| volumeProvisioner.affinity | object | `{}` | Set affinity |
| volumeProvisioner.dind-lv-monitor | object | See below | `dind-lv-monitor` DaemonSet parameters (local volumes cleaner) |
| volumeProvisioner.enabled | bool | `true` | Enable volume-provisioner |
| volumeProvisioner.env | object | `{}` | Add additional env vars |
| volumeProvisioner.image | object | `{"registry":"quay.io","repository":"codefresh/dind-volume-provisioner","tag":"1.33.3"}` | Set image |
| volumeProvisioner.nodeSelector | object | `{}` | Set node selector |
| volumeProvisioner.podAnnotations | object | `{}` | Set pod annotations |
| volumeProvisioner.podSecurityContext | object | See below | Set security context for the pod |
| volumeProvisioner.rbac | object | `{"create":true,"rules":[]}` | RBAC parameters |
| volumeProvisioner.rbac.create | bool | `true` | Create RBAC resources |
| volumeProvisioner.rbac.rules | list | `[]` | Add custom rule to the role |
| volumeProvisioner.replicasCount | int | `1` | Set number of pods |
| volumeProvisioner.resources | object | `{}` | Set resources |
| volumeProvisioner.serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | Service Account parameters |
| volumeProvisioner.serviceAccount.annotations | object | `{}` | Additional service account annotations |
| volumeProvisioner.serviceAccount.create | bool | `true` | Create service account |
| volumeProvisioner.serviceAccount.name | string | `""` | Override service account name |
| volumeProvisioner.tolerations | list | `[]` | Set tolerations |
| volumeProvisioner.updateStrategy | object | `{"type":"Recreate"}` | Upgrade strategy |

