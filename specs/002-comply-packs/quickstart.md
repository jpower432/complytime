# Quickstart: Comply-Packs

**Branch**: `002-comply-packs` | **Prerequisite**: `001-gemara-native-workflow`

## Personas

| Persona | Tool | Goal |
|:---|:---|:---|
| Platform Engineer | `complypack` | Build, validate, and publish comply-packs |
| Admin / App Owner | `complyctl` | Install a pack and run compliance scans |

---

## Part 1 — Platform Engineer: Build and Publish

### 1a. Create a Pack Manifest

Create `complypack.yaml` in your working directory:

```yaml
id: fedora-compliance
version: "1.0.0"
description: CIS and CUSP policies for Fedora workstations
platform:
  os: linux
  arch: amd64
policies:
  - url: registry.example.com/compliance/cis-fedora-l1-workstation@1.2.3
    id: cis-fedora-l1-workstation
    profile: cis_workstation_l1
    catalog: cis-benchmarks
  - url: public-registry.io/community/cusp-fedora-default@1.0.0
    id: cusp-fedora-default
    profile: cusp_fedora
    catalog: cusp
providers:
  - id: openscap
    binary: complyctl-provider-openscap
system-dependencies:
  - name: openscap-scanner
    kind: rpm
    value: openscap-scanner
    url: https://www.open-scap.org/tools/openscap-base/
```

### 1b. Validate Buildability

```bash
complypack doctor
```

Expected output:
```text
Pack manifest valid
Provider binary found: complyctl-provider-openscap
Registry reachable: registry.example.com
Registry reachable: public-registry.io
Policy exists: cis-fedora-l1-workstation@1.2.3
Policy exists: cusp-fedora-default@1.0.0
System dependency: openscap-scanner (rpm)
```

### 1c. Build the Pack

```bash
complypack build
```

This fetches policies from the registry into OCI Layout directories, locates provider binaries, generates `complytime.yaml.example`, and assembles the OCI artifact.

### 1d. Publish

```bash
complypack push registry.example.com/packs/fedora-compliance:1.0.0
```

### 1e. Airgap Transfer (optional)

```bash
# Connected host: export to tarball
skopeo copy \
  docker://registry.example.com/packs/fedora-compliance:1.0.0 \
  oci-archive:/tmp/fedora-compliance-1.0.0.tar

# Transfer /tmp/fedora-compliance-1.0.0.tar to airgapped host

# Airgapped host: import to local registry
skopeo copy \
  oci-archive:/tmp/fedora-compliance-1.0.0.tar \
  docker://localhost:5000/packs/fedora-compliance:1.0.0
```

---

## Part 2 — Admin: Install and Scan

### 2a. Install the Pack (Workspace-Local)

```bash
mkdir my-compliance && cd my-compliance
complyctl pack install registry.example.com/packs/fedora-compliance:1.0.0
```

One pack per workspace. If a pack is already installed, you'll be prompted to overwrite (or use `--force`). Install validates host `os`/`arch` compatibility before extracting. If extraction fails, all written files are rolled back.

Result:
```text
./
├── bin/
│   └── complyctl-provider-openscap
├── policies/
│   ├── cis-fedora-l1-workstation/
│   └── cusp-fedora-default/
├── complypack.yaml
└── complytime.yaml.example
```

### 2b. Configure

```bash
cp complytime.yaml.example complytime.yaml
```

Edit `complytime.yaml` to customize targets and variables:

```yaml
policies:
  - url: registry.example.com/compliance/cis-fedora-l1-workstation@1.2.3
    id: cis-fedora-l1-workstation
  - url: public-registry.io/community/cusp-fedora-default@1.0.0
    id: cusp-fedora-default

variables:
  workspace: ./.complytime/scan

targets:
  - id: local
    policies:
      - cis-fedora-l1-workstation
    variables:
      profile: cis_workstation_l1
```

### 2c. Set Provider Discovery (Workspace-Local Only)

```bash
export COMPLYTIME_PROVIDER_DIR=./bin/
```

Not needed with `-g` (global install).

### 2d. Pre-Flight Check

```bash
complyctl doctor
```

Doctor is advisory — it reports diagnostics but does not block `generate` or `scan`:
```text
Workspace config valid
Pack manifest valid
Provider binary: complyctl-provider-openscap (./bin/)
Policy cache: cis-fedora-l1-workstation
Policy cache: cusp-fedora-default
System dependency: openscap-scanner (rpm)
Target 'local' -> cis-fedora-l1-workstation: variables OK
```

### 2e. Scan

```bash
complyctl scan --policy-id cis-fedora-l1-workstation
```

The scan uses bundled providers and pre-cached policies — zero registry access required.

---

## Part 3 — Global Install (Alternative)

```bash
complyctl pack install -g registry.example.com/packs/fedora-compliance:1.0.0
```

Providers install to `~/.complytime/providers/`, policies to `~/.complytime/policies/`. No `COMPLYTIME_PROVIDER_DIR` override needed — standard discovery paths.

`complypack.yaml` and `complytime.yaml.example` still land in the current directory.

---

## CLI Commands

### `complypack` (Platform Engineer)

| Command | Description |
|:---|:---|
| `complypack build` | Fetch policies, locate providers, generate example config, assemble OCI artifact |
| `complypack push <ref>` | Push pack artifact to OCI registry |
| `complypack pull <ref>` | Pull pack artifact to local directory |
| `complypack doctor` | Validate `complypack.yaml` buildability |

### `complyctl` (Admin — new in 002)

| Command | Description |
|:---|:---|
| `complyctl pack install <ref>` | Pull + extract pack to workspace |
| `complyctl pack install -g <ref>` | Pull + extract pack to global paths |

### `complyctl` (Existing from 001 — unchanged)

| Command | Description |
|:---|:---|
| `complyctl init` | Create `complytime.yaml` |
| `complyctl get` | Fetch policies from registries |
| `complyctl doctor` | Pre-flight diagnostics (now pack-aware) |
| `complyctl list` | List cached policies |
| `complyctl providers` | List discovered providers |
| `complyctl generate` | Prepare scan artifacts |
| `complyctl scan` | Execute compliance scan |
