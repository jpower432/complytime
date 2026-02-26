# Tasks: Gemara-Native Decoupled Workflow

**Branch**: `001-gemara-native-workflow` | **Date**: 2026-02-14 (regenerated 2026-02-25e)
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Data Model**: [data-model.md](./data-model.md)

## Status Summary

| Phase | Scope | Status |
|:---|:---|:---|
| 1 | Setup — Legacy removal, dependency updates, proto codegen | **COMPLETE** |
| 2 | Foundational — Cache, registry, policy resolution, plugin SDK | **COMPLETE** |
| 3 | US1 — Init, Get, Doctor, List | **COMPLETE** |
| 4 | US2 — Generate, Scan, Execution Plan | **COMPLETE** |
| 5 | US3 — Output Formatters, Scan Summary | **COMPLETE** |
| 6 | Polish — Behavioral tests, spinner, pack data model, PolicyEntry | **COMPLETE** |
| 7 | Documentation — quickstart.md alignment | **COMPLETE** |
| 8 | Doctor Redesign — Version Comparison + Per-Provider Config (R55) | **COMPLETE** |
| 9 | 002-comply-packs — Pack CLI + Doctor Integration | **DEFERRED** |

---

## Phase 1: Setup — Legacy Removal + Dependencies (COMPLETE)

**Purpose**: Remove old C2P/OSCAL code paths, update dependencies, establish proto contract.

- [x] T001 Remove `info` command from `cmd/complyctl/cli/` and `root.go` registration
- [x] T002 Remove old `plan` command from `cmd/complyctl/cli/` and `root.go` registration
- [x] T003 Remove C2P framework imports (`compliance-to-policy-go/v2/framework`) from all source files
- [x] T004 Remove `oscal-sdk-go` dependency from `go.mod` and all imports
- [x] T005 Gut `internal/complytime/` — remove C2P-specific logic (`Config`, `Plugins`, `ActionsContextFromPlan`, `LoadFrameworks`, `FindComponentDefinitions`, `LoadProfile`, `LoadCatalogSource`, `WriteAssessmentResults`)
- [x] T006 Remove old OSCAL workflow: `loadPlan`, `actions.AggregateResults`, `actions.Report`, `actions.GeneratePolicy`, `WriteAssessmentResults`, profile/catalog resolution
- [x] T007 Remove `policytype` package from `cmd/openscap-plugin/policytype/`
- [x] T008 Replace `oscalPolicy` references in `cmd/openscap-plugin/` with `[]plugin.AssessmentConfiguration`
- [x] T009 Define proto contract in `specs/001-gemara-native-workflow/contracts/plugin.proto` — `EvaluatorService` with `Generate`, `Scan`, `HealthCheck` RPCs. `ConfidenceLevel` enum (NOT_SET, UNDETERMINED, LOW, MEDIUM, HIGH). `Result` enum (UNSPECIFIED, PASSED, FAILED, SKIPPED, ERROR). `ScanRequest` receives targets only (no `requirement_ids`, R47). `HealthCheckResponse` includes `required_global_variables` and `required_target_variables` (R51)
- [x] T010 Run `buf generate` to produce Go stubs from `plugin.proto`
- [x] T011 Run `buf lint` and `buf breaking` to verify proto quality
- [x] T012 Add Gemara media type constants to `internal/complytime/consts.go` — `MediaTypeCatalog`, `MediaTypeGuidance`, `MediaTypePolicy` (R27)
- [x] T013 Add scan output constants to `internal/complytime/consts.go` — `ScanOutputDir = ".complytime/scan"` (R42), emoji status constants (R43)
- [x] T014 Run `go mod tidy` and `go mod vendor` to clean dependencies
- [x] T015 Run `go build ./...`, `go vet ./...`, `gofmt -l .` — verify clean baseline

**Checkpoint**: Legacy code removed. Proto contract defined. Constants centralized. Clean build.

---

## Phase 2: Foundational — Cache, Registry, Policy, Plugin SDK (COMPLETE)

**Purpose**: Build shared infrastructure that all user stories depend on. No user-facing CLI changes yet.

### 2a: OCI Cache + State Tracking

- [x] T016 [P] Implement `internal/cache/cache.go` — OCI Layout store wrapper using `oras-go/v2/content/oci`. `NewStore(policyID)` creates per-policy OCI Layout directory under `~/.complytime/policies/{id}/`
- [x] T017 [P] Implement `internal/cache/state.go` — `CacheState` struct with `LastSync` and `Policies map[string]PolicyState`. `LoadState()`, `SaveState()` for `~/.complytime/state.json`
- [x] T018 Implement `internal/cache/sync.go` — `SyncPolicy()` using `oras.Copy()` from remote registry to local OCI Layout store. Atomic: failure rolls back to previous state (FR-006)
- [x] T019 Implement `internal/cache/source.go` — `PolicySource` interface + `RegistrySource` implementation
- [x] T020 [P] Implement `internal/cache/cachetest/mock_source.go` — `MockPolicySource` for tests

