# complyctl

[![OpenSSF Best Practices status](https://www.bestpractices.dev/projects/9761/badge)](https://www.bestpractices.dev/projects/9761)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/complytime/complyctl)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/complytime/complyctl/badge)](https://scorecard.dev/viewer/?uri=github.com/complyctl/complyctl)

A lightweight compliance runtime that pulls [Gemara](https://gemara.openssf.org/) policies from an OCI registry and executes scans via plugins.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Host                                                           │
│                                                                 │
│  ┌──────────────┐      complyctl get   ┌──────────────────────┐ │
│  │  OCI Registry │ ◄────────────────── │                      │ │
│  │              │  ──────────────────► │      complyctl CLI   │ │
│  │  Gemara      │   catalog + policy   │                      │ │
│  │  policies    │   layers (YAML)      │  init / get / list   │ │
│  └──────────────┘                      │  generate / scan     │ │
│                                        │  version             │ │
│                                        └──────┬───────┬───────┘ │
│                                               │       │         │
│                                  ┌────────────┘       │         │
│                                  │                    │         │
│                                  ▼                    ▼         │
│                        ┌──────────────┐    ┌────────────────┐   │
│                        │    Cache     │    │    Plugins     │   │
│                        │              │    │                │   │
│                        │ ~/.complytime│    │ ~/.complytime/ │   │
│                        │  policies/   │    │  plugins/      │   │
│                        │  state.json  │    │                │   │
│                        │              │    │ complytime-    │   │
│                        │ OCI Layout   │    │  plugin-*      │   │
│                        │ per policy   │    │                │   │
│                        └──────────────┘    │ gRPC: Health,  │   │
│                                            │ Generate, Scan │   │
│  ┌──────────────┐                          └────────────────┘   │
│  │  Workspace   │                                               │
│  │              │  complytime.yaml defines:                     │
│  │  ./complytime│   - registry URL                              │
│  │    .yaml     │   - policy IDs + versions                     │
│  │              │   - targets + variables                       │
│  │  ./gemara-   │   - parameter overrides                       │
│  │    scan/     │                                               │
│  │   (output)   │  Scan output (EvaluationLog, OSCAL,           │
│  └──────────────┘   SARIF, Markdown) written to workspace       │
└─────────────────────────────────────────────────────────────────┘
```

**Components:**

| Component | Description |
|:---|:---|
| **OCI Registry** | Remote store for Gemara policies. Each policy is a multi-layer OCI manifest containing catalog, guidance, and policy/assessment YAML layers distinguished by media type. |
| **Workspace** | Current directory containing `complytime.yaml`. Defines which registry, policies, and targets to use. Scan output lands in `./.complytime/scan/`. |
| **Cache** | Local OCI Layout stores under `~/.complytime/policies/`. One store per policy ID. `state.json` tracks digests for incremental sync. |
| **Providers** | Standalone executables in `~/.complytime/providers/` matching the `complyctl-provider-*` naming convention. Communicate via gRPC (`HealthCheck`, `Generate`, `Scan`). Evaluator ID derived from filename. |
| **CLI** | Orchestrates the workflow: fetch policies, resolve dependency graphs, dispatch to plugins, produce compliance reports. |

## Documentation

- [Installation](./docs/INSTALLATION.md)
- [Quick Start](./docs/QUICK_START.md)
- [Plugin Guide](./docs/PLUGIN_GUIDE.md)
- [E2E Testing](./docs/E2E_INTEGRATION.md)

## CLI Commands

| Command | Description |
|:---|:---|
| `init` | Initialize workspace and fetch policies from OCI registry |
| `get` | Fetch new/modified policies from OCI registry and update cache |
| `list` | List cached Gemara policies |
| `generate` | Generate policy graph and invoke plugins |
| `scan` | Scan targets and produce compliance reports |
| `version` | Print version |

Global flag: `--debug` / `-d` — output debug logs.

### `init`

```bash
# Interactive — prompts for registry URL, policy IDs, targets
complyctl init

# Non-interactive — validates existing complytime.yaml, then fetches policies
complyctl init
```

When `complytime.yaml` already exists, `init` validates and runs `get` automatically.

### `get`

```bash
complyctl get
```

Performs incremental sync from the OCI registry defined in `complytime.yaml`. Only downloads new or modified content. Uses Docker credential helpers for authentication — if `docker login` works, `complyctl get` works.

### `list`

```bash
complyctl list
complyctl list --plain
complyctl list --policy-id nist-800-53-r5
```

| Flag | Description |
|:---|:---|
| `--plain` | Minimal formatting (no interactive table) |
| `--policy-id` | Filter output to a single policy |

### `generate`

```bash
complyctl generate --policy-id nist-800-53-r5
```

| Flag | Short | Description |
|:---|:---|:---|
| `--policy-id` | `-p` | Policy ID to generate (required) |

Resolves the policy dependency graph from cache, extracts assessment configurations, applies parameter overrides from `complytime.yaml`, and dispatches to the matching plugin via Generate RPC.

### `scan`

```bash
# Default: EvaluationLog only
complyctl scan --policy-id nist-800-53-r5

# OSCAL assessment-results
complyctl scan --policy-id nist-800-53-r5 --format oscal

# Markdown report
complyctl scan --policy-id nist-800-53-r5 --format pretty

# SARIF for security tooling
complyctl scan --policy-id nist-800-53-r5 --format sarif
```

| Flag | Short | Description |
|:---|:---|:---|
| `--policy-id` | `-p` | Policy ID to scan (required) |
| `--format` | `-f` | Output format: `oscal`, `pretty`, `sarif` |

Output written to `./.complytime/scan/`.

## Workspace Configuration

```yaml
# complytime.yaml
registry:
  url: https://registry.example.com
policies:
  - id: nist-800-53-r5
  - id: cis-benchmark
    version: v1.0.0
targets:
  - id: production-cluster
    policy_ids:
      - nist-800-53-r5
    variables:
      kubeconfig: /path/to/kubeconfig
parameters:
  - policy_id: nist-800-53-r5
    parameter_id: session_timeout
    value: "900"
```

| Field | Description |
|:---|:---|
| `registry.url` | OCI registry base URL |
| `policies[].id` | Policy identifier in the registry |
| `policies[].version` | Optional pinned version (omit for latest) |
| `targets[].id` | Scan target identifier |
| `targets[].policy_ids` | Policies to evaluate against this target |
| `targets[].variables` | Plugin-specific key-value pairs (credentials, paths) |
| `parameters[]` | Override policy-defined parameter values locally |

## Contributing

- [Contributing Guidelines](./docs/CONTRIBUTING.md)
- [Style Guide](./docs/STYLE_GUIDE.md)
- [Code of Conduct](./docs/CODE_OF_CONDUCT.md)

*Interested in writing a plugin?* See the [Plugin Guide](./docs/PLUGIN_GUIDE.md).
