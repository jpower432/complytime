# Tasks: Comply-Packs

**Branch**: `002-comply-packs` | **Date**: 2026-02-27
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Data Model**: [data-model.md](./data-model.md)

## Status Summary

| Phase | Scope | Tasks | Status |
|:---|:---|:---|:---|
| 1 | Setup — Prerequisite verification, directory structure, Makefile | T001–T003 | Pending |
| 2 | Foundational — Shared types, constants, pack library | T004–T011 | Pending |
| 3 | US1 — Build and Publish the `fedora-complypack` | T012–T016 | Pending |
| 4 | US2 — Install and Use a Comply-Pack | T017–T019 | Pending |
| 5 | US3 — Airgap Transfer Validation | T020 | Pending |
| 6 | Polish — Reference manifest, vendor sync, verification | T021–T023 | Pending |

---

## Phase 1: Setup — Prerequisite + Project Structure

**Purpose**: Ensure 001 prerequisite code is available, create new directories, add build targets.

- [ ] T001 Verify `internal/complytime/pack.go` exists with 001 pack types (`PackManifest`, `PlatformConfig`, `PackPolicyEntry`, `PackProviderEntry`, `SystemDependency`, `LoadPackManifest`, `ValidatePackManifest`, `PackManifestExists`, `PackPolicyIDs`). If missing, rebase on `main` or copy from `cmd/openscap-plugin/vendor/github.com/complytime/complyctl/internal/complytime/pack.go` and verify `PackManifestFile` and `CurrentPackSchemaVersion` constants exist in `internal/complytime/consts.go`
- [ ] T002 Create directory structure: `cmd/complypack/`, `cmd/complypack/cli/`, `internal/pack/`
- [ ] T003 Add `build-complypack` target to `Makefile` — `go build -o bin/complypack ./cmd/complypack/`

**Checkpoint**: 001 pack types available. New directories exist. Makefile builds both binaries.

---

## Phase 2: Foundational — Shared Types + Pack Library

**Purpose**: Update shared types for 002 schema changes, add pack media type constants, build the core pack library. All user story CLI commands depend on this phase.

### 2a: Constants + Provider Directory Override

- [ ] T004 [P] Add pack constants and `EnvProviderDir` to `internal/complytime/consts.go` — constants: `PackManifestFile` (`complypack.yaml`), `CurrentPackSchemaVersion` (`1`), `MediaTypePackManifest` (`application/vnd.complytime.pack.manifest.v1+yaml`), `MediaTypePackConfig` (`application/vnd.complytime.pack.config.v1+yaml`), `MediaTypePackProvider` (`application/vnd.complytime.pack.provider.v1`), `MediaTypePackPolicyLayout` (`application/vnd.complytime.pack.policy-layout.v1.tar+gzip`), `ExampleConfigFile` (`complytime.yaml.example`), `PackBinDir` (`bin`), `PackPoliciesDir` (`policies`), `EnvProviderDir` (`COMPLYTIME_PROVIDER_DIR`). Update `ResolvePluginDir()` to check `os.Getenv(EnvProviderDir)` first — if set and non-empty, return `ExpandPath(envVal)` instead of default `~/.complytime/providers/` (FR-206)

### 2b: Pack Type Schema Changes

- [ ] T005 [P] Update `internal/complytime/pack.go` — (1) Add `Arch string` field (`yaml:"arch"`) to `PlatformConfig`. (2) Replace `Install string` field with `URL string` field (`yaml:"url,omitempty"`) in `SystemDependency`. (3) Update `ValidatePackManifest()` to require `platform.arch` when platform is specified. (4) Update doc comments to reflect `url` replacing `install` for remediation guidance
- [ ] T006 Remove unused `RegistryConfig` type from `internal/complytime/config.go` if still present (schema simplification from 001 clarifications)

### 2c: Pack Library — OCI Assembly + Extraction