### 2b: Registry Client + Auth

- [x] T021 [P] Implement `internal/registry/auth.go` — `NewCredentialFunc()` using `oras-credentials-go` `credentials.NewStoreFromDocker()` (R6, R24). Zero custom auth code
- [x] T022 Implement `internal/registry/client.go` — OCI registry client wrapping `oras-go/v2` `remote.Repository`. `PlainHTTP` support for local registries
- [x] T023 [P] Implement `internal/registry/fetcher.go` — `Fetcher` interface for testability. Also implements `internal/registry/resolver.go` — `GetDefinitions`, `DefinitionVersion` resolver methods (FR-002)

### 2c: Policy Resolution

- [x] T024 Implement `internal/policy/loader.go` — OCI Layout → layer extraction by media type matching (`MediaTypeCatalog`, `MediaTypeGuidance`, `MediaTypePolicy`). Returns typed content per layer (R25, R27)
- [x] T025 Implement `internal/policy/resolver.go` — `ResolvePolicyGraph()` builds `DependencyGraph` (controls, guidelines, assessments). `parsePolicyLayer` accepts only `gemara.Policy` with `adherence.assessment-plans` (R39)
- [x] T026 Implement `internal/policy/assessment.go` — `ExtractAssessmentConfigs()`, `GroupByEvaluator()` (per-plan `executor.id`, R32), `ValidateGlobalVars()` (R48)
- [x] T027 Implement `internal/policy/generation_state.go` — `GenerationState`, `SaveState()`, `LoadState()`, `IsFresh(currentDigest)` for digest-based freshness detection (R37)

### 2d: Plugin SDK (gRPC Scanning Interface)

- [x] T028 Implement `pkg/plugin/client.go` — gRPC client wrapper + domain types (`AssessmentConfiguration`, `AssessmentLog`, `ConfidenceLevel` enum). Public SDK for scanning provider authors
- [x] T029 [P] Implement `pkg/plugin/server.go` — gRPC server adapter mapping proto messages to domain types. Public SDK for scanning provider authors
- [x] T030 Implement `pkg/plugin/plugin.go` — `Handshake`, `GRPCEvaluatorPlugin`, `Serve()` using `hashicorp/go-plugin`
- [x] T031 Implement `pkg/plugin/manager.go` — Scanning provider lifecycle: load, health check, route Generate/Scan RPCs by evaluator ID
- [x] T032 Implement `pkg/plugin/discovery.go` — Filesystem discovery of `complyctl-provider-*` executables in `~/.complytime/providers/` (FR-029)
- [x] T033 Implement `pkg/plugin/initialization.go` — `NewClient(path, logger)` simplified (R21, no manifests, no checksums)
- [x] T034 [P] Implement `pkg/plugin/export_test.go` — `RegisterPluginForTest` test helper

### 2e: Workspace + Config

- [x] T035 Implement `internal/complytime/workspace.go` — `NewWorkspace()`, `Exists()`, `Path()`, `EnsureDir()`, `Save()` (R50)
- [x] T036 Implement `internal/complytime/config.go` — `WorkspaceConfig` struct with `Policies []PolicyEntry`, `Variables map[string]string`, `Targets []TargetConfig`. `PolicyEntry` struct (`URL`, `ID`, `EffectiveID()` method). `LoadFrom()`, `SaveTo()`, `Validate()`, `FindPolicy()`, `PolicyIDs()`, `UniqueRegistries()`, `ParsePolicyRef()`, `ValidateTargetPolicyVersions()`, `ResolveEnvVars()` (Session 2026-02-25d)
- [x] T037 [P] Implement `internal/complytime/config_test.go` — unit tests for `PolicyEntry`, `EffectiveID`, `FindPolicy`, `Validate`, `ValidateTargetPolicyVersions`, `UniqueRegistries`, `PolicyIDs`, `ParsePolicyRef`, `ResolveEnvVars`

### 2f: Terminal Helpers

- [x] T038 [P] Implement `internal/terminal/spinner.go` — charmbracelet/bubbles spinner wrapper (Constitution V)
- [x] T039 [P] Implement `internal/terminal/table.go` — reusable charmbracelet table helpers: `Model`, `ShowPlainTable` (R38)

### 2g: Verification

- [x] T040 Run `go build ./...` — all Phase 2 code compiles
- [x] T041 Run `go test ./internal/...` — all unit tests pass
- [x] T042 Run `go vet ./...` and `gofmt -l .` — clean

**Checkpoint**: All shared infrastructure built. Cache, registry, policy resolution, plugin SDK, config, terminal helpers. Ready for CLI commands.

---

## Phase 3: User Story 1 — Initialize, Fetch, Validate (COMPLETE)

**Story**: A system administrator needs to set up complyctl, fetch policies, and validate the environment.

