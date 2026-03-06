# AMPEL Policy Creation Skill

This skill enables creating AMPEL (Attestation Metadata Policy Expression Language) policies from Gemara requirements and compliance controls.

## Overview

AMPEL policies define verification rules for in-toto attestations. They evaluate attestation data using CEL (Common Expression Language) expressions to determine compliance with security controls.

## Core Concepts

### Policy Structure

AMPEL policies are JSON files with three main sections:

1. **Policy ID**: Unique identifier (kebab-case)
2. **Metadata**: Description and control references
3. **Tenets**: Array of verification checks

### Tenet Structure

Each tenet represents a specific verification check:

- **id**: Unique tenet identifier (typically sequential: "01", "02", etc.)
- **code**: CEL expression that evaluates attestation data
- **predicates.types**: Array of predicate type URIs that this tenet matches
- **assessment.message**: Success message when tenet passes
- **error.message**: Failure message when tenet fails
- **error.guidance**: Remediation guidance for failures

### Predicate Matching

AMPEL matches attestations by predicate type URI:

1. Each tenet specifies `predicates.types` array
2. AMPEL finds attestations matching those URIs
3. Matched predicates are available in `predicates` array (indexed by match order)
4. Access data via `predicates[0].data.values`

## AMPEL Policy Template

```json
{
  "id": "policy-id-kebab-case",
  "meta": {
    "description": "Clear description of what this policy verifies",
    "controls": [
      {
        "framework": "framework-name",
        "class": "control-class",
        "id": "control-id"
      }
    ]
  },
  "tenets": [
    {
      "id": "01",
      "code": "has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == \"rule-type\") : false",
      "predicates": {
        "types": ["http://github.com/carabiner-dev/snappy/specs/spec-name.yaml"]
      },
      "assessment": {
        "message": "Success message describing what passed"
      },
      "error": {
        "message": "Failure message describing what failed",
        "guidance": "Actionable remediation guidance"
      }
    }
  ]
}
```

## CEL Expression Patterns

### Basic Checks

```cel
// Check if data exists
has(predicates[0].data.values)

// Find rule by type
predicates[0].data.values.exists(rule, rule.type == "update")

// Check rule parameters
predicates[0].data.values.exists(rule, rule.type == "pull_request" && rule.parameters.required_approving_review_count >= 1)

// Multiple conditions
predicates[0].data.values.exists(rule, rule.type == "pull_request" && rule.parameters.dismiss_stale_reviews_on_push == true)
```

### Safe Evaluation Pattern

Always use safe evaluation to handle missing data:

```cel
has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == "update") : false
```

## Common Snappy Specs

### GitHub Branch Protection

**Predicate Type**: `http://github.com/carabiner-dev/snappy/specs/branch-rules.yaml`

**Rule Types**:
- `"update"`: Restricts direct pushes (requires PR)
- `"pull_request"`: Pull request requirements
  - `parameters.required_approving_review_count`: Minimum approvals (integer)
  - `parameters.dismiss_stale_reviews_on_push`: Dismiss stale reviews (boolean)
  - `parameters.require_last_push_approval`: Require last push approval (boolean)
  - `parameters.require_code_owner_review`: Require code owner review (boolean)
- `"non_fast_forward"`: Blocks force pushes

**Example Tenet**:

```json
{
  "id": "01",
  "code": "has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == \"update\") : false",
  "predicates": {
    "types": ["http://github.com/carabiner-dev/snappy/specs/branch-rules.yaml"]
  },
  "assessment": {
    "message": "Direct pushes are disabled. Pull requests required."
  },
  "error": {
    "message": "Direct pushes are enabled. Pull requests are not required.",
    "guidance": "Create a branch protection rule and enable 'Restrict updates' to require pull requests"
  }
}
```

## Conversion Workflow

### From Gemara Requirement

1. **Extract Requirement ID**: Convert to kebab-case for policy id
2. **Map Controls**: Convert Gemara control references to `meta.controls` format
3. **Identify Verification Points**: Break down requirement into specific checks
4. **Create Tenets**: One tenet per verification check
5. **Write CEL Expressions**: Evaluate attestation data structure
6. **Add Messages**: Provide clear success/failure messages and guidance

### Best Practices

- **One Check Per Tenet**: Each tenet should verify a single, specific condition
- **Clear Messages**: Assessment and error messages should be actionable
- **Safe Evaluation**: Always check for data existence before accessing
- **Descriptive IDs**: Use sequential IDs ("01", "02") that map to requirement sections
- **Control Mapping**: Include all relevant compliance framework references

## File Naming Convention

- Policy files: `{policy-id}.json` (matches policy id field)
- Location: `{workspace}/ampel/granular-policies/` (default)
- Example: `ac-3-access-control-enforcement.json`

## Validation Checklist

Before saving a policy, verify:

- [ ] Policy id is kebab-case and matches filename
- [ ] meta.description clearly describes verification purpose
- [ ] meta.controls includes all relevant framework references
- [ ] Each tenet has unique id
- [ ] Each tenet has valid CEL expression in code field
- [ ] Each tenet specifies predicates.types array
- [ ] Assessment and error messages are clear and actionable
- [ ] CEL expressions use safe evaluation pattern