- [ ] T007 [P] Implement `BuildPackArtifact()` in `internal/pack/artifact.go` — accepts `PackManifest`, provider binary paths, policy OCI Layout dirs, and generated example config bytes. Assembles a single OCI manifest with typed layers per R3: one `MediaTypePackManifest` layer, one `MediaTypePackConfig` layer, one `MediaTypePackProvider` layer per provider binary, one `MediaTypePackPolicyLayout` layer per policy (tar+gzip compressed OCI Layout dir). Each layer annotated with `org.opencontainers.image.title` for extraction path routing. Returns `ocispec.Manifest` + content store. Uses `oras-go/v2`
- [ ] T008 [P] Implement `ExtractPackArtifact()` in `internal/pack/extract.go` — accepts OCI manifest + content store and target `PackInstallPaths`. Iterates layers, routes each by media type: `MediaTypePackProvider` → binary path (chmod +x), `MediaTypePackPolicyLayout` → policy dir (untar), `MediaTypePackManifest` → manifest path, `MediaTypePackConfig` → example config path. Returns list of written file paths (for rollback tracking). Errors include layer media type and annotation for diagnostics
- [ ] T009 [P] Implement `GenerateExampleConfig()` in `internal/pack/example_config.go` — accepts `*PackManifest`, calls `PackToPolicyEntries()` to build `[]PolicyEntry`, constructs `WorkspaceConfig` with policies, a placeholder `local` target referencing all policy IDs, and `variables: {workspace: ./.complytime/scan}`. Serializes via `SaveTo()` pattern. Returns `[]byte` (FR-208, R6)
- [ ] T010 [P] Implement `FetchPolicies()` in `internal/pack/policy_fetch.go` — accepts `[]PackPolicyEntry` and a credential function. For each entry: `ParsePolicyRef(entry.URL)`, create per-registry OCI client (same pattern as `complyctl get`), `oras.Copy()` into a local OCI Layout directory under a temp build area. Returns map of policy ID → OCI Layout directory path. Errors include the policy URL for diagnostics
- [ ] T011 [P] Implement `PackInstallPaths`, `ValidatePlatform()`, and `InstallWithRollback()` in `internal/pack/install.go` — `PackInstallPaths` struct resolves workspace-local vs global (`-g`) extraction paths per R7. `ValidatePlatform(manifest *PackManifest)` compares `manifest.Platform.OS`/`.Arch` against `runtime.GOOS`/`runtime.GOARCH`, returns error on mismatch naming expected vs actual (FR-211). `InstallWithRollback()` calls `ExtractPackArtifact()`, tracks written files, removes all on error (FR-212). Detects existing `complypack.yaml` for single-pack overwrite prompt (FR-205)

**Checkpoint**: Pack types updated. Constants centralized. Core library builds OCI artifacts, extracts layers, generates example configs, fetches policies, validates platform, and rolls back on failure. `go test ./internal/pack/...` passes with positive and negative cases for each exported function.

---

## Phase 3: User Story 1 — Build and Publish the `fedora-complypack`

**Story**: A platform engineer builds the `fedora-complypack` and publishes it to an OCI registry.

**Goal**: `complypack doctor` → `complypack build` → `complypack push` → `complypack pull` delivers a distributable pack artifact.

**Independent Test**: Run `complypack doctor` in a directory with `complypack.yaml` — reports buildability. Run `complypack build` — assembles OCI artifact. Run `complypack push <ref>` — pushes to test registry. Run `complypack pull <ref>` — retrieves artifact.

- [ ] T012 [US1] Create CLI entry point in `cmd/complypack/main.go` and root command with subcommand registration in `cmd/complypack/cli/root.go` — cobra-based, same patterns as `cmd/complyctl/`. Root registers `build`, `push`, `pull`, `doctor` subcommands (R4)
- [ ] T013 [P] [US1] Implement `complypack build` in `cmd/complypack/cli/build.go` — loads `complypack.yaml` via `LoadPackManifest()`, calls `FetchPolicies()` to pull each policy into OCI Layout dirs, locates provider binaries on local filesystem (error if missing), calls `GenerateExampleConfig()`, calls `BuildPackArtifact()` to assemble OCI artifact. Writes artifact to local OCI store. Progress output to stderr (FR-201)
- [ ] T014 [P] [US1] Implement `complypack push` in `cmd/complypack/cli/push.go` — accepts `<registry>/<name>:<tag>` argument. Pushes built OCI artifact using `oras.Copy()` from local store to remote registry. Authentication via shared `oras-credentials-go` credential function (FR-202)
- [ ] T015 [P] [US1] Implement `complypack pull` in `cmd/complypack/cli/pull.go` — accepts `<registry>/<name>:<tag>` argument. Pulls OCI artifact to local directory for inspection or offline transfer using `oras.Copy()` (FR-203)
- [ ] T016 [P] [US1] Implement `complypack doctor` in `cmd/complypack/cli/doctor.go` — loads `complypack.yaml`, validates manifest schema, checks provider binaries exist locally, checks registries reachable and policies exist remotely (per unique registries from policy URLs), runs system dependency `kind`/`value` checks (failed checks link to dependency `url`). Reports all results with actionable remediation (FR-204, FR-209)