**Goal**: `complyctl init` → `complyctl get` → `complyctl doctor` → `complyctl list` delivers a working policy cache.

**Independent Test**: Run `complyctl init` (creates config), `complyctl get` (fetches policies), `complyctl doctor` (validates), `complyctl list` (shows cached policies).

### 3a: `complyctl init` (FR-003)

- [x] T043 [US1] Implement `cmd/complyctl/cli/init.go` — config-creation-only. Error if `complytime.yaml` exists. Interactive prompts: `promptPolicies()` asks for policy URLs + optional IDs (builds `[]PolicyEntry`), `promptTargets()` shows available effective IDs and collects target policies. Builds `WorkspaceConfig`, calls `workspace.Save()`, prints status to stderr, exits. No `get`, no `doctor` (Session 2026-02-25d)
- [x] T044 [US1] Register `initCmd` in `cmd/complyctl/cli/root.go`

### 3b: `complyctl get` (FR-002, FR-004, FR-005, FR-006)

- [x] T045 [US1] Implement `cmd/complyctl/cli/get.go` — load config, iterate `cfg.Policies` as `[]PolicyEntry`, create per-registry OCI clients via `UniqueRegistries()`. For each policy: `ParsePolicyRef(entry.URL)`, resolve version, `oras.Copy()` to local OCI Layout store. Per-policy progress to stderr (Tier 1 output, FR-035). Atomic sync with rollback on failure (FR-006)
- [x] T046 [US1] Register `getCmd` in `cmd/complyctl/cli/root.go`

### 3c: `complyctl doctor` (FR-039)

- [x] T047 [US1] Implement `internal/doctor/doctor.go` — `Run()` function. Checks: (1) config validation (`LoadFrom` + `Validate` + `ValidateTargetPolicyVersions`), (2) provider discovery + HealthCheck, (3) registry reachability (non-blocking warning), (4) HealthCheck-declared variable validation — global keys against `config.variables`, target keys against relevant `config.targets[].policies` using policy → evaluator → target mapping from cache (R51). Emoji + message output. Exit 0 if all blocking checks pass
- [x] T048 [US1] Implement `cmd/complyctl/cli/doctor.go` — CLI entrypoint. Load config, pass `cfg.Policies` to `UniqueRegistries()` for registry probes. Invoke `doctor.Run()`
- [x] T049 [US1] Register `doctorCmd` in `cmd/complyctl/cli/root.go`

### 3d: `complyctl list` (FR-031)

- [x] T050 [US1] Implement `cmd/complyctl/cli/list.go` — load config, iterate `cfg.Policies` as `[]PolicyEntry`. For each: `ParsePolicyRef(entry.URL)`, check cache status, display `entry.EffectiveID()` + version + cache status. Charmbracelet table + `--plain` flag (R38)
- [x] T051 [US1] Register `listCmd` in `cmd/complyctl/cli/root.go`

### 3e: `complyctl providers` (FR-032)

- [x] T052 [US1] Implement `cmd/complyctl/cli/providers.go` — discover plugins in `~/.complytime/providers/`, display evaluator ID + path + health + version. Charmbracelet table + `--plain` flag (R38)
- [x] T053 [US1] Register `providersCmd` in `cmd/complyctl/cli/root.go`

### 3f: Verification

- [x] T054 Run `go build ./...` — all Phase 3 code compiles
- [x] T055 Run `go test ./...` — all tests pass
- [x] T056 Run `go vet ./...` and `gofmt -l .` — clean

**Checkpoint**: Admin can init workspace, fetch policies, validate environment, list policies and providers. US1 complete.

---

## Phase 4: User Story 2 — Generate + Scan (COMPLETE)

**Story**: A system administrator needs to generate policy artifacts and execute compliance scans.

**Goal**: `complyctl generate` → `complyctl scan` (or just `complyctl scan` with auto-generate) produces EvaluationLog and terminal summary.

**Independent Test**: `complyctl scan --policy-id <ID>` auto-generates and scans. `complyctl generate --policy-id <ID>` followed by `complyctl scan --policy-id <ID>` reuses artifacts. `complyctl scan --policy-id <ID> --dry-run` outputs execution plan without scanning.

### 4a: `complyctl generate` (FR-007, FR-008, FR-009)

- [x] T057 [US2] Implement `cmd/complyctl/cli/generate.go` — load config, `FindPolicy()` returns `*PolicyEntry`, `ParsePolicyRef(entry.URL)` for graph resolution. `ResolvePolicyGraph()` → `ExtractAssessmentConfigs()` → `GroupByEvaluator()` → `ValidateGlobalVars()` → `RouteGenerate()` (global vars + test vars via Generate RPC). Persist `GenerationState`. Output `FormatExecutionPlan()` (charmbracelet tables)
- [x] T058 [US2] Register `generateCmd` in `cmd/complyctl/cli/root.go`

### 4b: `complyctl scan` (FR-024, FR-012, FR-033, FR-034)

