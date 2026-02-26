# E2E Testing

## Automated

```bash
make test-e2e
```

Builds `complyctl` + `complyctl-provider-test`, then runs all e2e tests with an in-process mock OCI registry. No external services required.

Tests live in `tests/e2e/`. Build tag: `e2e`.

| Test | Validates |
|:---|:---|
| `FullWorkflow` | get → list → generate → scan (oscal, pretty, sarif) |
| `PolicyCache` | OCI layout structure, state.json tracking |
| `MultiplePolicies` | Multi-policy fetch + list |
| `ScanDefaultFormat` | No --format = EvaluationLog only |
| `InvalidFormat` | `--format pdf` rejected |
| `MissingPolicy` | Uncached policy fails with clear message |
| `MockRegistryOCICompliance` | v2 endpoint, catalog, tags, manifests, 404s |
| `MockPluginHealthCheck` | Plugin discovery + HealthCheck + Generate RPC |
| `Help` | CLI help output structure |
| `Version` | Version command output |
| `ListFilterByPolicyID` | `--policy-id` filter on list |

## Manual Walkthrough

Step-by-step validation using the mock OCI registry and test plugin.

### Prerequisites

```bash
make build build-test-plugin
```

### Step 1: Start mock OCI registry

```bash
make mock-registry
```

Listens on `http://localhost:8765`. Pre-seeded policies:

| Policy ID | Layers | Tags |
|:---|:---|:---|
| `policies/nist-800-53-r5` | catalog + policy | v1.0.0, latest |
| `policies/cis-benchmark` | catalog | v2.0.0, latest |
| `catalogs/osps-b` | catalog | v1.0.0, latest |
| `guidance/nist` | guidance | v1.0.0, latest |

Verify it responds:

```bash
curl -s http://localhost:8765/v2/ | jq .
curl -s http://localhost:8765/v2/_catalog | jq .
```

### Step 2: Install test plugin

```bash
mkdir -p ~/.complytime/providers
cp bin/complyctl-provider-test ~/.complytime/providers/
```

The test plugin responds to all RPCs (HealthCheck, Generate, Scan) with predefined pass results. Evaluator ID: `test`.

### Step 3: Create workspace config

This walkthrough uses standalone mode (`complytime.yaml` only). For comply-pack workflows where a `complypack.yaml` coexists with the runtime config, see [COMPLY_PACK_QUICKSTART.md](COMPLY_PACK_QUICKSTART.md).

```bash
cat > complytime.yaml << 'EOF'
registry:
  url: http://localhost:8765
policies:
  - id: policies/nist-800-53-r5
targets:
  - id: local
    policy_ids:
      - policies/nist-800-53-r5
    variables:
      env: manual-test
EOF
```

### Step 4: Fetch policies

```bash
bin/complyctl get
```

**Verify:**

```bash
# Cache directory created
ls ~/.complytime/policies/policies/nist-800-53-r5/

# State file tracks policy
cat ~/.complytime/state.json | jq .
```

Expected: `oci-layout` file exists, state.json contains policy digest and version.

### Step 5: List cached policies

```bash
bin/complyctl list --plain
```

Expected: `policies/nist-800-53-r5` appears with cached version.

### Step 6: Generate policy graph

```bash
bin/complyctl generate --policy-id policies/nist-800-53-r5
```

Expected: `Gemara policy generation completed` log line. Plugin receives Generate RPC with assessment configurations extracted from the policy layer.

### Step 7: Scan — EvaluationLog only (default)

```bash
bin/complyctl scan --policy-id policies/nist-800-53-r5
```

**Verify:**

```bash
ls .complytime/scan/
cat .complytime/scan/evaluation-log-*.json | jq .
```

Expected: Single `evaluation-log-*.json` file. No OSCAL, SARIF, or Markdown files.

### Step 8: Scan — OSCAL format

```bash
rm -rf .complytime/scan
bin/complyctl scan --policy-id policies/nist-800-53-r5 --format oscal
```

**Verify:**

```bash
cat .complytime/scan/assessment-results-*.json | jq '.["assessment-results"].metadata'
```

Expected: `oscal-version: "1.2.0"`, results array with findings.

### Step 9: Scan — Markdown format

```bash
rm -rf .complytime/scan
bin/complyctl scan --policy-id policies/nist-800-53-r5 --format pretty
```

**Verify:**

```bash
cat .complytime/scan/report-*.md
```

Expected: Markdown with `# Compliance Scan Report` header, target sections, step results.

### Step 10: Scan — SARIF format

```bash
rm -rf .complytime/scan
bin/complyctl scan --policy-id policies/nist-800-53-r5 --format sarif
```

**Verify:**

```bash
cat .complytime/scan/scan-*.sarif.json | jq '.version'
```

Expected: SARIF version `"2.1.0"`.

### Step 11: Negative tests

```bash
# Invalid format
bin/complyctl scan --policy-id policies/nist-800-53-r5 --format pdf
# Expected: error containing "invalid format"

# Missing policy (without running get)
rm -rf ~/.complytime/policies
bin/complyctl scan --policy-id nonexistent
# Expected: error containing "not in cache"
```

### Cleanup

```bash
rm -rf .complytime/scan complytime.yaml
rm -rf ~/.complytime/policies ~/.complytime/state.json
rm ~/.complytime/providers/complyctl-provider-test
# Kill mock registry (Ctrl+C in its terminal)
```

## Adding New Tests

1. Open `tests/e2e/e2e_test.go`.
2. Add a `TestE2E_*` function using the shared helpers from `helpers_test.go`.
3. Use `startMockRegistry(t)` for an isolated in-process registry per test.
4. Use `installTestPlugin(t, homeDir)` to deploy the test plugin.
5. Use `runComplytime(t, binary, workDir, env, args...)` to execute commands.
6. Run: `make test-e2e`
