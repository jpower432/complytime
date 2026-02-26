# Research: Comply-Packs

**Branch**: `002-comply-packs` | **Date**: 2026-02-25

## R1: OCI Artifact Media Types for Pack Layers

**Decision**: Custom media types following OCI Artifacts specification and existing Gemara convention in `consts.go`.

| Layer | Media Type | Description |
|:---|:---|:---|
| Pack manifest | `application/vnd.complytime.pack.manifest.v1+yaml` | `complypack.yaml` |
| Example config | `application/vnd.complytime.pack.config.v1+yaml` | `complytime.yaml.example` |
| Provider binary | `application/vnd.complytime.pack.provider.v1` | Executable binary |
| Policy OCI layout | `application/vnd.complytime.pack.policy-layout.v1.tar+gzip` | Per-policy OCI Layout dir, tar+gzip compressed |

**Rationale**: Custom media types enable `complyctl pack install` to identify and route each layer to the correct extraction path. Follows the same `application/vnd.{org}.{type}.v1+{format}` pattern used by Gemara media types (`application/vnd.gemara.catalog.v1+yaml`). OCI spec requires media types on all layers.

**Alternatives considered**:
- Generic `application/octet-stream` — rejected: no layer identification without annotations
- Gemara media types — rejected: these are content types (catalog, guidance, policy), not pack distribution types

**Layer annotations**:
- Provider binaries: `org.opencontainers.image.title` = `bin/{binary-name}` (extraction path)
- Policy layouts: `org.opencontainers.image.title` = `policies/{policy-id}.tar.gz` (extraction path)
- Manifest/config: `org.opencontainers.image.title` = filename (`complypack.yaml`, `complytime.yaml.example`)

## R2: Multi-Architecture Provider Binary Strategy

**Decision**: Defer multi-arch to a future iteration. Each pack targets a single OS/architecture pair. The `platform.os` field in `complypack.yaml` documents the target. Build separate packs for separate platforms.

**Rationale**: Pack builder runs on the host platform and bundles host-native binaries. Cross-compilation adds complexity (CGO dependencies for OpenSCAP). Multi-arch OCI Index support can be added later without breaking the single-manifest format — OCI Index wraps manifests, so single-manifest packs are forward-compatible.

**Alternatives considered**:
- OCI Index with platform-specific manifests now — rejected: premature complexity. No immediate multi-arch requirement
- Fat binary approach — rejected: Go cross-compilation is simple but provider binaries may have C dependencies (OpenSCAP via CGO)

## R3: Pack OCI Artifact Structure

**Decision**: Single OCI manifest with typed layers. Each layer identified by its media type and `org.opencontainers.image.title` annotation.

**Structure**:
```text
OCI Manifest
├── Config: application/vnd.oci.empty.v1+json (empty config, standard OCI practice)
├── Layer 1: complypack.yaml (pack manifest)
├── Layer 2: complytime.yaml.example (generated example config)
├── Layer 3: bin/complyctl-provider-openscap (provider binary)
├── Layer N: policies/{policy-id}.tar.gz (OCI Layout, one per policy)
└── Annotations:
    └── org.opencontainers.image.source: <repo-url>
```

**Rationale**: Single manifest is the simplest OCI artifact structure. `oras-go/v2` has first-class support for pushing/pulling multi-layer manifests. Each layer's media type enables routing during extraction. Annotations provide human-readable extraction paths. Empty config blob is standard OCI practice for non-container artifacts (used by Helm, Flux, etc.).

**Alternatives considered**:
- Nested OCI Image Index — rejected: adds indirection without benefit for single-platform packs
- Separate OCI artifacts per component — rejected: defeats bundling purpose, requires multiple pull operations

## R4: `complypack` CLI Framework

**Decision**: Use `cobra` for consistency with `complyctl`. Same patterns, same developer experience.

**Rationale**: `cobra` is already a dependency. Platform engineers familiar with `complyctl` will recognize the CLI structure. Shared patterns reduce cognitive load for contributors working across both binaries. `complypack` has 4 commands (`build`, `push`, `pull`, `doctor`) — `cobra` is lightweight enough for this.