- [x] T059 [US2] Implement `cmd/complyctl/cli/scan.go` — load config, `FindPolicy()` returns `*PolicyEntry`, check `GenerationState.IsFresh()`: fresh → reuse, stale → warn + auto-generate, missing → auto-generate. Print brief one-line summary (FR-034). `RouteScan()` with targets only (R47). Build `EvaluationLog` from `AssessmentLog[]`. Write to `{ScanOutputDir}/evaluation-log.yaml`. Display `ScanSummary` (FR-037). `--dry-run` flag: generate + execution plan + exit (FR-033). `--format` flag: oscal/pretty/sarif (Phase 5). `--policy-id` required
- [x] T060 [US2] Register `scanCmd` in `cmd/complyctl/cli/root.go`

### 4c: Execution Plan Output (FR-033, R36)

- [x] T061 [P] [US2] Implement `internal/output/execution_plan.go` — `FormatExecutionPlan()`. Two charmbracelet tables: (1) Evaluator Routing (evaluator ID, requirement count, plugin path, status), (2) Target Scope (target ID, policy ID, evaluator IDs). Unmatched evaluators show ERROR status (R36, R38)

### 4d: EvaluationLog Builder (FR-012)

- [x] T062 [P] [US2] Implement `internal/output/evaluator.go` — builds `*gemara.EvaluationLog` from `[]AssessmentLog`. Maps `AssessmentLog` → `gemara.ControlEvaluation` + `gemara.AssessmentLog`. Result aggregation via go-gemara (R45). YAML output via `goccy/go-yaml`

### 4e: OpenSCAP Provider Updates

- [x] T063 [US2] Update `cmd/openscap-plugin/server/server.go` — implement `Generate` and `Scan` RPCs using `[]plugin.AssessmentConfiguration` (R30). Scan evaluates all requirements from Generate-time state (R47)
- [x] T064 [US2] Update `cmd/openscap-plugin/xccdf/tailoring.go` — accept `[]plugin.AssessmentConfiguration` instead of `oscalPolicy` (R30)
- [x] T065 [US2] Update `cmd/openscap-plugin/xccdf/datastream.go` — use assessment configuration parameters
- [x] T066 [US2] Run `go mod vendor` in `cmd/openscap-plugin/` to sync vendored dependencies

### 4f: Test Plugin

- [x] T067 [P] [US2] Implement `cmd/test-plugin/` — E2E test scanning provider binary. Implements Generate, Scan, HealthCheck RPCs. Returns deterministic AssessmentLog entries for test verification. NOT referenced by production code

### 4g: Verification

- [x] T068 Run `go build ./...` — all Phase 4 code compiles
- [x] T069 Run `go test ./...` — all tests pass
- [x] T070 Run `go vet ./...` and `gofmt -l .` — clean

**Checkpoint**: Admin can generate artifacts and scan. Auto-generate, reuse, stale-detect all work. Dry-run outputs execution plan. US2 complete.

---

## Phase 5: User Story 3 — Output Formats + Scan Summary (COMPLETE)

**Story**: A system administrator needs scan results in multiple formats and a clear terminal summary.

**Goal**: `complyctl scan --format <oscal|pretty|sarif>` produces formatted reports. Terminal always shows ActionError-style summary.

**Independent Test**: `complyctl scan --policy-id <ID> --format oscal` produces OSCAL JSON. `--format pretty` produces Markdown + EvaluationLog. `--format sarif` produces SARIF JSON. Default (no `--format`) produces EvaluationLog only. Terminal always shows emoji + message summary + totals table.

### 5a: Output Formatters (FR-014, FR-025, FR-026, FR-027)

- [x] T071 [P] [US3] Implement `internal/output/oscal.go` — OSCAL export using `go-oscal` types (`AssessmentResults`, `Finding`, `Observation`). Maps `EvaluationLog` → OSCAL (FR-014)
- [x] T072 [P] [US3] Implement `internal/output/markdown.go` — Markdown report. Optionally embeds EvaluationLog when `--format pretty` (FR-025, FR-027)
- [x] T073 [P] [US3] Implement `internal/output/sarif.go` — SARIF export using `go-gemara/gemaraconv` SARIF conversion (FR-026)

### 5b: Scan Summary (FR-037, R45)

- [x] T074 [US3] Implement `internal/output/scan_summary.go` — `FormatScanSummary()`. Non-passing results: emoji + message per failure/error/skip (no requirement ID). Sort by severity: failed → error → skipped. Single-row charmbracelet totals table. Result aggregation via go-gemara. Message from `Steps[].Message` (first match)

### 5c: Wire Formatters into Scan

- [x] T075 [US3] Update `cmd/complyctl/cli/scan.go` — wire `--format` flag dispatch: `oscal` → `oscal.go`, `pretty` → `markdown.go`, `sarif` → `sarif.go`. Default: EvaluationLog only (FR-028). Always display `ScanSummary` in terminal (FR-037)

