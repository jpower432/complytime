# Plugin Guide

complyctl extends to arbitrary policy engines through plugins. Each plugin is a standalone executable that communicates with the CLI via gRPC using the [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin) framework.

## Discovery

Scanning providers are discovered by scanning `~/.complytime/providers/` for executables matching the naming convention:

```
complyctl-provider-<evaluator-id>
```

The CLI strips the `complyctl-provider-` prefix to derive the **evaluator ID** used for routing Generate and Scan requests.

| Example Binary | Evaluator ID |
|:---|:---|
| `complyctl-provider-openscap` | `openscap` |
| `complyctl-provider-kubernetes` | `kubernetes` |
| `complyctl-provider-test` | `test` |

No manifest files, no configuration files. The executable must be in the plugin directory and have execute permission.

## gRPC Interface

Plugins implement the `Plugin` interface (defined in `pkg/plugin/manager.go`):

```go
type Plugin interface {
    HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error)
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
    Scan(ctx context.Context, req *ScanRequest) (*ScanResponse, error)
}
```

### HealthCheck

Called during plugin discovery. Plugins that return `Healthy: false` or fail the RPC are skipped.

### Generate

Called by `complyctl generate`. Receives assessment configurations (plan ID, requirement ID, parameters) for the plugin to prepare declarative policies in whatever format the underlying engine expects.

### Scan

Called by `complyctl scan`. Receives targets (with variables) and requirement IDs. Returns `AssessmentLog` entries — one per requirement evaluated — each containing steps with pass/fail/skip/error results and a confidence score.

## Protobuf Contract

The full protobuf definition lives at `specs/001-gemara-native-workflow/contracts/plugin.proto`. Key messages:

| Message | Purpose |
|:---|:---|
| `AssessmentConfiguration` | Plan ID, requirement ID, parameters map |
| `Target` | Target ID + plugin-defined variables |
| `AssessmentLog` | Requirement ID, steps, message, confidence |
| `Step` | Name, result (PASSED/FAILED/SKIPPED/ERROR), message |

## Authoring a Plugin (Go)

Use `plugin.Serve()` to register and start the gRPC server. The handshake is handled automatically.

```go
package main

import (
    "context"

    "github.com/complytime/complyctl/pkg/plugin"
)

var _ plugin.Plugin = (*myPlugin)(nil)

type myPlugin struct{}

func (p *myPlugin) HealthCheck(_ context.Context, _ *plugin.HealthCheckRequest) (*plugin.HealthCheckResponse, error) {
    return &plugin.HealthCheckResponse{Healthy: true, Version: "1.0.0"}, nil
}

func (p *myPlugin) Generate(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
    for _, cfg := range req.Configuration {
        // Translate assessment config into engine-native format
        _ = cfg.RequirementID
        _ = cfg.Parameters
    }
    return &plugin.GenerateResponse{Success: true}, nil
}

func (p *myPlugin) Scan(_ context.Context, req *plugin.ScanRequest) (*plugin.ScanResponse, error) {
    assessments := make([]plugin.AssessmentLog, 0, len(req.RequirementIDs))
    for _, reqID := range req.RequirementIDs {
        assessments = append(assessments, plugin.AssessmentLog{
            RequirementID: reqID,
            Steps: []plugin.Step{{
                Name:    "my-check",
                Result:  plugin.ResultPassed,
                Message: "check passed",
            }},
            Message:    "evaluation complete",
            Confidence: 1.0,
        })
    }
    return &plugin.ScanResponse{Assessments: assessments}, nil
}

func main() {
    plugin.Serve(&myPlugin{})
}
```

Build and install:

```bash
go build -o complyctl-provider-myplugin ./cmd/myplugin
cp complyctl-provider-myplugin ~/.complytime/providers/
```

## Routing

The CLI routes requests based on **evaluator ID** extracted from the Gemara policy graph:

1. Policy assessment configs include an `evaluator_id` field
2. CLI groups configs by evaluator ID
3. Each group is dispatched to the matching plugin
4. If no match is found, the request is broadcast to all loaded plugins

## Variables

Targets in `complytime.yaml` (runtime config) can define `variables` — arbitrary key-value pairs passed to plugins via `Target.Variables` in the Scan RPC. Use these for authentication tokens, connection strings, file paths, or any plugin-specific configuration. In comply-pack environments, the pack manifest (`complypack.yaml`) declares which providers are bundled; the runtime config remains the source for target variables.

```yaml
targets:
  - id: production-cluster
    policy_ids:
      - nist-800-53-r5
    variables:
      kubeconfig: /path/to/kubeconfig
      namespace: default
```

## Reference Implementation

See `cmd/test-plugin/main.go` for a complete working example. Build with:

```bash
make build-test-plugin
```
