## Harvester MCP Server

The MCP server allows the [Rancher AI agent](https://github.com/rancher-sandbox/rancher-ai-agent) to securely retrieve or update Harvester resources in Harvester clusters. It expects the Rancher token in a header, which the agent will always provide for authentication.

> [!NOTE]
> This repo was originally forked from https://github.com/rancher/rancher-ai-mcp and is now dedicated to Harvester / SUSE Virtualization.

## Overview

This Model Context Protocol (MCP) server provides a secure bridge between the Rancher AI agent and Harvester clusters, enabling AI-powered cluster management through a standardized tool interface. The server runs as a Kubernetes deployment within the Rancher environment and exposes tools for resource inspection, modification, and cluster operations.

## Architecture

### Package Structure

- **`cmd/`** - CLI commands and server initialization
  - `serve.go` - HTTP/TLS server setup with dynamic listener support
  - `root.go` - Root command configuration

- **`pkg/client/`** - Kubernetes client abstraction
  - Dynamic client wrapper with cluster ID resolution
  - Rancher API integration for cluster management
  - Support for both local and downstream cluster operations

- **`pkg/toolsets/virtualization/`** - Harvester / SUSE Virtualization tools
  - `toolsets.go` - Central registry for tool collections
  - VM, image, volume, network, and node operations

- **`pkg/response/`** - Response formatting utilities
  - Structured text and content generation for MCP responses

- **`pkg/converter/`** - Data transformation utilities
  - Group/Version/Resource (GVR) conversion helpers

### TLS & Security

The server supports two modes:

1. **TLS Mode (Production)**: Uses Rancher's dynamic listener with auto-generated certificates
   - Certificates stored as Kubernetes secrets
   - Automatic cert rotation and renewal
   - Client certificate authentication support
   - TLS 1.2+ with secure cipher suites

2. **Insecure Mode (Development)**: Plain HTTP for local testing
   - Enabled via `--insecure` flag or `INSECURE_SKIP_TLS=true`

### Available Tools

Each tool is exposed through the MCP protocol and can be invoked by the Rancher AI agent:

| Tool                          | Description                                                                       |
|-------------------------------|------------------------------------------------------------------------------------|
| `listVirtualMachines`         | Lists VirtualMachine resources on a Harvester / SUSE Virtualization cluster        |
| `listVirtualMachineImages`    | Lists VirtualMachineImage resources on a Harvester / SUSE Virtualization cluster   |
| `listVolumes`                 | Lists storage volumes (PersistentVolumeClaims) on a Harvester / SUSE Virtualization cluster |
| `listVirtualMachineNetworks`  | Lists NetworkAttachmentDefinition resources on a Harvester / SUSE Virtualization cluster |
| `listClusterNetworks`         | Lists ClusterNetwork resources on a Harvester / SUSE Virtualization cluster        |
| `listNodes`                   | Lists nodes (hosts) on a Harvester / SUSE Virtualization cluster                   |
| `inspectVolume`               | Inspects a storage volume (PersistentVolumeClaim), including Longhorn details when applicable |
| `createVirtualMachineImage`   | Creates a VirtualMachineImage resource by downloading from a URL                   |
| `createVirtualMachine`        | Creates a VirtualMachine from an existing image                                    |
| `createVirtualMachinePlan`    | Returns the manifest that would be submitted to create a VirtualMachine, without creating it |
| `deleteVirtualMachine`        | Deletes a VirtualMachine                                                            |
| `deleteVirtualMachinePlan`    | Returns the resource that would be deleted, without actually removing it           |

## Configuration

### Command-line Flags

```bash
--port <int>              Port to listen on (default: 9092)
--insecure                Skip TLS verification (default: false)
```