### 5d: Verification

- [x] T076 Run `go build ./...` — all Phase 5 code compiles
- [x] T077 Run `go test ./...` — all tests pass (including output formatter tests)
- [x] T078 Run `go vet ./...` and `gofmt -l .` — clean

**Checkpoint**: All output formats work. Scan summary displays after every scan. US3 complete.

---

## Phase 6: Polish — Behavioral Tests, Pack Data Model, PolicyEntry (COMPLETE)

**Purpose**: Cross-cutting concerns, governance tests, pack data model for 002, and final config refactoring.

### 6a: Behavioral Assessment Tests

- [x] T079 [P] Implement `tests/behavioral/reusable_steps.go` — shared test infrastructure for behavioral assessment. Config YAML uses `PolicyEntry` format
- [x] T080 [P] Implement `tests/behavioral/transport_security.go` — TLS transport security assessment
- [x] T081 [P] Implement `tests/behavioral/log_security.go` — log credential redaction assessment
- [x] T082 [P] Implement `tests/behavioral/credential_protection.go` — credential protection assessment
- [x] T083 Implement `.github/workflows/behavioral_assessment.yml` — CI workflow for behavioral tests

### 6b: Pack Data Model (types only, CLI deferred to 002)

- [x] T084 [P] Implement `internal/complytime/pack.go` — `PackManifest`, `PlatformConfig`, `PackPolicyEntry`, `PackProviderEntry`, `SystemDependency` structs. `LoadPackManifest()`, `ValidatePackManifest()`, `PackManifestExists()`, `PackPolicyIDs()` (R53)
- [x] T085 [P] Implement `internal/complytime/pack_test.go` — pack manifest validation, loading, YAML parsing, duplicate detection
- [x] T086 Add `PackManifestFile` constant to `internal/complytime/consts.go` (R53)

### 6c: E2E + Integration Test Updates

- [x] T087 Update `tests/e2e/e2e_test.go` — embedded YAML configs use `PolicyEntry` format (url + id, not registry + policy_ids)
- [x] T088 Update `tests/e2e/helpers_test.go` — `writeWorkspaceConfig()` generates `PolicyEntry` format
- [x] T089 Update `tests/integration_test.sh` — complytime.yaml content uses `PolicyEntry` format
- [x] T090 Update `cmd/mock-oci-registry/testdata/sample-complytime.yaml` — `PolicyEntry` format

### 6d: Version Command

- [x] T091 Update `cmd/complyctl/cli/version.go` — clean version output, no log file created (FR-035)

### 6e: CLI Options + Root

- [x] T092 Implement `cmd/complyctl/cli/options.go` — shared CLI option types
- [x] T093 Update `cmd/complyctl/cli/root.go` — register all commands (init, get, list, providers, generate, scan, doctor, version). No `pack` command (deferred to 002)

### 6f: Verification

- [x] T094 Run `go build ./...` — full build green
- [x] T095 Run `go test ./...` — all tests pass
- [x] T096 Run `go vet ./...` and `gofmt -l .` — clean
- [x] T097 Verify no stale references — search for `Pack `, `RegistryConfig`, `policy_ids` in Go source (except `pack.go` data model); zero matches expected

**Checkpoint**: Behavioral tests, pack data model, PolicyEntry propagated. Full build green. All 001 implementation complete.

---

## Phase 7: Documentation Alignment (COMPLETE)

**Purpose**: Align documentation artifacts with the implemented `PolicyEntry` model. quickstart.md still references the superseded dual-mode config (`pack` field, `policy_ids`).

### 7a: quickstart.md Update

- [x] T098 Update `specs/001-gemara-native-workflow/quickstart.md` Section 1 — replace pack-mode `init` example with `PolicyEntry` prompts (policy URLs + optional IDs). Replace YAML showing `pack:` field with `policies:` list of `url:` + `id:` entries. Replace `policy_ids:` with `policies:` in targets
- [x] T099 Update `specs/001-gemara-native-workflow/quickstart.md` Section 1b — replace standalone-mode YAML showing `registry:` + `policies:` (id/version format) with `PolicyEntry` format. Replace `policy_ids:` with `policies:` in targets
- [x] T100 Update `specs/001-gemara-native-workflow/quickstart.md` — remove all references to dual-mode config, pack-mode init, and `complyctl pack init`. Add note that pack CLI is deferred to 002
- [x] T101 Review remaining sections of quickstart.md (workflow steps 2-5) for stale `policy_ids` or `registry` references and update to match current implementation

### 7b: Verification

- [x] T102 Grep quickstart.md for `policy_ids`, `pack:`, `registry:` — zero matches expected (except historical context or 002 forward-references)

**Checkpoint**: All documentation aligned with `PolicyEntry` implementation.

---

## Phase 8: Doctor Redesign — Version Comparison + Per-Provider Config (COMPLETE)

