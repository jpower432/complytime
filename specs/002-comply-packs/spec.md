# Feature Specification: Comply-Packs

**Feature Branch**: `002-comply-packs`
**Created**: 2026-02-25
**Status**: Draft
**Prerequisite**: `001-gemara-native-workflow` (complete)
**Input**: Platform engineers need a way to bundle curated policies, scanning providers, and starter configs into a distributable artifact so admins can get up and running quickly — including in airgapped environments.

## Clarifications

### Session 2026-02-25

- Q: Who builds packs? → A: Platform engineers who curate policy sets for their org. Separate persona from the admin/app owner who runs scans.
- Q: What problem does a pack solve? → A: Bundles policies + supporting plugins + example config for rapid onboarding. Also enables offline transfer. Pack is a **delivery mechanism**, not a runtime mode.
- Q: Should the consumer's `complytime.yaml` look identical with or without a pack? → A: Yes. Pack generates a standard `complytime.yaml.example` using the `PolicyEntry{url, id}` format from 001. No pack-awareness in `complyctl` runtime.
- Q: Can a pack pull policies from multiple registries? → A: Yes. Each `PackPolicyEntry` carries its own `url` (full OCI reference). No top-level `registry` field — the pack builder parses each URL to determine which registry to pull from. Generated `complytime.yaml.example` copies URLs directly.
- Q: Tarball or OCI artifact for distribution? → A: OCI-only. Tools like `skopeo` handle airgap scenarios (copy between registries, export to tarball for USB transfer). No custom tarball format to maintain.
- Q: Should the pack builder be part of `complyctl`? → A: No. `complypack` is a standalone binary (like `ocb` for the OpenTelemetry Collector). Lives in this repo at `cmd/complypack/` to share types from `internal/complytime/pack.go`.
- Q: Where does `complyctl pack install` put files? → A: Defaults to workspace-local (`./bin/`, `./policies/`, `./complytime.yaml.example`). `-g` flag installs to global paths (`~/.complytime/providers/`, `~/.complytime/policies/`).

## Scope

### In Scope

- **`complypack` CLI binary** (`cmd/complypack/`): `build`, `push`, `pull`, `doctor` subcommands for platform engineers
- **`complyctl pack install`** subcommand: Unpacks an OCI pack artifact into the workspace (or global paths with `-g`)
- **`COMPLYTIME_PROVIDER_DIR`** env var: Overrides default provider discovery path so workspace-local `./bin/` providers are found
- **OCI artifact format**: Pack distributed as a single OCI artifact containing provider binaries, pre-cached policy OCI layouts, `complypack.yaml` manifest, and `complytime.yaml.example`
- **Doctor dual-file mode**: When `complypack.yaml` is present alongside `complytime.yaml`, `complyctl doctor` validates both

### Out of Scope

- Custom tarball format (OCI handles this via `skopeo copy` / `oras` for airgap)
- Pack runtime mode in `complyctl` (no `pack` field in `WorkspaceConfig`)
- Auto-update or version management for installed packs
- Pack signing or verification (defer to OCI signing standards like cosign)

## User Scenarios & Testing

### User Story 1 — Build and Publish a Comply-Pack (Priority: P1)

A platform engineer curates a set of Gemara policies and scanning providers for their org's Fedora workstations. They build a comply-pack and publish it to the org's OCI registry so admins can install it.

**Why this priority**: Without a buildable, publishable pack, the entire distribution pipeline doesn't exist.

**Independent Test**: Run `complypack build` in a directory with `complypack.yaml`, verify the OCI artifact is assembled. Run `complypack push` to a test registry, verify the artifact is retrievable.

**Acceptance Scenarios**:

