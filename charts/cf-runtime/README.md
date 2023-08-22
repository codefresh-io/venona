## Codefresh Runner

![Version: 6.0.0](https://img.shields.io/badge/Version-6.0.0-informational?style=flat-square)

Helm chart for deploying [Codefresh Runner](https://codefresh.io/docs/docs/installation/codefresh-runner/) to Kubernetes.

## Table of Content

- [Prerequisites](#prerequisites)
- [Get Repo Info](#get-repo-info)
- [Install Chart](#install-chart)
- [Chart Configuration](#chart-configuration)
- [Upgrade Chart](#upgrade-chart)
  - [To 2.x](#to-2x)
  - [To 3.x](#to-3x)
  - [To 4.x](#to-4x)
  - [To 5.x](#to-5x)
  - [To 6.x](#to-5x)
- [Architecture](#architecture)
- [Configuration](#configuration)
  - [EBS backend volume configuration](#ebs-backend-volume-configuration)
  - [Custom volume mounts](#custom-volume-mounts)
  - [Custom global environment variables](#custom-global-environment-variables)
  - [Volume reuse policy](#volume-reuse-policy)
  - [Volume cleaners](#volume-cleaners)
  - [Openshift](#openshift)
  - [On-premise](#on-premise)

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

- Specify the following mandatory values

```yaml
# -- Global parameters
# @default -- See below
global:
  # -- User token in plain text (required if `global.codefreshTokenSecretKeyRef` is omitted!)
  # Ref: https://g.codefresh.io/user/settings (see API Keys)
  codefreshToken: ""
  # -- User token that references an existing secret containing API key (required if `global.codefreshToken` is omitted!)
  codefreshTokenSecretKeyRef: {}
  # E.g.
  # codefreshTokenSecretKeyRef:
  #   name: my-codefresh-api-token
  #   key: codefresh-api-token

  # -- Account ID (required!)
  # Can be obtained here https://g.codefresh.io/2.0/account-settings/account-information
  accountId: ""

  # -- K8s context name (required!)
  context: ""
  # E.g.
  # context: prod-ue1-runtime-1

  # -- Agent Name (optional!)
  # If omitted, the following format will be used `