**Purpose**: Replace doctor's registry reachability probe with per-policy version comparison. Add per-provider configuration summary with `--verbose` drill-down. Implements R55 (Session 2026-02-25e).

**Dependencies**: All 001 phases complete (Phases 1-7). Uses existing `DefinitionVersion()` from `internal/registry/resolver.go`, `PolicyState` from `internal/cache/state.go`, and `ProviderHealth` from `internal/doctor/doctor.go`.

**Independent Test**: `complyctl doctor` shows per-policy version status (latest vs stale) and per-provider config summary (resolved/missing counts). `complyctl doctor --verbose` expands provider variable detail to per-key status. Unreachable registries produce per-registry warning, skip version check for those policies.

### 8a: Version Comparison Check (replaces CheckRegistries)

- [x] T103 [US1] Implement `CheckPolicyVersions()` in `internal/doctor/doctor.go` — accepts `*WorkspaceConfig`, `cacheDir string`, registry resolver. For each policy in config: load `PolicyState` from `state.json` (cached version/digest), group policies by registry via `UniqueRegistries()`. Per registry: attempt `DefinitionVersion()` query for latest version. If unreachable → emit non-blocking warning per registry (e.g., `⚠️ registry/X: unreachable — version check skipped`), skip all policies from that registry. Per reachable policy: compare cached version against remote latest — stale → warning with both versions + remediation (`run complyctl get`), up-to-date → pass. Returns `[]CheckResult`
- [x] T104 [US1] Update `Run()` in `internal/doctor/doctor.go` — replace `CheckRegistries(registries)` call with `CheckPolicyVersions(cfg, cacheDir, versionResolver)`. Add `VersionResolver` interface parameter (satisfied by `internal/registry/resolver.go`) for testability. Keep `CheckCache()` before `CheckPolicyVersions()` (version check needs cache to exist)
- [x] T105 [P] [US1] Implement `VersionResolver` interface in `internal/doctor/doctor.go` — `ResolveLatestVersion(registry, repository string) (string, error)`. Wraps `internal/registry/resolver.go` `DefinitionVersion()`. Interface enables mock injection for tests

### 8b: Per-Provider Configuration Summary (replaces failures-only output)

- [x] T106 [US1] Refactor `CheckVariables()` in `internal/doctor/doctor.go` — add `verbose bool` parameter. Default mode: per-provider summary line with resolved count + missing count (e.g., `✅ provider/openscap: 3/3 global vars, 2/2 target vars` or `❌ provider/kube-eval: 1/2 global vars — missing workspace`). Verbose mode: append per-key status lines below each provider summary (e.g., `   global: workspace ✅, output_dir ✅`). Keep existing global + target variable validation logic (R51). Returns `[]CheckResult` — one per provider in default mode, additional detail results in verbose mode
- [x] T107 [US1] Update `Run()` signature in `internal/doctor/doctor.go` — add `verbose bool` parameter. Pass through to `CheckVariables()`. All other checks unaffected by verbose flag

### 8c: `--verbose` CLI Flag

- [x] T108 [US1] Add `--verbose` bool flag to `doctorCmd` in `cmd/complyctl/cli/doctor.go` — cobra `BoolVar`. Pass to `doctor.Run()`. Short flag: `-v`
- [x] T109 [US1] Update `runDoctor()` in `cmd/complyctl/cli/doctor.go` — pass verbose flag to `doctor.Run()`. Pass version resolver (registry client wrapper). Load cache state for version comparison

### 8d: Registry Resolver Adapter

- [x] T110 [P] [US1] Implement version resolver adapter in `cmd/complyctl/cli/doctor.go` or `internal/doctor/` — wraps `internal/registry/client.go` to satisfy `VersionResolver` interface. Creates per-registry clients dynamically (same pattern as `get` command). Handles auth via `NewCredentialFunc()`

### 8e: Tests

- [x] T111 [P] [US1] Add unit tests for `CheckPolicyVersions()` in `internal/doctor/doctor_test.go` — test cases: (1) all policies at latest → all pass, (2) one stale policy → one warning with version info, (3) unreachable registry → per-registry warning + no staleness lines for its policies, (4) mixed registries — one reachable + one unreachable, (5) empty policies list → no checks
- [x] T112 [P] [US1] Add unit tests for refactored `CheckVariables()` in `internal/doctor/doctor_test.go` — test cases: (1) default mode → per-provider summary counts, (2) verbose mode → per-key status lines, (3) all vars present → pass with counts, (4) missing global var → fail with count + missing name, (5) missing target var → fail with count + missing name + target ID
- [x] T113 [P] [US1] Add integration test for `complyctl doctor --verbose` in `tests/e2e/` or `tests/integration_test.sh` — verify verbose output contains per-key lines, default output shows counts only

### 8f: Verification