1. **Given** a directory with a valid `complypack.yaml` and access to the registries referenced in policy URLs, **When** a platform engineer runs `complypack build`, **Then** the system fetches policies from their respective registries into local OCI layouts, locates provider binaries, generates `complytime.yaml.example` with `PolicyEntry` URLs copied from manifest entries, and assembles the pack as an OCI artifact
2. **Given** a built pack artifact, **When** the engineer runs `complypack push <registry>/<name>:<tag>`, **Then** the pack is pushed to the OCI registry as a standard OCI artifact
3. **Given** a `complypack.yaml` with missing provider binaries or unreachable registry, **When** the engineer runs `complypack doctor`, **Then** the system reports which components are missing or unreachable with actionable remediation
4. **Given** a published pack, **When** the engineer runs `complypack pull <registry>/<name>:<tag>`, **Then** the pack artifact is pulled to a local directory for inspection or offline transfer

---

### User Story 2 — Install and Use a Comply-Pack (Priority: P2)

An admin receives an OCI pack reference from their platform engineering team. They install it into their workspace, customize the example config with their targets, and run a compliance scan.

**Why this priority**: This is the consumer-side workflow that delivers the "up and running quickly" value proposition.

**Independent Test**: Run `complyctl pack install <oci-ref>`, verify `./bin/`, `./policies/`, and `./complytime.yaml.example` are populated. Copy example to `complytime.yaml`, edit targets, run `complyctl doctor` then `complyctl scan --policy-id <ID>`.

**Acceptance Scenarios**:

1. **Given** an OCI pack reference, **When** an admin runs `complyctl pack install <oci-ref>`, **Then** the system pulls the pack artifact and extracts: provider binaries to `./bin/`, pre-cached policy OCI layouts to `./policies/`, `complypack.yaml` to `./`, and `complytime.yaml.example` to `./`
2. **Given** a workspace-local install, **When** the admin sets `COMPLYTIME_PROVIDER_DIR=./bin/` and runs `complyctl providers`, **Then** the bundled providers are discovered and listed
3. **Given** an OCI pack reference and the `-g` flag, **When** the admin runs `complyctl pack install -g <oci-ref>`, **Then** provider binaries install to `~/.complytime/providers/` and policy caches install to `~/.complytime/policies/`. No `COMPLYTIME_PROVIDER_DIR` override needed
4. **Given** an installed pack with `complypack.yaml` and a customized `complytime.yaml`, **When** the admin runs `complyctl doctor`, **Then** the system validates both files: pack integrity (provider binaries present, policy caches intact, system deps installed) and config validity (target-policy cross-refs, variable resolution)
5. **Given** a complete install, **When** the admin copies `complytime.yaml.example` to `complytime.yaml`, edits targets/variables, and runs `complyctl scan --policy-id <ID>`, **Then** the scan executes using bundled providers and pre-cached policies with zero registry access required

---

### User Story 3 — Airgap Transfer (Priority: P3)

A platform engineer needs to transfer a comply-pack to an airgapped environment with no internet access.

**Why this priority**: Airgap is a key differentiator, but is enabled by OCI tooling (skopeo) rather than custom code. Low implementation cost.

**Independent Test**: Use `skopeo copy` to export pack to a tarball, transfer to airgapped host, `skopeo copy` to local registry, `complyctl pack install` from local registry.

**Acceptance Scenarios**:

1. **Given** a published pack, **When** the engineer runs `skopeo copy docker://<src> oci-archive:/pack.tar`, transfers the tarball to an airgapped host, and runs `skopeo copy oci-archive:/pack.tar docker://<local-registry>/<name>:<tag>`, **Then** the pack is available in the local registry
2. **Given** the pack in a local registry, **When** an admin runs `complyctl pack install <local-registry>/<name>:<tag>`, **Then** the install proceeds identically to a connected install

---

### Edge Cases

| Condition | Expected Behavior |
|:---|:---|
| `complypack build` with missing provider binary | Error listing missing binary path + remediation |
| `complypack build` with unreachable policy registry | Error listing unreachable URL + suggestion to check credentials |
| `complyctl pack install` with existing `./bin/` or `./policies/` | Prompt for overwrite confirmation or use `--force` flag |
| `complyctl pack install` where OCI ref is not a valid pack artifact | Error: "not a valid comply-pack artifact" with expected media type |
| `complyctl doctor` with `complypack.yaml` but missing bundled providers | Report missing binary paths from manifest with install remediation |
| `complyctl doctor` with `complypack.yaml` but no `complytime.yaml` | Report: pack installed, config missing, suggest `cp complytime.yaml.example complytime.yaml` |
| Pack policy ID not in consumer's `complytime.yaml` targets | Not an error — consumer selects subset of bundled policies |
| System dependency check fails | Doctor reports failed check command + suggests install command from manifest |

