# Quick Development Start Guide

This guide walks through setting up a local development environment for
`harvester-mcp-server`, end to end: Rancher + Harvester, the Rancher AI Agent
stack, the AI UI extension, and finally `harvester-mcp-server` itself.

## Prerequisites

- A Rancher Prime deployment.
- A Harvester cluster, imported into Rancher.

## 1. Install Rancher AI Agent and Rancher MCP Server

> [!NOTE]
> Perform this step on the Rancher `local` cluster (the cluster that runs Rancher).

Follow the [quick start guide](https://documentation.suse.com/cloudnative/rancher-ai/latest/en/quick-start.html)
to bring up `rancher-ai-agent` and `rancher-ai-mcp`.

Clone `rancher-ai-agent` repo:

```bash
# export KUBECONFIG=/path/to/local/cluster/kubeconfig
git clone https://github.com/rancher/rancher-ai-agent
cd rancher-ai-agent
```

For development, create a `values.yaml` first:

```bash
cat > values.yaml <<EOF
global:
  cattle:
    systemDefaultRegistry: <dev-registry>

# Edit your LLM preferences here
# Gemini example:
googleApiKey: <google_api_key>
geminiLlmModel: <model>
activeLlm: gemini

insecureSkipTls: true

aiAgent:
  image:
    repository: rancher/rancher-ai-agent
    tag: v1.1.0-alpha.8
    pullPolicy: IfNotPresent
mcp:
  readOnly: false
  image:
    repository: rancher/rancher-ai-mcp
    tag: v1.1.0-alpha.7
    pullPolicy: IfNotPresent

log:
  level: debug
EOF
```

Check the [rancher-ai-agent tags](https://github.com/rancher/rancher-ai-agent/tags)
and [rancher-ai-mcp tags](https://github.com/rancher/rancher-ai-mcp/tags) for
the latest versions. Non-formal releases are only available from the dev registry, consult developers and set `systemDefaultRegistry` accordingly.

Install the chart:

```bash
# export KUBECONFIG=/path/to/local/cluster/kubeconfig
helm install rancher-ai-agent \
  --namespace cattle-ai-agent-system \
  --create-namespace \
  -f values.yaml chart/agent
```

Verify both deployments are up in `cattle-ai-agent-system`:

```bash
$ kubectl get deployments -n cattle-ai-agent-system
NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
rancher-ai-agent       1/1     1            1           4m15s
rancher-mcp-server     1/1     1            1           4m15s
```

## 2. Install rancher-ai-gui

Follow the [UI extension guide](https://documentation.suse.com/cloudnative/rancher-ai/latest/en/quick-start.html#_install_the_ui_extension).

Alternatively, install a development version:

1. On the `local` cluster, go to **Apps > Repositories > Create**, choose
   **Git Repository**, and fill in the required fields. For example, to use an alternative repo:

   - **Name**: alternative-ai-ui
   - **Git Repo URL**: `<repo_url>`
   - **Git Branch**: `<gh-branch>`
2. Go to **Extensions > Available > AI Assistant** and install a version.

Reload the page to see the "Ask Liz" button in the top-right corner. Send a
Rancher-related prompt, e.g. "list pods in my local cluster", to confirm
everything is working.

## 3. Install the harvester-mcp-server chart

### Clone the repo

```bash
git clone https://github.com/harvester/harvester-mcp-server
cd harvester-mcp-server
```

### Build and push the image

Until the first release is published, you need to build the image yourself
and push it to a registry you control:

```bash
export TARGET_PLATFORMS=linux/amd64
# edit this to your Docker Hub repo, and make sure you're logged in
export REPO=johndoe
export TAG=dev
export BUILDX_ATTEST_FLAGS=

make
```

This example pushes to `docker.io/johndoe/harvester-mcp-server:dev`.

### Install the Helm chart

```bash
# assume you are still in the harvester-mcp-server repo root

# export KUBECONFIG=/path/to/local/cluster/kubeconfig

cat > values.yaml <<EOF
mcp:
  readOnly: false
  image:
    repository: johndoe/harvester-mcp-server
    tag: dev
    pullPolicy: Always

insecureSkipTls: true

log:
  level: debug
EOF

helm install harvester-mcp-server \
  --namespace cattle-ai-agent-system \
  -f values.yaml charts/harvester-mcp-server
```

You should now see the `harvester-mcp-server` deployment in the namespace:

```bash
$ kubectl get deployments -n cattle-ai-agent-system
NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
harvester-mcp-server   1/1     1            1           2d23h
rancher-ai-agent       1/1     1            1           18m
rancher-mcp-server     1/1     1            1           18m
```

## 4. Add the Harvester AI Agent config

```bash
# assume you are still in the harvester-mcp-server repo root
# export KUBECONFIG=/path/to/local/cluster/kubeconfig
kubectl apply -f aiagentconfig.yaml
```

## Restart the rancher-ai-agent pod

You can do one of the following steps to reload the rancher-ai-agent pod:
- Rancher GUI: **Global Settings > AI Assistant**, click the **Apply button**.
- Rollout the `rancher-ai-agent` deployment:

  ```bash
  kubectl rollout restart deployment/rancher-ai-agent -n cattle-ai-agent-system
  ```