- [x] T114 Run `go build ./...` — all Phase 8 code compiles
- [x] T115 Run `go test ./internal/doctor/...` — all doctor tests pass
- [x] T116 Run `go test ./...` — full test suite passes
- [x] T117 Run `go vet ./...` and `gofmt -l .` — clean
- [x] T118 Verify `CheckRegistries()` is no longer called — search `internal/doctor/` for `CheckRegistries`; zero matches expected (function removed or renamed)

**Checkpoint**: Doctor shows per-policy version staleness, per-provider config summary with counts, `--verbose` for key detail. Unreachable registries handled gracefully. R55 complete.

---

## Phase 9: 002-comply-packs — Pack CLI + Doctor Integration (DEFERRED)

**Purpose**: Implement pack CLI subcommands, doctor dual-file validation, pack build/push/pull, and provider directory override. Deferred to `002-comply-packs` branch. Pack manifest types available from Phase 6b.

**Dependencies**: All 001 phases complete (Phases 1-8).

### 9a: Doctor Dual-File Mode

- [ ] T119 Update `internal/doctor/doctor.go` `Run()` — detect `complypack.yaml` presence using `PackManifestExists()`; if found, load and validate pack manifest (schema, provider binaries, cache digests); if `complytime.yaml` absent but `complypack.yaml` present, report pack OK + config missing with remediation guidance (R53)
- [ ] T120 Add `CheckPackManifest()` to `internal/doctor/doctor.go` — validate manifest schema, verify provider binaries exist in `./bin/`, verify cached policy OCI layouts exist in `./policies/`
- [ ] T121 Add `CheckSystemDeps()` to `internal/doctor/doctor.go` — run each `system-dependencies[].check` command, report pass/fail per dependency
- [ ] T122 Update `cmd/complyctl/cli/doctor.go` — pass `complypack.yaml` path to `doctor.Run()` when present

### 9b: Pack Build

- [ ] T123 Implement `internal/pack/build.go` — `Build()` function: fetch policies via `get` logic, copy provider binaries to `bin/`, generate `complytime.yaml.example`, assemble tarball
- [ ] T124 Add `GenerateExampleConfig()` — create `complytime.yaml.example` from pack manifest policies + empty targets

### 9c: Pack CLI Subcommands

- [ ] T125 Create `cmd/complyctl/cli/pack.go` — `packCmd` command group (`Use: "pack"`, `Short: "Manage comply-packs"`)
- [ ] T126 Add `pack init` subcommand — prompts for pack id, version, registry URL, policy IDs, provider IDs. Creates `complypack.yaml`
- [ ] T127 Add `pack doctor` subcommand — validates manifest for buildability (registry reachable, policies exist, profiles valid)
- [ ] T128 Add `pack build` subcommand — invokes `internal/pack/build.go`
- [ ] T129 Add `pack push` subcommand — pushes tarball as OCI artifact to registry
- [ ] T130 Add `pack pull` subcommand — retrieves pack tarball from registry
- [ ] T131 Register `packCmd` in `cmd/complyctl/cli/root.go`

### 9d: Config Validation — Pack Context

- [ ] T132 Update doctor `CheckConfig` — if in pack context, validate that target policy references exist in pack's policy list
- [ ] T133 Update `Validate()` — optional `registry.url` when in pack context (policies pre-cached)

### 9e: Provider Directory Override

- [ ] T134 Add `COMPLYTIME_PROVIDER_DIR` env var support to `pkg/plugin/discovery.go` — overrides default `~/.complytime/providers/`

### 9f: Verification

- [ ] T135 Run `go build ./...` — all Phase 9 code compiles
- [ ] T136 Run `go test ./...` — all tests pass
- [ ] T137 Run `go vet ./...` and `gofmt -l .` — clean
- [ ] T138 E2E test: Fedora comply-pack workflow (build → push → pull → doctor → scan)
- [ ] T139 Verify no field overlap between `PackManifest` and `WorkspaceConfig`

**Checkpoint**: Pack CLI complete. Doctor validates both files. Build/push/pull works. Provider dir override works.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies. Foundation for everything
- **Phase 2 (Foundational)**: Depends on Phase 1. All shared infrastructure
- **Phase 3 (US1)**: Depends on Phase 2. Init/get/doctor/list/providers CLI commands
- **Phase 4 (US2)**: Depends on Phase 2. Can run in parallel with Phase 3 (different CLI files). Generate/scan depend on cache infrastructure from Phase 2
- **Phase 5 (US3)**: Depends on Phase 4 (output formatters wire into scan). Can run 5a in parallel with Phase 4 (different files)
- **Phase 6 (Polish)**: Depends on Phases 3-5. Cross-cutting cleanup
- **Phase 7 (Docs)**: Depends on Phase 6. Documentation alignment only
- **Phase 8 (Doctor Redesign)**: Depends on Phases 1-7. Modifies `internal/doctor/doctor.go` and `cmd/complyctl/cli/doctor.go`. Uses existing `internal/registry/resolver.go` and `internal/cache/state.go`
- **Phase 9 (002-comply-packs)**: Depends on all 001 phases (1-8). Separate branch