**Alternatives considered**:
- `flag` stdlib only — rejected: loses subcommand structure, help generation, completion
- `urfave/cli` — rejected: introduces a second CLI framework into the repo. Constitution V: don't add unnecessary dependencies

## R5: `COMPLYTIME_PROVIDER_DIR` Implementation

**Decision**: Single env var check in `ResolvePluginDir()` (`internal/complytime/consts.go`). If `COMPLYTIME_PROVIDER_DIR` is set and non-empty, return that path (expanding `~/`). Otherwise, fall back to `~/.complytime/providers/`.

**Rationale**: Centralizes the override in the function all callers already use. Zero changes to `pkg/plugin/discovery.go` — it receives the resolved path from the caller. Constitution I: single source of truth for provider directory resolution.

**Alternatives considered**:
- Per-command `--provider-dir` flag — rejected: Constitution VII (convention over configuration). Env var is sufficient and avoids flag proliferation
- Config file field in `complytime.yaml` — rejected: provider directory is a runtime concern, not a policy concern. Env var is the standard mechanism for path overrides

## R6: Pack Build → Example Config Generation

**Decision**: `complypack build` reads `PackManifest`, copies `PackPolicyEntry.URL` directly into `PolicyEntry.URL`, and writes `complytime.yaml.example` using `SaveTo()` from `config.go`.

**Mapping logic**:
```text
For each PackPolicyEntry in manifest.Policies:
  PolicyEntry{URL: entry.URL, ID: entry.ID}
```

No URL synthesis needed — each `PackPolicyEntry` already carries the full OCI reference. This eliminates the `Registry.URL + "/" + Policy.ID + "@" + Policy.Version` construction that assumed a single registry.

**Rationale**: Reuses existing `WorkspaceConfig` serialization — no custom YAML generation. Direct URL copy means multi-registry packs produce correct example configs without any transformation logic. Target section uses a placeholder `local` target referencing all policies. Global variables include `workspace: ./.complytime/scan` (standard default).

## R7: Workspace-Local Install Layout

**Decision**: `complyctl pack install <ref>` extracts to the current working directory:

```text
./
├── bin/
│   └── complyctl-provider-*     # Provider binaries (chmod +x)
├── policies/
│   └── {policy-id}/             # OCI Layout dirs (untarred from layers)
├── complypack.yaml              # Pack manifest (for doctor)
└── complytime.yaml.example      # Starter config (cp to complytime.yaml)
```

With `-g`:
```text
~/.complytime/
├── providers/
│   └── complyctl-provider-*     # Provider binaries (chmod +x)
└── policies/
    └── {policy-id}/             # OCI Layout dirs
# complypack.yaml and complytime.yaml.example still go to ./
```

**Rationale**: Workspace-local is the default because it's explicit and self-contained — the entire compliance setup lives alongside the code being assessed. Global install (`-g`) puts providers and policies in the standard `complyctl` discovery paths, eliminating the need for `COMPLYTIME_PROVIDER_DIR`. Both modes place `complypack.yaml` and `complytime.yaml.example` in the workspace — these are user-facing files the admin needs to see and interact with.

## R8: Doctor Pack Checks

**Decision**: When `complypack.yaml` exists in the workspace, `complyctl doctor` runs additional pack-layer checks after standard config checks:

| Check | Blocking | Description |
|:---|:---|:---|
| Pack manifest valid | Yes | Parse + validate `complypack.yaml` via `LoadPackManifest()` + `ValidatePackManifest()` |
| Provider binaries present | Yes | For each `PackProviderEntry`, verify binary exists at expected path (workspace `./bin/` or global `~/.complytime/providers/`) |
| Policy caches present | Yes | For each `PackPolicyEntry`, verify OCI Layout dir exists at expected path |
| System dependencies | No (warn) | For each `SystemDependency`, run `check` command. Fail = warn with `install` guidance |

**Rationale**: Pack checks validate the delivery mechanism. Config checks validate the runtime. Both are needed for a complete pre-flight diagnostic. System dependency checks are warnings because the admin may have installed packages via alternative methods.

**Missing config handling**: If `complypack.yaml` present but `complytime.yaml` absent, doctor reports pack OK + config missing with `cp complytime.yaml.example complytime.yaml` guidance. This is a warning, not a blocking error — the admin may be in the process of setting up.