**Checkpoint**: Platform engineer can validate, build, push, and pull comply-packs. `go test ./cmd/complypack/...` passes. US1 complete.

---

## Phase 4: User Story 2 — Install and Use a Comply-Pack

**Story**: An admin installs a comply-pack and runs a compliance scan using bundled providers and pre-cached policies.

**Goal**: `complyctl pack install <ref>` → edit `complytime.yaml` → `complyctl doctor` → `complyctl scan` delivers zero-registry-access scanning.

**Independent Test**: Run `complyctl pack install <oci-ref>` — extracts to `./bin/`, `./policies/`, `./complypack.yaml`, `./complytime.yaml.example`. Set `COMPLYTIME_PROVIDER_DIR=./bin/`, run `complyctl providers` — bundled provider listed. Run `complyctl doctor` — pack + config checks pass. Run `complyctl scan --policy-id <ID>` — scan executes with bundled content.

- [ ] T017 [US2] Implement `complyctl pack install` subcommand in `cmd/complyctl/cli/pack.go` — `pack` command group (`Use: "pack"`) with `install` subcommand. Accepts `<oci-ref>` argument and `-g` flag for global install. Pulls OCI artifact from registry, calls `ValidatePlatform()` to check host compatibility (FR-211), calls `InstallWithRollback()` for atomic extraction (FR-212). Single-pack per workspace: detects existing `complypack.yaml`, prompts for overwrite or accepts `--force` (FR-205)
- [ ] T018 [US2] Register `packCmd` in `cmd/complyctl/cli/root.go`
- [ ] T019 [US2] Add `checkPack()` advisory diagnostics to `internal/doctor/doctor.go` — when `PackManifestExists()` returns true: load and validate manifest, verify provider binaries exist at resolved paths (check `COMPLYTIME_PROVIDER_DIR` first, then `./bin/`, then global), verify policy OCI Layout dirs exist, run system dependency `kind`/`value` checks (warn with documentation `url` on failure). If `complytime.yaml` absent but `complypack.yaml` present, report pack OK + config missing with `cp complytime.yaml.example complytime.yaml` guidance. Advisory only — does not block `generate` or `scan` (FR-207, FR-209, FR-210). Update `Run()` to invoke `checkPack()` when pack manifest detected

**Checkpoint**: Admin can install, configure, validate, and scan with a comply-pack. `go test ./internal/doctor/...` passes with pack-aware cases. US2 complete.

---

## Phase 5: User Story 3 — Airgap Transfer Validation

**Story**: A platform engineer transfers a comply-pack to an airgapped environment using standard OCI tooling.

**Goal**: `skopeo copy` → tarball transfer → `skopeo copy` → `complyctl pack install` works end-to-end.

**Independent Test**: Use `skopeo copy docker://<src> oci-archive:/pack.tar`, transfer tarball, `skopeo copy oci-archive:/pack.tar docker://<local-registry>/<name>:<tag>`, run `complyctl pack install <local-registry>/<name>:<tag>` — install proceeds identically to connected install.

- [ ] T020 [US3] Validate airgap workflow end-to-end — build a test pack, push to registry, export with `skopeo copy` to `oci-archive`, import to local registry, install with `complyctl pack install`. Verify all extracted contents match the original build. Document the validated workflow in `specs/002-comply-packs/quickstart.md` Part 1e if not already present

**Checkpoint**: Airgap transfer works with standard `skopeo` commands. US3 complete.

---

## Phase 6: Polish — Reference Pack, Vendor Sync, Verification

**Purpose**: Create the `fedora-complypack` reference manifest, sync vendored copies, and verify full build.

- [ ] T021 Create `fedora-complypack` reference manifest at `examples/fedora-complypack/complypack.yaml` — uses the schema from data-model.md with `os: linux`, `arch: amd64`, Fedora policies, OpenSCAP provider, SCAP Security Guide system dependency with `kind: rpm`, `value: scap-security-guide`, and documentation `url`
- [ ] T022 Copy updated `internal/complytime/pack.go` and `internal/complytime/consts.go` to `cmd/openscap-plugin/vendor/github.com/complytime/complyctl/internal/complytime/` to sync vendored copies with 002 schema changes
- [ ] T023 Run `go build ./...`, `go test ./...`, `go vet ./...`, `gofmt -l .` — verify clean build across both binaries and all tests pass

