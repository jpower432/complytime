---
name: /ampel-create-policy
id: ampel-create-policy
category: Policy
description: Create an AMPEL policy from a Gemara requirement
---

Create an AMPEL policy from a Gemara requirement.

I'll help you convert a Gemara requirement to an AMPEL policy file with the correct structure, including policy metadata, control references, and verification tenets.

---

**Input**: The user should provide:
- Gemara requirement ID (e.g., 'ac-3')
- Compliance framework and control references (e.g., NIST-800-53 AC-3)
- Description of what the policy should verify

**Steps**

1. **Gather requirement information**
   - Extract or ask for the Gemara requirement ID
   - Identify compliance control references (framework, class, id)
   - Understand what the policy should verify

2. **Create AMPEL policy structure**
   - Policy id: Convert requirement ID to kebab-case (e.g., 'ac-3' → 'ac-3-access-control-enforcement')
   - meta.description: Clear description of what the policy verifies
   - meta.controls: Array of control references with framework, class, and id

3. **Define verification tenets**
   - Create tenets array with verification checks
   - Each tenet should have:
     - id: Unique tenet identifier
     - code: CEL expression that evaluates the attestation data
     - predicates.types: Attestation predicate type URIs
     - assessment.message: Success message
     - error.message and error.guidance: Failure remediation

4. **Validate policy structure**
   - Ensure all required fields are present: id, meta, tenets
   - Verify meta has description and controls
   - Check each tenet has id, code, and predicates fields

5. **Write policy file**
   - Save as JSON file with kebab-case name matching policy id
   - Place in appropriate AMPEL policy directory

**Output**

After creating the policy:
- Policy file location
- Summary of policy structure
- List of tenets created
- Next steps (e.g., test with ampel verify)

**AMPEL Policy Structure**

AMPEL policies are JSON files with the following structure:

```json
{
  "id": "BP-1.01",
  "meta": {
    "description": "Validate branch protection settings require pull requests",
    "controls": [
      { "framework": "repo-branch-protection", "class": "source-code", "id": "BP-1" }
    ]
  },
  "tenets": [
    {
      "id": "01",
      "code": "has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == \"update\") : false",
      "predicates": {
        "types": ["http://github.com/carabiner-dev/snappy/specs/branch-rules.yaml"]
      },
      "assessment": {
        "message": "Direct pushes are disabled in default branch. PR required."
      },
      "error": {
        "message": "Direct pushes are enabled so PRs are not required.",
        "guidance": "Create a branch ruleset protecting your default branch and enable \"Restrict updates\""
      }
    }
  ]
}
```

**Input Data Structure (Snappy Attestations)**

AMPEL policies evaluate in-toto attestations produced by snappy. The attestation structure is:

```json
{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [...],
  "predicateType": "http://github.com/carabiner-dev/snappy/specs/branch-rules.yaml",
  "predicate": {
    "data": {
      "values": [
        {
          "type": "update",
          "parameters": {}
        },
        {
          "type": "pull_request",
          "parameters": {
            "required_approving_review_count": 1,
            "dismiss_stale_reviews_on_push": true,
            "require_last_push_approval": true,
            "require_code_owner_review": false
          }
        },
        {
          "type": "non_fast_forward",
          "parameters": {}
        }
      ]
    }
  }
}
```

**How Predicates Work**

1. Each tenet specifies `predicates.types` array with predicate type URIs (e.g., `http://github.com/carabiner-dev/snappy/specs/branch-rules.yaml`)
2. AMPEL matches attestations by predicate type URI
3. Matched predicates are available in the `predicates` array (indexed by match order)
4. Access attestation data via `predicates[0].data.values` (array of rule objects)

**CEL Expression Patterns**

The `code` field uses CEL (Common Expression Language) to evaluate attestation data:

- **Check if data exists**: `has(predicates[0].data.values)`
- **Find rule by type**: `predicates[0].data.values.exists(rule, rule.type == "update")`
- **Check rule parameters**: `predicates[0].data.values.exists(rule, rule.type == "pull_request" && rule.parameters.required_approving_review_count >= 1)`
- **Multiple conditions**: `predicates[0].data.values.exists(rule, rule.type == "pull_request" && rule.parameters.dismiss_stale_reviews_on_push == true)`

**Common Rule Types (GitHub Branch Protection)**

- `"update"`: Restricts direct pushes (requires PR)
- `"pull_request"`: Pull request requirements
  - `parameters.required_approving_review_count`: Minimum approvals (integer)
  - `parameters.dismiss_stale_reviews_on_push`: Dismiss stale reviews (boolean)
  - `parameters.require_last_push_approval`: Require last push approval (boolean)
  - `parameters.require_code_owner_review`: Require code owner review (boolean)
- `"non_fast_forward"`: Blocks force pushes

**Example CEL Expressions**

```cel
// Check if direct pushes are disabled
has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == "update") : false

// Check minimum approval count
has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == "pull_request" && rule.parameters.required_approving_review_count >= 1) : false

// Check multiple PR requirements
has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == "pull_request" && rule.parameters.dismiss_stale_reviews_on_push == true && rule.parameters.require_last_push_approval == true) : false

// Check force push blocking
has(predicates[0].data.values) ? predicates[0].data.values.exists(rule, rule.type == "non_fast_forward") : false
```

**Gemara to AMPEL Conversion Guidelines**

- Map Gemara requirement ID to AMPEL policy id (kebab-case)
- Convert Gemara control references to meta.controls array format
- Transform Gemara assessment requirements into AMPEL tenets
- Each tenet represents a specific verification check
- Use CEL expressions to evaluate snappy attestation data structure
