# Comply-Pack Process Forecast

**From pack creation to compliance scan in a workspace.**

## Personas and Handoff

```text
Platform Engineer                    Admin / App Owner
─────────────────                    ─────────────────
curate policies + providers    ──→   receive OCI reference
build pack                           install pack
push to registry                     customize config
                                     run scan
```

The pack is the handoff artifact. Everything before `push` is the platform engineer's domain. Everything after `install` is the admin's domain. The OCI reference is the only thing that crosses the boundary.

## Stage 1 — Curate

The platform engineer creates `complypack.yaml` in a working directory. This manifest declares what goes into the pack: which policies to fetch (each with its own OCI URL), which provider binaries to bundle, and what system dependencies the consumer needs.

```yaml
id: fedora-compliance
version: "1.0.0"
policies:
  - url: registry.example.com/compliance/cis-fedora-l1-workstation@1.2.3
    id: cis-fedora-l1-workstation
    profile: cis_workstation_l1
  - url: public-registry.io/community/cusp-fedora-default@1.0.0
    id: cusp-fedora-default
providers:
  - id: openscap
    binary: complyctl-provider-openscap
system-dependencies:
  - name: openscap-scanner
    check: oscap --version
    install: dnf install -y openscap-scanner
```

Each policy carries its own URL — policies from different registries coexist in one pack. The provider binary (`complyctl-provider-openscap`) must exist on the engineer's machine. Policies are fetched from their respective registries during build — they don't need to be local.

## Stage 2 — Build

```text
complypack build
```

This does four things in sequence:

1. **Fetches policies** from their respective registries into local OCI Layout directories (one per policy)
2. **Locates provider binaries** on the local filesystem
3. **Generates `complytime.yaml.example`** — a ready-to-use consumer config with `PolicyEntry` URLs copied directly from the manifest:

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
      - cusp-fedora-default
```

4. **Assembles the OCI artifact** — a single OCI manifest with typed layers:

```text
OCI Manifest
├── complypack.yaml                              (pack manifest)
├── complytime.yaml.example                      (generated config)
├── bin/complyctl-provider-openscap              (provider binary)
├── policies/cis-fedora-l1-workstation.tar.gz    (OCI Layout)
└── policies/cusp-fedora-default.tar.gz          (OCI Layout)
```

Each layer has a custom media type so `complyctl pack install` knows where to extract it.

## Stage 3 — Publish

```text
complypack push registry.example.com/packs/fedora-compliance:1.0.0
```

The pack is now a standard OCI artifact in the registry. Anyone with pull access and the OCI reference can install it.

## Stage 4 — Distribute (optional: airgap)

For connected environments, the OCI reference is the distribution mechanism. For airgapped environments, `skopeo` handles the transfer:

```text
Connected host:
  skopeo copy docker://registry.example.com/packs/fedora-compliance:1.0.0 \
               oci-archive:/tmp/fedora-compliance.tar

USB / sneakernet transfer

Airgapped host:
  skopeo copy oci-archive:/tmp/fedora-compliance.tar \
               docker://local-registry:5000/packs/fedora-compliance:1.0.0
```

No custom tooling. Standard OCI operations.

## Stage 5 — Install

The admin receives the OCI reference and installs the pack into their workspace:

```text
complyctl pack install registry.example.com/packs/fedora-compliance:1.0.0
```

The workspace now contains:

```text
./
├── bin/
│   └── complyctl-provider-openscap     ← provider binary, chmod +x
├── policies/
│   ├── cis-fedora-l1-workstation/      ← OCI Layout (untarred)
│   └── cusp-fedora-default/            ← OCI Layout (untarred)
├── complypack.yaml                     ← pack manifest (for doctor)
└── complytime.yaml.example             ← starter config
```

**Alternative: global install** (`-g`) puts providers into `~/.complytime/providers/` and policies into `~/.complytime/policies/` — the standard `complyctl` discovery paths. No env var override needed.

## Stage 6 — Configure

```text
cp complytime.yaml.example complytime.yaml
```

The example config is functional out of the box. The admin edits it to narrow scope or add target-specific variables:

```yaml
policies:
  - url: registry.example.com/compliance/cis-fedora-l1-workstation@1.2.3
    id: cis-fedora-l1-workstation

variables:
  workspace: ./.complytime/scan

targets:
  - id: local
    policies:
      - cis-fedora-l1-workstation
    variables:
      profile: cis_workstation_l1
```

For workspace-local installs, the admin sets one env var so `complyctl` finds the bundled providers:

```text
export COMPLYTIME_PROVIDER_DIR=./bin/
```

## Stage 7 — Validate

```text
complyctl doctor
```

Doctor detects `complypack.yaml` and runs pack-layer checks alongside standard config checks:

```text
✅ Workspace config valid
✅ Pack manifest valid
✅ Provider: complyctl-provider-openscap
✅ Policy cache: cis-fedora-l1-workstation
✅ System dependency: openscap-scanner
✅ Target 'local' → cis-fedora-l1-workstation: variables OK
```

If `complytime.yaml` is missing, doctor reports the pack is installed and suggests `cp complytime.yaml.example complytime.yaml`.

## Stage 8 — Scan

```text
complyctl scan --policy-id cis-fedora-l1-workstation
```

The scan uses the bundled provider and pre-cached policy. Zero registry access. The output lands in `./.complytime/scan/` as an EvaluationLog plus any formatted reports.

From the admin's perspective, this is identical to the non-pack workflow. The pack just pre-populated the workspace with everything needed to run.

## Artifact Flow Summary

```text
Platform Engineer                          OCI Registry                          Admin Workspace
─────────────────                          ────────────                          ───────────────

complypack.yaml ──→ complypack build ──→ OCI Artifact ──→ complyctl pack install ──→ ./bin/
                    (fetch policies)      (typed layers)   (extract by media type)    ./policies/
                    (locate providers)                                                complypack.yaml
                    (generate example)                                                complytime.yaml.example
                                                                                          │
                                                                                          ▼
                                                                                     cp → complytime.yaml
                                                                                          │
                                                                                          ▼
                                                                                     complyctl doctor
                                                                                          │
                                                                                          ▼
                                                                                     complyctl scan
                                                                                          │
                                                                                          ▼
                                                                                     ./.complytime/scan/
                                                                                     evaluation-log.yaml
```