**Checkpoint**: Reference pack manifest exists. Vendor synced. Full build green. 002 implementation complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1 (pack.go must exist before modification)
- **Phase 3 (US1)**: Depends on Phase 2 (CLI commands use `internal/pack/` library)
- **Phase 4 (US2)**: Depends on Phase 2 (uses `internal/pack/install.go` and `internal/doctor/`). Independent of Phase 3 (different binaries, different CLI files)
- **Phase 5 (US3)**: Depends on Phase 3 + Phase 4 (needs both `complypack push` and `complyctl pack install`)
- **Phase 6 (Polish)**: Depends on all prior phases

### Parallel Opportunities

**Within Phase 2**: T004 + T005 are parallel (different files: `consts.go` vs `pack.go`). T007, T009, T010 are fully parallel (no interdependencies). T008 depends on T011 (`ExtractPackArtifact` accepts `PackInstallPaths` defined in `install.go`). T006 is conditional — execute only if `RegistryConfig` is present in `config.go`.

**Phase 3 vs Phase 4**: T012–T016 (US1, `cmd/complypack/`) and T017–T019 (US2, `cmd/complyctl/cli/pack.go` + `internal/doctor/`) are independent — different binaries, different source files. Can be implemented in parallel after Phase 2.

**Within Phase 3**: T013–T016 are parallel (different CLI command files in `cmd/complypack/cli/`), all depend on T012 (root command registration).

```text
Phase 1:            T001 → T002 → T003
Phase 2 types:      [T004, T005] → T006
Phase 2 library:    [T007, T009, T010, T011]  (after T004+T005)
                    T008  (after T011)
Phase 3 (US1):      T012 → [T013, T014, T015, T016]
Phase 4 (US2):      [T017, T018, T019]  (parallel with Phase 3)
Phase 5 (US3):      T020  (after Phase 3 + 4)
Phase 6:            [T021, T022] → T023
```

---

## Implementation Strategy

### MVP (Phases 1–3): Build and Publish

Setup + foundational library + US1 delivers a working `complypack` binary that platform engineers can use to build, validate, push, and pull comply-packs.

**Task count**: 16 (T001–T016)

### Core Value (Phase 4): Install and Scan

US2 delivers the consumer-side workflow: `complyctl pack install` + pack-aware `doctor`. Admin can go from OCI reference to running scan in under 3 minutes.

**Task count**: 3 (T017–T019)

### Full Feature (Phase 5): Airgap

US3 validates the airgap workflow with standard OCI tooling. No custom code — validation only.

**Task count**: 1 (T020)

### Polish (Phase 6): Reference + Verification

Reference `fedora-complypack` manifest, vendor sync, full build verification.

**Task count**: 3 (T021–T023)

---

## Key Implementation Notes

- **Pack types from 001**: `PackManifest`, `PlatformConfig`, `PackPolicyEntry`, `PackProviderEntry`, `SystemDependency` exist in `internal/complytime/pack.go` (created in 001 Phase 6b). 002 modifies them (add `Arch`, replace `Install` with `URL`)
- **Shared auth**: Both `complypack` and `complyctl` use `oras-credentials-go` via `internal/registry/auth.go`. Zero custom auth code
- **Single-platform**: Each pack targets one `os`/`arch` pair. `complyctl pack install` validates host compatibility before extraction. Multi-arch via OCI Image Index deferred
- **Advisory doctor**: Pack-aware `checkPack()` reports issues but does not block `generate` or `scan`. Doctor is a debugging tool
- **Install rollback**: `InstallWithRollback()` tracks all written files during extraction. On any error, removes all tracked files. No partial installs
- **Single pack per workspace**: Second install replaces first (overwrite prompt or `--force`). No merge logic
- **System dep remediation**: `SystemDependency` declares `kind`/`value` for safe hardcoded checks and `url` for documentation. No install commands
- **Example config generation**: `complytime.yaml.example` uses standard `PolicyEntry` format from 001. URLs copied directly from pack manifest. Placeholder `local` target
- **Convention over configuration**: `complypack.yaml` presence triggers pack-aware doctor. `COMPLYTIME_PROVIDER_DIR` is the only env var override. Sensible defaults for all install paths

---

## Notes

- [P] tasks = different files, no dependencies — safe to parallelize
- [USn] label maps task to specific user story
- Each phase is independently testable at its checkpoint
- Tests are REQUIRED per constitution — each implementation task (T004–T019) MUST include `_test.go` file(s) with at least one positive and one negative test case for every exported function. Tests are part of the task's definition of done, not a separate step
- Total tasks: 23. US1: 5. US2: 3. US3: 1