### Parallel Opportunities

**Within Phase 2**: T016+T017+T020 parallel (different files). T021+T023 parallel. T028+T029+T034 parallel. T035+T038+T039 parallel.

**Phase 3 vs Phase 4**: T043 (init.go) parallel with T057 (generate.go) — different CLI files. T047 (doctor.go) parallel with T059 (scan.go). T050 (list.go) parallel with T061 (execution_plan.go).

**Within Phase 5**: T071+T072+T073 fully parallel (oscal.go, markdown.go, sarif.go — independent formatters).

**Within Phase 6**: T079-T082 fully parallel (different behavioral test files). T084+T085+T086 parallel with T087-T090 (pack model vs test configs — different files).

**Within Phase 8**: T105 (VersionResolver interface) parallel with T106 (CheckVariables refactor) — different functions. T110 (resolver adapter) parallel with T111+T112 (unit tests) after T103-T107 complete. T111+T112+T113 fully parallel (different test files/functions).

---

## Implementation Strategy

### MVP (Phases 1-3) — COMPLETE

Setup + foundational infrastructure + US1 (init, get, doctor, list, providers). Delivers a working policy cache that can be fetched and validated.

### Core Value (Phase 4) — COMPLETE

US2 (generate, scan). Delivers the primary compliance scanning capability.

### Full Feature (Phase 5) — COMPLETE

US3 (output formatters, scan summary). Delivers multi-format output and terminal UX.

### Polish (Phase 6) — COMPLETE

Behavioral tests, pack data model, PolicyEntry refactoring.

### Documentation (Phase 7) — COMPLETE

quickstart.md aligned with PolicyEntry model.

### Doctor Redesign (Phase 8) — COMPLETE

Per-policy version comparison, per-provider config summary, `--verbose` flag. R55.

### Comply-Packs (Phase 9) — DEFERRED to 002

Full pack CLI lifecycle. Separate feature branch.

---

## Key Implementation Notes

- **PolicyEntry model (Session 2026-02-25d)**: `Policies []PolicyEntry` where each entry has `url` (full OCI reference) and optional `id` (shortname). `EffectiveID()` returns explicit `id` or derives from last URL path segment. No `pack` field, no `RegistryConfig` in `WorkspaceConfig`. Targets use `policies` (list of effective IDs). `FindPolicy()` matches by effective ID → full URL → repo path. `UniqueRegistries()` extracts distinct registries for per-registry client creation
- **Three-tier variables (R48)**: Global variables flow via `GenerateRequest.global_variables`. Test variables flow via `AssessmentConfiguration.parameters`. Target variables flow via `Target.variables`. Do not conflate
- **Targets-only Scan RPC (R47)**: No `requirement_ids`. Providers evaluate all requirements from Generate-time state
- **Multi-evaluator (R32)**: `GroupByEvaluator()` groups by per-plan executor ID. Execution plan shows unmatched evaluators as ERROR rows
- **Zero custom auth (R6, R24)**: All credential resolution via `oras-credentials-go`. No `Authenticator` struct
- **Result aggregation (R45)**: Delegates to go-gemara. No custom aggregation in complyctl
- **HealthCheck variable declaration (R51)**: Providers declare required variable *names* via `HealthCheckResponse`. Doctor validates presence in config. Valid *values* remain provider's responsibility
- **Pack separation (Session 2026-02-25d)**: Pack builder is a separate tool. All pack CLI commands deferred to 002. Pack manifest types remain in `internal/complytime/pack.go` as data model
- **Charmbracelet consistency (R38)**: All tabular outputs use `bubbles/table` + `lipgloss` via `internal/terminal` helpers. `--plain` flag only on `list` and `providers`
- **parsePolicyLayer error contract (R39)**: Only `gemara.Policy` with `adherence.assessment-plans` accepted. No fallback
- **Terminal output tiers (R40)**: Tier 1 (`init`, `get`) writes progress to stderr. Tier 2 (`list`, `providers`, `generate`, `scan`, `doctor`) writes summaries to stdout. No log file for `version`, `list`, `init`, `get`
- **Convention over configuration (R50)**: `complytime.yaml` is a static convention. No `--config` flag. `--policy-id` always required
- **Doctor version comparison (R55)**: `CheckPolicyVersions()` replaces `CheckRegistries()`. Compares cached vs remote latest per-policy. Non-blocking warnings for stale policies and unreachable registries. `--verbose` expands per-provider variable detail only. Uses existing `DefinitionVersion()` and `PolicyState` — no new infrastructure

---

## Notes

- [P] tasks = different files, no dependencies — safe to parallelize
- [USn] label maps task to specific user story
- Each phase is independently testable at its checkpoint
- All 001 tasks (T001-T118) are COMPLETE
- Phase 8 (T103-T118) is the doctor redesign — R55, Session 2026-02-25e
- Phase 9 (T119-T139) is deferred to `002-comply-packs` branch