## Requirements

### Functional Requirements

- **FR-201**: `complypack build` MUST read `complypack.yaml`, fetch all declared policies from their respective registries (each `PackPolicyEntry.url` is a full OCI reference) into OCI Layout directories, locate declared provider binaries, generate `complytime.yaml.example` with `PolicyEntry` URLs copied directly from manifest entries, and assemble the result as an OCI artifact
- **FR-202**: `complypack push <ref>` MUST push the built OCI artifact to the specified registry reference using standard OCI distribution. Authentication MUST use `oras-credentials-go` (same as `complyctl`)
- **FR-203**: `complypack pull <ref>` MUST pull the OCI artifact to a local directory for inspection or offline handoff
- **FR-204**: `complypack doctor` MUST validate the `complypack.yaml` for buildability: registry reachable, policies exist remotely, declared provider binaries exist on the local filesystem, system dependency `check` commands pass
- **FR-205**: `complyctl pack install <oci-ref>` MUST pull the pack artifact and extract contents to the workspace: provider binaries to `./bin/`, policy OCI layouts to `./policies/`, `complypack.yaml` and `complytime.yaml.example` to `./`. With `-g` flag, install providers to `~/.complytime/providers/` and policies to `~/.complytime/policies/`
- **FR-206**: `complyctl` MUST support `COMPLYTIME_PROVIDER_DIR` environment variable to override the default provider discovery path (`~/.complytime/providers/`). When set, provider discovery scans the specified directory instead
- **FR-207**: `complyctl doctor` MUST detect `complypack.yaml` presence and validate pack integrity alongside `complytime.yaml` config validity. Pack checks: provider binaries exist at expected paths, policy OCI layouts exist, system dependency `check` commands pass. If `complytime.yaml` absent but `complypack.yaml` present, report pack OK + config missing with `cp complytime.yaml.example complytime.yaml` guidance
- **FR-208**: `complytime.yaml.example` generated by `complypack build` MUST use the standard `PolicyEntry` format from 001 — each policy as `url:` (copied from `PackPolicyEntry.URL`) + `id:` (from `PackPolicyEntry.ID`). Targets section MUST be a placeholder with a `local` target referencing all pack policies. Variables section MUST include `workspace: ./.complytime/scan`

### Non-Functional Requirements

- **NFR-201**: Zero changes to `complyctl` runtime behavior for non-pack users. Pack support is additive
- **NFR-202**: `complypack` binary MUST share types from `internal/complytime/pack.go` — no type duplication
- **NFR-203**: OCI artifact format MUST be pullable by standard OCI tools (`skopeo`, `oras`, `crane`)

### Key Entities

- **Pack Artifact**: OCI artifact containing layers for provider binaries, policy OCI layouts, `complypack.yaml`, and `complytime.yaml.example`. Custom media types for layer identification
- **Pack Manifest** (`complypack.yaml`): `PackManifest` struct from `internal/complytime/pack.go`. Each `PackPolicyEntry` carries its own `url` (full OCI reference) enabling multi-registry packs. No top-level `registry` field
- **Example Config** (`complytime.yaml.example`): Standard `WorkspaceConfig` YAML with `PolicyEntry` URLs copied directly from pack manifest entries

## Success Criteria

- **SC-201**: Platform engineer can build and push a pack in under 5 minutes for a typical pack (5 policies, 1 provider)
- **SC-202**: Admin can install a pack and run a scan in under 3 minutes (excluding policy content download time for global install)
- **SC-203**: Airgap transfer works with standard `skopeo` commands — no custom tooling required
- **SC-204**: Existing 001 workflows (no pack) are completely unaffected — zero behavioral changes
