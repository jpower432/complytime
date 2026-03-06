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
    Describe(ctx context.Context, req *DescribeRequest) (*DescribeResponse, error)
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
    Scan(ctx context.Context, req *ScanRequest) (*ScanResponse, error)
}
```

### Describe

Called during plugin discovery and `complyctl doctor` diagnostics. Returns plugin health, version, and declared variable requirements (`RequiredGlobalVariables`, `RequiredTargetVariables`). Plugins that return `Healthy: false` or fail the RPC are skipped during loading.

### Generate

Called by `complyctl generate`. Receives a three-tier variable model (R48):

| Tier | Field | Source |
|:---|:---|:---|
| 1 — Global | `GlobalVariables` | `complytime.yaml` top-level `variables` |
| 2 — Target | `TargetVariables` | `complytime.yaml` `targets[].variables` (one target per call) |
| 3 — Test | `Configuration[].Parameters` | Per-requirement parameters from the assessment plan |

The plugin prepares declarative policies in whatever format the underlying engine expects.

### Scan

Called by `complyctl scan`. Receives targets only — no requirement IDs are sent. The provider evaluates all requirements from Generate-time state (R47). Returns `AssessmentLog` entries — one per requirement evaluated — each containing steps with pass/fail/skip/error results and a `ConfidenceLevel` enum.

## Protobuf Contract

The canonical protobuf definition lives at `api/plugin/plugin.proto`. Key types:

| Type | Purpose |
|:---|:---|
| `GenerateRequest` | Global variables, target variables, assessment configurations |
| `AssessmentConfiguration` | Plan ID, requirement ID, parameters map |
| `Target` | Target ID + plugin-defined variables |
| `AssessmentLog` | Requirement ID, steps, message, confidence level |
| `Step` | Name, result, message |
| `DescribeResponse` | Health, version, required global/target variable names |
| `ConfidenceLevel` | Enum: NOT_SET, UNDETERMINED, LOW, MEDIUM, HIGH |
| `Result` | Enum: UNSPECIFIED, PASSED, FAILED, SKIPPED, ERROR |

## Plugin Commands

Plugins support an optional development command:

- **Default behavior**: Starts the gRPC server for runtime policy execution (handles `Describe`, `Generate`, `Scan` RPCs)
- **`init --tool <agent>`**: Generates AI agent integration artifacts for the specified tool (cursor, opencode, claude-code)

Plugins use `plugin.RegisterInit()` to register the init command handler. If no command is provided or the command is unrecognized, the plugin serves by default.

### Init Command Registration

Plugins register the init command using `plugin.RegisterInit()` with templates:

```go
if plugin.RegisterInit(plugin.InitTemplates{
    CommandTemplate: commandTemplate,  // Markdown template for command definition
    SkillTemplate:   skillTemplate,      // Markdown template for skill definition
    CommandName:     "my-command",       // Kebab-case command name
    SkillName:       "my-skill",         // Optional: defaults to CommandName
}) {
    return  // Command handled, exit
}
```

### Init Command Installation

The `init --tool <agent>` command installs skill definitions directly to the correct locations:

- **Cursor**: Writes to `.cursor/commands/{command-name}.md` and `.cursor/skills/{skill-name}/SKILL.md`
- **OpenCode**: Appends to `.cursorrules` in workspace root
- **Claude Code**: Writes to `.claude/{command-name}-tool.md`

The command automatically detects the workspace root (by finding `.git` directory or using current directory) and installs files accordingly.

## Authoring a Plugin (Go)

Plugins serve by default. Use `plugin.Serve()` to register and start the gRPC server. The handshake is handled automatically. Optionally, plugins can register an `init` command to set up development tooling.

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/complytime/complyctl/pkg/plugin"
)

var _ plugin.Plugin = (*myPlugin)(nil)

type myPlugin struct{}

func (p *myPlugin) Describe(_ context.Context, _ *plugin.DescribeRequest) (*plugin.DescribeResponse, error) {
    return &plugin.DescribeResponse{
        Healthy: true,
        Version: "1.0.0",
        RequiredGlobalVariables: []string{"output_dir"},
        RequiredTargetVariables: []string{"kubeconfig"},
    }, nil
}

func (p *myPlugin) Generate(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
    _ = req.GlobalVariables
    _ = req.TargetVariables
    for _, cfg := range req.Configuration {
        _ = cfg.RequirementID
        _ = cfg.Parameters
    }
    return &plugin.GenerateResponse{Success: true}, nil
}

func (p *myPlugin) Scan(_ context.Context, req *plugin.ScanRequest) (*plugin.ScanResponse, error) {
    var assessments []plugin.AssessmentLog
    for _, target := range req.Targets {
        assessments = append(assessments, plugin.AssessmentLog{
            RequirementID: target.TargetID + "-check",
            Steps: []plugin.Step{{
                Name:    "my-check",
                Result:  plugin.ResultPassed,
                Message: "check passed",
            }},
            Message:    "evaluation complete",
            Confidence: plugin.ConfidenceLevelHigh,
        })
    }
    return &plugin.ScanResponse{Assessments: assessments}, nil
}

func main() {
    // Register init command handler (optional)
    if plugin.RegisterInit(plugin.InitTemplates{
        CommandTemplate: commandTemplate,  // Embed or load your command template
        SkillTemplate:   skillTemplate,      // Embed or load your skill template
        CommandName:     "my-command",
    }) {
        return  // Command handled, exit
    }

    // Default behavior: serve the plugin
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

Plugins receive variables through a three-tier model (R48):

| Tier | Config Location | Delivered Via | Scope |
|:---|:---|:---|:---|
| Global | `complytime.yaml` `variables` | `GenerateRequest.GlobalVariables` | Workspace-wide |
| Target | `complytime.yaml` `targets[].variables` | `GenerateRequest.TargetVariables` (Generate) / `Target.Variables` (Scan) | Per-target |
| Test | Assessment plan parameters | `AssessmentConfiguration.Parameters` | Per-requirement |

Plugins declare their required variable names via the `Describe` RPC (`RequiredGlobalVariables`, `RequiredTargetVariables`). The `complyctl doctor` command validates these exist in the workspace config (R51).

```yaml
variables:
  output_dir: /tmp/scan-results

targets:
  - id: production-cluster
    policies:
      - nist-800-53-r5
    variables:
      kubeconfig: /path/to/kubeconfig
      namespace: default
```

## Reference Implementation

See `cmd/test-plugin/main.go` for a complete working example. See `cmd/ampel-plugin/main.go` for an example implementing the default serve behavior and optional `init` command. See `docs/PLUGIN_TEMPLATE.md` for a template showing the command structure.

Build with:

```bash
make build-test-plugin
```
